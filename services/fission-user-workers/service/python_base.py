import flask
from flask import request
from flask import current_app
import usercode


def main():
    res = flask.Response(usercode.process(request.get_data(as_text=True)))
    res.headers["X-AMQP-Topic"] = "success"
    return res
