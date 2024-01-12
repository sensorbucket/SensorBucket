import unittest
import json
from unittest.mock import patch
from base import Measurement, Message, ErrMessageNoSteps, main as handle_request
from flask import Flask

msg_stub = {
    "tracing_id": "123123123123123",
    "owner_id": 5,
    "received_at": 1702980356,
    "pipeline_id": "9812673498172398712938",
    "step_index": 0,
    "pipeline_steps": [
        "a",
        "b",
        "c"
    ],
    "timestamp": 1702980356,
    "device": None,
    "measurements": [],
    "metadata": {},
    "payload": "Cj5AuEgBFYAXOQE6YrxaNdyrguucTkoSnEdwVywBWIateVYg8SeCW0gsC16HMgEEVcbllZhFP90OK6wfhcI5xiISCODKu50DEgoaCAjI0AcQCigBKkUKEGZjYzIzZGZmZmUyZDcxN2QQif78mgkaBgjtw4WsBjCL//////////8BPQAAgL9qEAAAAAAAAAAAAEwBAT2d26yAAQI="
}


class TestMeasurement(unittest.TestCase):

    def setUp(self):
        app = Flask(__name__)
        app.add_url_rule("/", view_func=handle_request, methods=["POST"])
        self.app = app.test_client()
        self.app.testing = True

    def test_initialization(self):
        # Create a Measurement instance
        measurement = Measurement(
            timestamp=123456789,
            sensor_external_id="sensor_1",
            value=25.5,
            observed_property="temperature",
            unit_of_measurement="Celsius"
        )

        # Check if the values are correctly assigned
        self.assertEqual(measurement.timestamp, 123456789)
        self.assertEqual(measurement.sensor_external_id, "sensor_1")
        self.assertEqual(measurement.value, 25.5)
        self.assertEqual(measurement.observed_property, "temperature")
        self.assertEqual(measurement.unit_of_measurement, "Celsius")

    def test_json_dict(self):
        # Create a Measurement instance
        measurement = Measurement(
            timestamp=123456789,
            sensor_external_id="sensor_1",
            value=25.5,
            observed_property="temperature",
            unit_of_measurement="Celsius"
        )

        # Convert to JSON dictionary
        json_dict = measurement.json_dict()

        # Check if the dictionary has the correct keys and values
        self.assertIsInstance(json_dict, dict)
        self.assertEqual(json_dict['timestamp'], 123456789)
        self.assertEqual(json_dict['sensor_external_id'], "sensor_1")
        self.assertEqual(json_dict['value'], 25.5)
        self.assertEqual(json_dict['observed_property'], "temperature")
        self.assertEqual(json_dict['unit_of_measurement'], "Celsius")

    def test_valid_request(self):
        with self.app as client:
            with patch('usercode.process') as mock_process:
                incoming = json.dumps(msg_stub)
                expected = Message(**msg_stub)
                expected.payload = "changed"

                mock_process.return_value = expected
                response = client.post('/', data=incoming, content_type='application/json')

                self.assertEqual(response.status_code, 200)
                responseJSON = json.loads(response.text)
                self.assertEqual(responseJSON["payload"], expected.payload)

    def test_fails_no_next_step(self):
        with self.app as client:
            with patch('usercode.process') as mock_process:
                incoming = json.dumps(msg_stub)
                processed_msg = Message(**msg_stub)
                processed_msg.pipeline_steps = ["first"]
                mock_process.return_value = processed_msg
                topic = "amqp_test_topic"

                response = client.post('/', data=incoming, content_type='application/json', headers={
                    "X-AMQP-Topic": topic
                })

                # Still returns a 200 because the serveice did not error, only the domain
                # Otherwise Fission will not accept the response and queue it
                self.assertEqual(response.status_code, 200)
                responseJSON = json.loads(response.text)
                self.assertEqual(responseJSON["worker"], topic)
                self.assertIn("ErrMessageNoSteps", responseJSON["error"])

    def test_fails_invalid_process_response(self):
        with self.app as client:
            with patch('usercode.process') as mock_process:
                incoming = json.dumps(msg_stub)
                mock_process.return_value = "invalid response"
                topic = "amqp_test_topic"

                response = client.post('/', data=incoming, content_type='application/json', headers={
                    "X-AMQP-Topic": topic
                })

                # Still returns a 200 because the serveice did not error, only the domain
                # Otherwise Fission will not accept the response and queue it
                self.assertEqual(response.status_code, 200)
                responseJSON = json.loads(response.text)
                self.assertEqual(responseJSON["worker"], topic)
                self.assertEqual(responseJSON["queue"], topic)
                self.assertIn("ErrUserCodeFailure", responseJSON["error"])


if __name__ == '__main__':
    unittest.main()
