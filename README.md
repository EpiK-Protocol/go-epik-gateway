# EpiK Gateway

EpiK Gateway will help you retrieve and index AI data from EpiK Protocol network.

## Usage

### Start EpiK Node
Graph data is stored in EpiK's node. Epik's Mainnet nodes need to be installed and synchronized before use. Refer to the node documentation for starting the [Mainnet](https://github.com/EpiK-Protocol/go-epik/wiki/How-to-join-Mainnet).

#### add EPK pledge for data retrieve.
After the node is started, a data index pledge needs to be added to the default wallet. Refer to the pledge document for data index [pledge](https://github.com/EpiK-Protocol/go-epik/wiki/How-to-join-Mainnet#8-pledge-for-retrieval).

### Start Nebula Database
After the graph data is indexed from the node, it needs to be stored in the graph database. EpiK uses Nebula as its diagram database. Install the [Nebula](https://docs.nebula-graph.io/2.6.1/) database before starting the node.

### Build 
Build your gateway node with [go](https://go.dev/) installed.

```
make
```

### Config
Your gateway node needs to be connected with an [EpiK](https://github.com/epiK-Protocol/go-epik) node and a [Nebula](https://docs.nebula-graph.io/2.6.1/) node to retrieve AI data from EpiK Protocol network and import it into a nebula graph database.

```
# gateway node
app:
    name: graph
    log_level: debug
    log_dir: logs
    key_path: ~/.ssh/id_rsa

storage:
    db_dir: .epikgraphdata #graph storage path.
    data_dir: data #local data storage path.

server:
    port: 8080 #local graph sever port.

# epik node
chains: 
    ssh_host: "xx.xx.xx.xx" #epik node host
    ssh_port: 22 # epik node port
    ssh_user: "root" # epik node user
    miner: "f0xxx" #retrieve miner
    rpc_host: "http://xxx" #epik node rpc host,eg:http://xxx.xxx.xxx.xxx:1234
    rpc_token: "xxx" # epik node api token.

# nebula node
nebula:
    address: xx.xx.xx.xx
    port: 9669
    user_name: root
    password: *******
```

### Start Gateway

```
./epik-gateway
```

### Explore AI Data

After `epik-gateway` node start, open `epik-graph-explorer/index.html` to browse graph data.

### Expert AI Data snapshot

All AI data is exported and snapshots of the data are stored and available for download.

#### CrossmodalSearch

expert:f0156987

* [CrossmodalSearch_vertex.csv](https://s3.ap-northeast-1.amazonaws.com/cdn.epikg.com/expert-data/20220413/CrossmodalSearch_vertex.csv)
* [CrossmodalSearch_edge.csv](https://s3.ap-northeast-1.amazonaws.com/cdn.epikg.com/expert-data/20220413/CrossmodalSearch_edge.csv)

#### SmartTransport

expert: f0156829

* [SmartTransport_vertex.csv](https://s3.ap-northeast-1.amazonaws.com/cdn.epikg.com/expert-data/20220413/SmartTransport_vertex.csv)
* [SmartTransport_edge.csv](https://s3.ap-northeast-1.amazonaws.com/cdn.epikg.com/expert-data/20220413/SmartTransport_edge.csv)

#### GeneralVoice

expert: f01111430

* [GeneralVoice_vertex.csv](https://s3.ap-northeast-1.amazonaws.com/cdn.epikg.com/expert-data/20220413/GeneralVoice_vertex.csv)
* [GeneralVoice_edge.csv](https://s3.ap-northeast-1.amazonaws.com/cdn.epikg.com/expert-data/20220413/GeneralVoice_edge.csv)

### Data Import

You can import graph data such as [neo4j](https://neo4j.com/developer/guide-import-csv/) using snapshot files.

Import Guide:

* Install [neo4j](https://neo4j.com/docs/getting-started/current/get-started-with-neo4j/)
* Split vertex csv files of domain experts and export them into entity ids and attributes according to different labels
* Import data into Neo4J by entity and relationship ids and attributes.
    * Example:

    ```
    
    // 1. load crossmodalSearch Commodity vertex
    LOAD CSV WITH HEADERS FROM 'file:///crossmodalSearch_commodity_vertex.csv' AS row MERGE (e:Commodity {id: row.Id, value: row.Value})
    
    // 2. load crossmodalSearch other vertex

    // 3. load crossmodalSearch edge
   LOAD CSV WITH HEADERS FROM "file:///crossmodalSearch_edge.csv" AS line match (from:person{id:line.src}),(to:person{id:line.dst})
   merge (from)-[r:rel{name:line.name}]->(to)

    ```
