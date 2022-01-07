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
