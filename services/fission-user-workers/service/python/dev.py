from flask import Flask
from base import main

app = Flask(__name__)
app.add_url_rule("/", "usercode", main, methods=['GET', 'POST', 'PUT', 'HEAD', 'PATCH', 'DELETE', 'OPTIONS'])
