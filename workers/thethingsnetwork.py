import json
import base64
from dateutil.parser import isoparse


def create_gateway_measurements(msg, gw):
    tss = gw.get("time")
    ts = gw.get("timestamp") or (isoparse(tss).timestamp()) if tss is not None else msg.timestamp

    eui = gw.get("gateway_ids").get("eui")
    rssi = gw.get("rssi")
    snr = gw.get("snr")

    if rssi is not None:
        msg.create_measurement(float(rssi), f"rssi_{eui}", "dB").set_metadata(
            {"gateway_eui": eui}).set_sensor("antenna").set_timestamp(ts).add()
    if snr is not None:
        msg.create_measurement(float(snr), f"snr_{eui}", "dB").set_metadata(
            {"gateway_eui": eui}).set_sensor("antenna").set_timestamp(ts).add()


def process(payload, msg):
    payload = json.loads(payload.decode())
    uplink = payload.get("uplink_message")
    if uplink is None:
        return b''

    received_at = payload.get("received_at")
    if received_at is not None:
        received_at = isoparse(received_at)
        msg.set_time(received_at)

    f_port = uplink.get("f_port")
    if f_port is not None:
        msg.metadata["f_port"] = int(f_port)

    f_payload = uplink.get("frm_payload") or b''
    if f_payload != b'':
        f_payload = base64.b64decode(f_payload)

    gateways = uplink.get("rx_metadata")
    if gateways is not None and len(gateways) > 0:
        for gw in gateways:
            create_gateway_measurements(msg, gw)

    msg.match_device({
        "dev_eui": payload.get("end_device_ids").get("dev_eui")
    })

    return f_payload
