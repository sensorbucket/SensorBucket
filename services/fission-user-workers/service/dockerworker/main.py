import requests
import os
import sys
import importlib
from zipfile import ZipFile
from base64 import b64decode
import pika
import flask
import time

app = flask.Flask("dockerworker")


def main():
    print("Worker starting...\nValidating environment values...")
    # Load all environment variables
    code_entrypoint = os.environ.get("CODE_ENTRYPOINT") or "main.main"
    code_entrypoint_parts = code_entrypoint.split('.')
    code_file = code_entrypoint_parts[0] if len(code_entrypoint_parts) >= 1 else "main"
    code_func = code_entrypoint_parts[1] if len(code_entrypoint_parts) >= 2 else "main"

    worker_id = os.environ.get("WORKER_ID")
    if worker_id is None:
        print("WORKER_ID variable is required")
        sys.exit(1)

    amqp_url = os.environ.get("AMQP_HOST")
    if amqp_url is None:
        print("AMQP_HOST variable is required")
        sys.exit(1)
    amqp_xchg = os.environ.get("AMQP_XCHG")
    if amqp_xchg is None:
        print("AMQP_XCHG variable is required")
        sys.exit(1)

    code_url = os.environ.get("CODE_URL")
    if code_url is None:
        print("CODE_URL environment variable must be given")
        sys.exit(1)

    print(f"Downloading source from url: {code_url}")
    # fetch usercode and extract
    res = requests.get(code_url)
    res.raise_for_status()
    b64_zipped_source = res.json().get("data")
    if b64_zipped_source is None:
        print("code url returned none data")
        sys.exit(1)
    print("Extracting source...")
    with open("/tmp/code.zip", 'wb') as usercode_zip_file:
        usercode_zip_file.write(b64decode(b64_zipped_source))
    with ZipFile("/tmp/code.zip", 'r') as zipObj:
        zipObj.extractall("/usercode")

    if '/usercode' not in sys.path:
        sys.path.append('/usercode')

    print(f"Loading method '{code_func}' from file '/usercode/{code_file}.py'")
    usermod = importlib.machinery.SourceFileLoader('mod', f'/usercode/{code_file}.py').load_module()
    userfunc = getattr(usermod, code_func)
    if userfunc is None:
        print(f"Could not load '{code_func}' from file /usercode/{code_file}.py function from code")
        sys.exit(1)

    # Create callback closures
    def on_message(channel, method_frame, header_frame, body):
        print("Processing message...")
        data = body
        with app.test_request_context(data=data):
            response = userfunc()
        response_topic = response.headers["X-AMQP-Topic"]
        response_data = response.data
        channel.basic_ack(delivery_tag=method_frame.delivery_tag)
        channel.basic_publish(amqp_xchg, response_topic, response_data, pika.BasicProperties(
            message_id=header_frame.message_id,
            headers={"timestamp": int(time.time() * 1000)}
        ))
        print(f"Message processed and published at {amqp_xchg}/{response_topic}")

    # Load AMQP connection
    print("Readying AMQP connection...")
    parameters = pika.URLParameters(amqp_url)
    conn = pika.BlockingConnection(parameters=parameters)
    chan = conn.channel()
    chan.queue_declare("worker_" + worker_id, False, True, False, False, {"x-queue-type": "quorum"})
    chan.queue_bind("worker_" + worker_id, amqp_xchg, worker_id)
    chan.basic_consume("worker_" + worker_id, on_message)
    print("Worker ready. Starting consuming....")

    try:
        chan.start_consuming()
    except KeyboardInterrupt:
        chan.stop_consuming()
    conn.close()


if __name__ == "__main__":
    main()
