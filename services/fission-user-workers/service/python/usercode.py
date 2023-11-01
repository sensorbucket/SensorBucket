def process(msg):
    msg.create_measurement(1.234, "test", "C").add()
    msg.payload = f"test: {msg.payload}"
    return msg
