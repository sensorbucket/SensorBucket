import flask
import base64
from flask import request
import usercode
import json
import os
import traceback
import requests as r
from datetime import datetime
from typing import List, Dict, Optional, Any, Union
from copy import deepcopy


class ErrMessageNoSteps(Exception):
    pass


class MissingRequiredProperties(Exception):
    pass


class MissingRequiredEnvironmentVariable(Exception):
    pass


class PropertiesMatchedNotExactlyOneDevice(Exception):
    pass


class ErrUserCodeFailure(Exception):
    pass


if "ENDPOINT_DEVICES" not in os.environ:
    raise MissingRequiredEnvironmentVariable("Environment variable: ENDPOINT_DEVICES must be set")
DEVICES_EP = os.environ["ENDPOINT_DEVICES"]


class Serializer(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, bytes):
            return base64.b64encode(obj).decode('ascii')
        if callable(getattr(obj, "json_dict", None)):
            return obj.json_dict()
        return json.JSONEncoder.default(self, obj)


class Measurement:
    def __init__(self,
                 timestamp: int,
                 sensor_external_id: str,
                 value: float,
                 observed_property: str,
                 unit_of_measurement: str,
                 latitude: Optional[float] = None,
                 longitude: Optional[float] = None,
                 altitude: Optional[float] = None,
                 properties: Optional[Dict[str, Any]] = None):
        self.timestamp = int(timestamp)
        self.sensor_external_id = str(sensor_external_id)
        self.value = float(value)
        self.observed_property = str(observed_property)
        self.unit_of_measurement = str(unit_of_measurement)
        self.latitude = float(latitude) if latitude else None
        self.longitude = float(longitude) if longitude else None
        self.altitude = float(altitude) if altitude else None
        self.properties = properties if properties else {}

    def json_dict(self):
        d = self.__dict__
        return d


class Message:
    def __init__(self,
                 tracing_id: str,
                 tenant_id: int,
                 access_token: str,
                 received_at: int,
                 pipeline_id: str,
                 step_index: int,
                 pipeline_steps: List[str],
                 timestamp: int,
                 device: Optional[Any] = None,
                 measurements: Optional[List[Measurement]] = None,
                 payload: Optional[bytes] = None,
                 metadata: Optional[Dict[str, Any]] = None):
        self.tracing_id = tracing_id
        self.tenant_id = tenant_id
        self.access_token = access_token
        self.received_at = received_at
        self.pipeline_id = pipeline_id
        self.step_index = step_index
        self.pipeline_steps = pipeline_steps
        self.timestamp = timestamp
        self.device = device
        self.measurements = measurements if measurements else []
        self.metadata = metadata if metadata else {}
        self.payload = payload

    def match_device(self, properties: Dict[str, Any]):
        res = r.get(DEVICES_EP, params={
            "properties": json.dumps(properties),
        }, headers={
            "authorization": "bearer " + self.access_token
        })
        res.raise_for_status()
        data = res.json()
        devices = data["data"]
        if len(devices) == 0:
            raise PropertiesMatchedNotExactlyOneDevice(
                f"can't find device with properties: {properties}")
        if len(devices) > 1:
            raise PropertiesMatchedNotExactlyOneDevice(
                f"too many devices match properties: {properties}")
        self.device = devices[0]

    def create_measurement(self, value: float, obs: str, uom: str):
        builder = MeasurementBuilder(self)
        return builder.set_value(value, obs, uom)

    def current_step(self) -> Union[str, ErrMessageNoSteps]:
        try:
            return self.pipeline_steps[self.step_index]
        except IndexError:
            raise ErrMessageNoSteps("pipeline message has no steps remaining")

    def next_step(self) -> Union[str, ErrMessageNoSteps]:
        self.step_index += 1
        try:
            return self.pipeline_steps[self.step_index]
        except IndexError:
            raise ErrMessageNoSteps("pipeline message has no steps remaining")

    @classmethod
    def from_json(cls, json_str: str):
        data = json.loads(json_str)
        required_fields = ['tracing_id', 'tenant_id', 'access_token', 'received_at',
                           'pipeline_id', 'step_index', 'pipeline_steps', 'timestamp']
        if not all(field in data for field in required_fields):
            missing = [field for field in required_fields if field not in data]
            raise MissingRequiredProperties(f"Missing required properties in JSON: {','.join(missing)}")

        payload = data.get('payload') or None
        if payload is None:
            payload = b''
        elif isinstance(payload, str):
            payload = base64.b64decode(payload)
        elif not isinstance(payload, bytes):
            raise TypeError("Payload must be a base64 encoded string or a bytestring")
        data["payload"] = payload

        return cls(**data)

    def set_time(self, dt: datetime):
        self.timestamp = int(dt.timestamp())*1000
        return self

    def json_dict(self):
        d = self.__dict__
        return d


class MeasurementBuilder:
    def __init__(self, message: Message):
        self.message = message
        self.measurement = Measurement(
            timestamp=message.timestamp,
            sensor_external_id='',
            value=0.0,
            observed_property='',
            unit_of_measurement='',
            properties={}
        )

    def set_timestamp(self, ts: int):
        self.measurement.timestamp = ts
        return self

    def set_time(self, dt: datetime):
        self.measurement.timestamp = int(dt.timestamp())*1000
        return self

    def set_sensor(self, eid: str):
        self.measurement.sensor_external_id = eid
        return self

    def set_value(self, value: float, obs: str, uom: str):
        self.measurement.value = value
        self.measurement.observed_property = obs
        self.measurement.unit_of_measurement = uom
        return self

    def set_metadata(self, meta: Dict[str, Any]):
        self.measurement.properties = meta
        return self

    def set_location(self, latitude: float, longitude: float, altitude: float):
        self.measurement.latitude = latitude
        self.measurement.longitude = longitude
        self.measurement.altitude = altitude
        return self

    def add(self):
        self.message.measurements.append(self.measurement)


class PipelineError:
    def __init__(self,
                 received_by_worker: Message,
                 processing_attempt: Message,
                 worker: str,
                 queue: str,
                 timestamp: int,
                 error: str):
        self.received_by_worker = received_by_worker
        self.processing_attempt = processing_attempt
        self.worker = worker
        self.queue = queue
        self.timestamp = timestamp
        self.error = error

    def json_dict(self):
        d = self.__dict__
        return d


def main():
    topic = request.headers.get("X-Amqp-Topic", "No Topic")
    message: Message = None
    original_message: Message = None
    try:
        message = Message.from_json(request.get_data(as_text=True))
        original_message = deepcopy(message)
        message = usercode.process(message)

        if not isinstance(message, Message):
            raise ErrUserCodeFailure("usercode return value must be Message object")

        next_step = message.next_step()
        res = flask.Response(json.dumps(message, cls=Serializer), content_type="application/json")
        res.headers["X-Amqp-Topic"] = next_step
        return res
    except BaseException as e:
        tracing_id = "00000000-0000-0000-0000-000000000000"
        if original_message is not None:
            tracing_id = original_message.tracing_id
        print(f"exception occured for id: {tracing_id}:\n {e}")
        pipeline_error = PipelineError(
            received_by_worker=original_message,
            processing_attempt=message,
            worker=topic,
            queue=topic,
            timestamp=int(datetime.now().timestamp()),
            error=traceback.format_exc()
        )
        res = flask.Response(json.dumps(pipeline_error, cls=Serializer), content_type="application/json")
        res.headers["X-AMQP-Topic"] = "errors"
        return res
