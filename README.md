# go-epik-gateway

go-epik-gateway handle the epik's expert data retrieved and replay the graph sql.
Then users can explorer the graph data in brower.

## Usage

### Build 
build the gateway node with [go](https://go.dev/) installed.

```
make
```

### Config
The EpiK Gateway needs to be configured with the [EpiK](https://github.com/epiK-Protocol/go-epik) and [NEBULA](https://docs.nebula-graph.com.cn/2.6.1/) nodes to enable the Gateway to retrieve domain expert data and import it into nebula's diagram database.

```
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

# epik node config
chains: 
-   
    ssh_host: "xx.xx.xx.xx" #epik node host
    ssh_port: 22 # epik node port
    ssh_user: "root" # epik node user
    miner: "f0xxx" #retrieve miner
    rpc_host: "http://xxx" #epik node rpc host,eg:http://xxx.xxx.xxx.xxx:1234
    rpc_token: "xxx" # epik node api token.

# nebula graph sql config
nebula:
    address: xx.xx.xx.xx
    port: 9669
    user_name: root
    password:
```

### Start Gateway

```
./epik-gateway
```

### Explore AI Data

After `epik-gateway` node start, open `epik-graph-explorer/index.html` to browse graph data.
