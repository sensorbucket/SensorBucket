from base64 import b64decode
from binascii import hexlify, unhexlify


def bcd(data):
    return data >> 4 & 0xf, data & 0xf


def decode_temperature(msg, data):
    t = data[0] & 0x7f
    if t & 0x80 > 0:
        t -= 128
    msg.create_measurement(float(t), "temperature", "Cel").set_sensor("adeunis").add()
    return data[1:]


def trigger_accelerometer(msg, data):
    print("Trig AX")
    return data


def trigger_button(msg, data):
    print("Trig BTN")
    return data


def decode_coordinates(msg, data):
    l0, l1 = bcd(data[0])
    l2, l3 = bcd(data[1])
    l4, l5 = bcd(data[2])
    l6, _ = bcd(data[3])
    NS = (data[3] & 0b1) * -2 + 1
    degree = l0 * 10 + l1
    minutes = l2 * 10 + l3 + l4/10 + l5/100 + l6/1000
    lat = degree + minutes/60 * NS

    l0, l1 = bcd(data[4])
    l2, l3 = bcd(data[5])
    l4, l5 = bcd(data[6])
    l6, _ = bcd(data[7])
    EW = (data[7] & 0b1) * -2 + 1
    degree = l0 * 100 + l1 * 10 + l2
    minutes = l3 * 10 + l4 + l5/10 + l6/100
    lng = degree + minutes/60 * EW

    tmp = data[8]
    hdop = (tmp >> 4) & 0x0f
    sats = (tmp & 0x0f)

    msg.create_measurement(float(hdop), "hdop", "#").set_sensor("adeunis").add()
    msg.create_measurement(float(sats), "sats", "#").set_sensor("adeunis").add()

    return data[9:]


def decode_ul(msg, data):
    msg.create_measurement(float(data[0]), "uplinks", "#").set_sensor("adeunis").add()
    return data[1:]


def decode_dl(msg, data):
    msg.create_measurement(float(data[0]), "downlinks", "#").set_sensor("adeunis").add()
    return data[1:]


def decode_battery(msg, data):
    battery = data[0] << 8 | data[1]
    msg.create_measurement(float(battery), "battery", "mV").set_sensor("adeunis").add()
    return data[2:]


def decode_signal(msg, data):
    rssi = data[0] * -1
    snr = data[1] & 0x7f
    if snr & 0x80 > 0:
        snr -= 128

    msg.create_measurement(float(rssi), "dl_rssi", "dB").set_sensor("adeunis").add()
    msg.create_measurement(float(snr), "dl_snr", "dB").set_sensor("adeunis").add()

    return data[2:]


mapping = [
    [decode_temperature, "temperature"],  # 0
    [trigger_accelerometer, "accelerometer"],  # 1
    [trigger_button, "button"],  # 2
    [decode_coordinates, "coordinates"],  # 3
    [decode_ul, "ul"],  # 4
    [decode_dl, "dl"],  # 5
    [decode_battery, "battery"],  # 6
    [decode_signal, "signal"],  # 7
]


def process(msg):
    status = msg.payload[0]
    data = msg.payload[1:]

    field_ix = 0
    while status > 0:
        if (status & 0x80) > 0:
            decode, name = mapping[field_ix]
            data = decode(msg, data)
        status = (status << 1) & 0xFE
        field_ix += 1

    return msg
