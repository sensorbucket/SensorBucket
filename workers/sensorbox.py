uom = {
    "no2": "ppb", "no2_op1": "mV", "no2_op2": "mV", "ox": "ppb",
    "ox_op1": "mV", "ox_op2": "mV", "humidity": "%", "pressure": "hPa",
    "temperature": "Cel", "pm_mc_1": "ug/m3", "pm_mc_2_5": "ug/m3",
    "pm_mc_4": "ug/m3", "pm_mc_10": "ug/m3", "pm_nc_0_5": "1/cm3",
    "pm_nc_1": "1/cm3", "pm_nc_2_5": "1/cm3", "pm_nc_4": "1/cm3",
    "pm_nc_10": "1/cm3", "pm_typical_size": "nm"
}

sensor = {
    "no2": "no2b43f", "no2_op1": "no2b43f", "no2_op2": "no2b43f",
    "ox": "oxb431", "ox_op1": "oxb431", "ox_op2": "oxb431",
    "humidity": "prht", "pressure": "prht", "temperature": "prht",
    "pm_mc_1": "sps30", "pm_mc_2_5": "sps30", "pm_mc_4": "sps30",
    "pm_mc_10": "sps30", "pm_nc_0_5": "sps30", "pm_nc_1": "sps30",
    "pm_nc_2_5": "sps30", "pm_nc_4": "sps30", "pm_nc_10": "sps30",
    "pm_typical_size": "sps30"
}


def to_short(inp, s):
    tmp = (int(inp[s*2+1]) << 8) | int(inp[s*2])
    if tmp & 0x8000:
        tmp ^= 0xffff
        tmp += 1
        tmp = -tmp
    return tmp


def decode_uplink(data):
    if len(data) < 38:
        raise ValueError("Insufficient data length")
    return {
        "no2": to_short(data, 0),
        "no2_op1": to_short(data, 1) / 10,
        "no2_op2": to_short(data, 2) / 10,
        "ox": to_short(data, 3),
        "ox_op1": to_short(data, 4) / 10,
        "ox_op2": to_short(data, 5) / 10,
        "humidity": to_short(data, 6) / 100,
        "pressure": to_short(data, 7) / 10,
        "temperature": to_short(data, 8) / 100,
        "pm_mc_1": to_short(data, 9),
        "pm_mc_2_5": to_short(data, 10),
        "pm_mc_4": to_short(data, 11),
        "pm_mc_10": to_short(data, 12),
        "pm_nc_0_5": to_short(data, 13),
        "pm_nc_1": to_short(data, 14),
        "pm_nc_2_5": to_short(data, 15),
        "pm_nc_4": to_short(data, 16),
        "pm_nc_10": to_short(data, 17),
        "pm_typical_size": to_short(data, 18) / 1000
    }


def process(data, msg):
    if msg.metadata['f_port'] != 1:
        return b''
    try:
        measurements = decode_uplink(msg['Payload'])
        for k, v in measurements.items():
            msg.new_measurement().set_sensor(sensor[k]).set_value(v, k, uom[k]).add()
    except Exception as e:
        raise ValueError(f"Error in processing: {e}")
    return b''
