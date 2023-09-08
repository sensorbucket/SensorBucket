import flask
from flask import request
from flask import current_app
import usercode


def main():
    current_app.logger.info("Received request")
    res = flask.Response(usercode.process(request.get_data(as_text=True)))
    res.headers["X-AMQP-Topic"] = "success"
    return res
