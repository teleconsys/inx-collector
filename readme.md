# INX-collector

INX-collector is an INX plug-in for IOTA Hornet nodes that provides a block storage service for blocks received by the node. In other words, it makes the node a Selective Permanode. INX-collector is currently implemented for the SHIMMER network.

_Collector_ fulfils the need of many applications using the IOTA network as a secure and persistent transport layer. Due to the high performance of the network, the SHIMMER node must prune every block that is not essential to the ledger state, in a nutshell it deletes every block that do not contain economic transactions after a given amount of time. This implies that two applications exchanging data must in an asynchronous, time-discrete manner, could lose data if it has already been pruned by the network. With _Collector_ this problem is solved: it is possible to define which blocks to keep and for how long, simplifying the task of the receiving application. _Collector_ seems to act as a simple permanode, but it allows to be selective on the data that must be stored, and can also generate and store Proofs of Inclusion (POI) in order to ensure data integrity.  

The plug-in offers two modes trough which it stores blocks:

- retaining a `block` by `block id` via REST api
- retaining all referenced blocks containing a `tagged payload` received by the node from the network. The plugin only stores that with the specified `tags`. This mode can be set either via REST api (non-persistent) or with a specific configuration string in the configuration file (persistent).

The collected blocks are stored in an object storage that can be either local to the node or remote. In the case of mission critical application scenarios, several nodes/plug-ins may share the same remote object storage. This allows the client to obtain blocks stored by several alternative nodes and avoids data loss if certain single node is down. The plug-in only selects Tagged blocks referenced by a `milestone` and, if specified, can also generate and store the Proof of Inclusion (POI). The client can obtain just the block or the full POI, and return them via REST api response in either fashion.

The duration of block storage in the Object Storage is limited by two factors:

- maximum occupancy space of the object storage set in its configuration
- maximum retention duration set at the individual bucket level

Clients can create multiple buckets to store blocks with specific application logic and assign a specific lifecycle rule to each bucket. It is possible to delete blocks via REST api, but not to delete the entire bucket. This must be done from the object storage console.

## Requirements:

- A synced and healthy HORNET node connected to the SHIMMER network
- An S3-compliant object storage (_minio_ is used in this example)

## Set up:

To setup the plugin an S3-compliant object storage must be running and accessible for the plugin to connect to. In this example we provide a demonstration trough the use of a _minio_ local storage deployed using the same docker_compose.yml used to deploy the HORNET node.

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
      MINIO_ROOT_USER: admin
      MINIO_ROOT_PASSWORD: password
    command: server --console-address ":9001" /data
```
[If you want to see the minio console from your browser, you should match port 9001 with an available port on your local machine]

To run a collector plugin, the docker image is already been built as `giordyfish/inx-collector:1.0.0`, we need to also add it to the .yml file: 

```yml
services:

  inx-collector:
    container_name: inx-collector
    image: giordyfish/inx-collector:1.0.0
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
      - "--storage.accessKeyID=${STORAGE_ACCESS_ID:-}"
      - "--storage.secretAccessKey=${STORAGE_SECRET_KEY:-}"
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

This command can also be run if the HORNET node is already up and running, the plugin will just start to its side.
The log output can be followed with:

```bash
docker logs -f inx-collector 
```

The plugin can be stopped at any time without alting the HORNET node by entering:

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

LISTENER_FILTERS={"filters":[{"tag":"testTag","duration":"20h","withPOI":true},{"tag":"testTag2"},{"tag":"testTag3", "bucketName":"test-bucket-1"}]}

POI_URL=inx-poi:9687
POI_PLUGIN:true
```

#### STORAGE parameters:

| .............Parameter............... | ...........Description.............                | .......Default.......   | ..........Env_variable_name.......... |
| ------------------------------------- | -------------------------------------------------------------- | ----------------------- | ------------------------------------- |
| endpoint                              | defines the endpoint for the S3 storage                        | minio:9000              | STORAGE_ENDPOINT                      |
| accessKeyId                           | defines the access id for the S3 storage                       | ""                      | STORAGE_ACCESS_ID                     |
| secretAccessKey                       | defines the password for the given access id of the S3 storage | ""                      | STORAGE_SECRET_KEY                    |
| region                                | defines the region of the S3 storage                           | eu-south-1              | STORAGE_REGION                        |
| secure                                | defines whether the connection to S3 storage should be secure  | true                    | STORAGE_SECURE                        |
| objectExtension                       | sets the file extension for the object inside the storage      | ""                      | STORAGE_EXTENSION                     |
| defaultBucketName                     | sets the default bucket's name                                 | shimmer-mainnet-default | STORAGE_DEFAULT_BUCKET                |
| defaultBucketExpirationDays           | sets the default bucket's expiration days                      | 30                      | STORAGE_DEFAULT_EXPIRATION            |

#### POI parameters:

| .............Parameter............... | ...........Description.............                                    | .......Default....... | ..........Env_variable_name.......... |
| ------------------------------------- | ---------------------------------------------------------------------------------- | --------------------- | ------------------------------------- |
| hostUrl                               | defines the url of an exposed POI API                                              | inx-poi:9687   | POI_URL                               |
| isPlugin                              | defines wether the POI host is a POI plugin or a hornet node with an active plugin | true                  | POI_PLUGIN                            |

#### LISTENER parameters:

| .............Parameter............... | ...........Description............. | .......Default....... | ..........Env_variable_name.......... |
| ------------------------------------- | ----------------------------------------------- | --------------------- | ------------------------------------- |
| filters                               | A json string which sets startup filters        | ""        | LISTENER_FILTERS                      |

#### RESTapi parameters:

| .............Parameter............... | ...........Description.............                                        | .......Default....... | ..........Env_variable_name.......... |
| ------------------------------------- | -------------------------------------------------------------------------------------- | --------------------- |:-------------------------------------:|
| bindAddress                           | defines the bind address on which the Collector HTTP server listens                    | localhost:9030        | \                                     |
| advertiseAddress                      | defines the address of the Collector HTTP server which is advertised to the INX Server | ""                    | \                                     |
| debugRequestLoggerEnabled             | defines whether the debug logging for requests should be enabled                       | false                 | \                                     |

## Usage:

You can interact with the plugin using the provided REST APIs. 

### REST api

API documentation is available [here](https://app.swaggerhub.com/apis-docs/Giordyfish/inx-collector/1.0.0)

### Filters

Filters are the objects you can use to continuously listen on the Tangle, according to the filter specification. A filter is a struct with the following fields:

```go
type Filter struct {
	Tag        string    
	Id         string    
	BucketName string   
	WithPOI    bool     
	Duration   string   
}
```
The `Tag` is requires, as it is the tag you want to listen to. The `Id` is the `filterId`, it is generated from the software and returned by the API when you create a filter, in this way you can stop that filter using its `Id`. `BucketName` specifies the bucket where the filter stores the blocks. `WithPOI` specifies if the Proof of Inclusion has to be stored. `Duration` specifies the duration of the filter, the string must follow the format specified [here](https://pkg.go.dev/time#ParseDuration), if the `Duration` is empty, the filter will run until is manually stopped. 

 :warning: **Filters instanced via REST api are not persistent!**: filters instanced via API will be lost everytime the plugin is shut down. If you want a persistent filter which starts everytime the plugin runs, you should sets these `startup filters` as an environment variable, the format is that of a json string. To understand how to set those filters look at the example provided in the "Tunable parameters" section.