def process(payload, msg):
    msg.create_measurement(1.234, "test", "C").add()
    payload = f"test: {payload}"
    return payload
