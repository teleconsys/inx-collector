INX-collector
====================================

INX-collector is an INX plug-in for IOTA Hornet nodes that provides a block storage service for blocks received by the node. In other words, it makes the node a Selective Permanode. INX-collector is currently implemented for the SHIMMER network.

_Collector_ fulfills the need of many applications using the IOTA network as a secure and persistent transport layer. Due to the high performance of the network, the SHIMMER node must prune every block that is not essential to the ledger state, in a nutshell, it deletes every block that does not contain economic transactions after a given amount of time. This implies that two applications exchanging data must in an asynchronous, time-discrete manner, could lose data if it has already been pruned by the network. With _Collector_ this problem is solved: it is possible to define which blocks to keep and for how long, simplifying the task of the receiving application. _Collector_ seems to act as a simple permanode, but it allows to be selective on the data that must be stored, and can also generate and store Proofs of Inclusion (POI) to ensure data integrity.  

The plug-in offers two modes through which it stores blocks:

- retaining a `block` by `block id` via REST API
- retaining all referenced blocks containing a `tagged payload` received by the node from the network. The plugin only stores that with the specified `tags`. If a `publicKey` is also provided, the plugin will store only those messages that provide a valid `signature` for the specified `publicKey` (this feature is implemented using the [`datapayloads`](https://github.com/iotaledger/datapayloads.go) lib). This mode can be set either via REST API (non-persistent) or with a specific configuration string in the configuration file (persistent).

The collected blocks are stored in an object storage that can be either local to the node or remote. In the case of mission-critical application scenarios, several nodes/plug-ins may share the same remote object storage. This allows the client to obtain blocks stored by several alternative nodes and avoids data loss if a certain single node is down. The plug-in only selects Tagged blocks referenced by a `milestone` and, if specified, can also generate and store the Proof of Inclusion (POI). The client can obtain just the block or the full POI, and return them via REST API response in either fashion.

The duration of block storage in the Object Storage is limited by two factors:

- maximum occupancy space of the object storage set in its configuration
- maximum retention duration set at the individual bucket level

Clients can create multiple buckets to store blocks with specific application logic and assign a specific lifecycle rule to each bucket. It is possible to delete blocks via REST API, but not to delete the entire bucket. This must be done from the object storage console.

Filters
---------------------------------

Filters are the objects you can use to continuously listen on the Tangle, according to the filter specification. A filter is a struct with the following fields:

```go
type Filter struct {
  Tag        string
  PublicKey  string    
  Id         string    
  BucketName string   
  WithPOI    bool     
  Duration   string   
}
```
The `Tag` is required, as it is the tag you want to listen to. The `Id` is the `filterId`, it is generated from the software and returned by the API when you create a filter, in this way you can stop that filter using its `Id`. `BucketName` specifies the bucket where the filter stores the blocks. `WithPOI` specifies if the Proof of Inclusion has to be stored. `Duration` specifies the duration of the filter, the string must follow the format specified [here](https://pkg.go.dev/time#ParseDuration), if the `Duration` is empty, the filter will run until is manually stopped. 

### **By using the `PublicKey` field, and by sending `SignedData` using the [`datapayloads lib`](https://github.com/iotaledger/datapayloads.go), you can selectively and automatically store all your application data.**
If you add an ed25519 `PublicKey` to your filter (as a hexadecimal string) the plugin will still listen to the specified `Tag`, but will only store the payloads containing a [`SignedDataContainer`](https://github.com/iotaledger/datapayloads.go/blob/develop/signed_data_container.go) whose `signature` is valid against the `PublicKey`. 

### :warning: **Filters instanced via REST API are not persistent!** :warning:
Filters instanced via API will be lost every time the plugin is shut down. If you want a persistent filter that starts every time the plugin runs, you should set these `startup filters` as an environment variable, the format is that of a JSON string. To understand how to set those filters look at the example provided in the "Tunable parameters" section.

Instructions
---------------------------------

To set up and use inx-collector you need to configure it and attach it to your shimmer node. Then, you can easily interact with it by [`REST APIs`](https://app.swaggerhub.com/apis-docs/Giordyfish/inx-collector/1.1.0). For a detailed set of instructions regarding how to set up your plugin, you can look at the [`INSTRUCTIONS`](INSTRUCTIONS.md).


Contacts
---------------------------------

If you want to get in touch with us feel free to contact us at <g.pescetelli@teleconsys.it>