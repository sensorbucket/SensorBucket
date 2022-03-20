## Pipeline definition
```yaml file="mfm-pipeline.yaml"
version: 1.0 # Version of the pipeline specification

pipeline:
  # What if MFM worker supports multiple sources?
  - name: HTTP Ingress
    source: gitlab.com/sensorbucket/ttn-worker.git@1.0.0
    ingress: true
  - name: TTN Worker
    source: gitlab.com/sensorbucket/ttn-worker.git@1.0.0
    configuration: # Configuration specific to the worker
      introspectURL: https://auth.service.local/introspect
  - name: MFM Worker
    source: gitlab.com/sensorbucket/mfm-worker.git@1.0.0
    configuration: {} # Configuration specific to the worker

```

## Worker definition

```yaml file="ttn-worker.yaml"
version: 1.0 # Version of the worker specification

id: sensorbucket/ttn-worker@1.0.0 # The ID and version of this worker 
description: "A worker to receive and process TTN data"
supports: # Data must be supplied from one of these workers
  - sensorbucket/message-queue
configuration: # Schema defining the possible configuration values for this worker
  introspectURL:
    format: url
    required: true
    description: |
      The introspect endpoint of the authentication service.
      This is used to authenticate incoming requests.
assets: {} # TTN Worker has no assets
```

```yaml file="mfm-worker.yaml"
version: 1.0
id: sensorbucket/mfm-worker@1.0.0
description: "A worker to process MFM data"
supports: # This worker only supports data originating from a ttn-worker
    - sensorbucket/ttn-worker@^1.0.0 # Semver matching: https://devhints.io/semver
configuration: {} # This worker has no configuration
assets: # This worker has a custom asset as defined below
  sensor: 
    type: object
    required: ["index", "devEUI"]
    properties:
      index:
        type: number
        description: The index of the sensor module on the Multiflexmeter
      devEUI:
        type: string
        description: The Device EUI of the Multiflexmeter
        pattern: "[a-fA-F0-9]{16}$"
```