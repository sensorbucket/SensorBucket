class InvalidPayload(BaseException):
    pass


def process(data, msg):
    if len(data) != 2:
        raise InvalidPayload("Data expected to be two bytes")
    value = (int(data[0]) << 8) | int(data[1])
    msg.create_measurement(value, "pm_2.5", "ug/m3").set_sensor("0").add()
    return b''
