# INX-Collector Instructions

## Requirements:

- A synced and healthy HORNET node connected to the SHIMMER network
- An S3-compliant object storage (_minio_ is used in this example)

## Set up:

To set up the plugin an S3-compliant object storage must be running and accessible for the plugin to connect to. In this example we provide a demonstration through the use of a _minio_ local storage deployed using the same docker_compose.yml used to deploy the HORNET node.

To do that just add a service to the yml file:

```yml
services:

  minio:
    container_name: minio
    image: minio/minio
    stop_grace_period: 5m
    volumes:
      - minio_storage:/data
    environment:
      MINIO_ROOT_USER: your_access_id
      MINIO_ROOT_PASSWORD: your_password
    command: server --console-address ":9001" /data
```
[If you want to see the minio console from your browser, you should match port 9001 with an available port on your local machine]

To run a collector plugin, the docker image is already been built as `giordyfish/inx-collector:1.1.0`, we need to also add it to the .yml file: 

```yml
services:

  inx-collector:
    container_name: inx-collector
    image: giordyfish/inx-collector:1.1.0
    stop_grace_period: 5m
    restart: unless-stopped
    depends_on:
      hornet:
        condition: service_healthy
      minio:
        condition: service_started
    command:
      - "--inx.address=hornet:9029"
      - "--restAPI.bindAddress=inx-collector:9030"
      - "--storage.endpoint=${STORAGE_ENDPOINT:-minio:9000}"
      - "--storage.accessKeyID=${STORAGE_ACCESS_ID:-your_access_id}"
      - "--storage.secretAccessKey=${STORAGE_SECRET_KEY:-your_password}"
      - "--storage.region=${STORAGE_REGION:-eu-south-1}"
      - "--storage.objectExtension=${STORAGE_EXTENSION:-}"
      - "--storage.secure=${STORAGE_SECURE:-false}"
      - "--storage.defaultBucketName=${STORAGE_DEFAULT_BUCKET:-shimmer-mainnet-default}"
      - "--storage.defaultBucketExpirationDays=${STORAGE_DEFAULT_EXPIRATION:-30}"
      - "--listener.filters=${LISTENER_FILTERS:-}"
      - "--POI.hostUrl=${POI_URL:-http://inx-poi:9687}"
      - "--POI.isPlugin=${POI_PLUGIN:-true}"
```

Now that the docker-compose file has been modified, the plugin can be run by entering the command:

```bash
docker compose up -d 
```

This command can also be run if the HORNET node is already up and running, the plugin will just start on its side.
The log output can be followed with:

```bash
docker logs -f inx-collector 
```

The plugin can be stopped at any time without stopping the HORNET node by entering:

```bash
docker stop inx-collector 
```

### Tunable parameters

All the parameters can be configured by setting environment variables in the .env file. Example:
```env
STORAGE_ENDPOINT=minio:9000
STORAGE_ACCESS_ID=yourID
STORAGE_SECRET_KEY=yourKey
STORAGE_SECURE=false
STORAGE_REGION=eu-south-1
STORAGE_EXTENSION=.json
STORAGE_DEFAULT_BUCKET=shimmer-mainnet-default
STORAGE_DEFAULT_EXPIRATION=30

LISTENER_FILTERS={"filters":[{"tag":"testTag","publicKey":"7a882de7592ad1d6af7d19153b964f35891e2bdbc2e56beea659222b679781cc","duration":"20h","withPOI":true},{"tag":"testTag2"},{"tag":"testTag3", "bucketName":"test-bucket-1"}]}

POI_URL=inx-poi:9687
POI_PLUGIN:true
```

#### STORAGE parameters:

|          Parameter          |                           Description                          |         Default         |      Env_variable_name     |
|:---------------------------:|:--------------------------------------------------------------:|:-----------------------:|:--------------------------:|
|           endpoint          |             defines the endpoint for the S3 storage            |        minio:9000       |      STORAGE_ENDPOINT      |
|         accessKeyId         |            defines the access id for the S3 storage            |            ""           |      STORAGE_ACCESS_ID     |
|       secretAccessKey       | defines the password for the given access id of the S3 storage |            ""           |     STORAGE_SECRET_KEY     |
|            region           |              defines the region of the S3 storage              |        eu-south-1       |       STORAGE_REGION       |
|            secure           |  defines whether the connection to S3 storage should be secure |           true          |       STORAGE_SECURE       |
|       objectExtension       |    sets the file extension for the object inside the storage   |            ""           |      STORAGE_EXTENSION     |
|      defaultBucketName      |                 sets the default bucket's name                 | shimmer-mainnet-default |   STORAGE_DEFAULT_BUCKET   |
| defaultBucketExpirationDays |            sets the default bucket's expiration days           |            30           | STORAGE_DEFAULT_EXPIRATION |

#### POI parameters:

| Parameter |                                     Description                                    |    Default   | Env_variable_name |
|:---------:|:----------------------------------------------------------------------------------:|:------------:|:-----------------:|
|  hostUrl  |                        defines the url of an exposed POI API                       | inx-poi:9687 |      POI_URL      |
|  isPlugin | defines whether the POI host is a POI plugin or a hornet node with an active plugin |     true     |     POI_PLUGIN    |

#### LISTENER parameters:

| Parameter |                Description               | Default | Env_variable_name |
|:---------:|:----------------------------------------:|:-------:|:-----------------:|
|  filters  | a json string which sets startup filters |    ""   |  LISTENER_FILTERS |

#### RESTapi parameters:

|         Parameter         |                                       Description                                      |     Default    |
|:-------------------------:|:--------------------------------------------------------------------------------------:|:--------------:|
|        bindAddress        |           defines the bind address on which the Collector HTTP server listens          | localhost:9030 |
|      advertiseAddress     | defines the address of the Collector HTTP server which is advertised to the INX Server |       ""       |
| debugRequestLoggerEnabled |            defines whether the debug logging for requests should be enabled            |      false     |

## Usage:

You can interact with the plugin using the provided REST APIs. 

### REST API

API documentation is available [here](https://app.swaggerhub.com/apis-docs/Giordyfish/inx-collector/1.1.0)
