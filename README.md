# go-epik-data

go-epik-data handle the epik's expert data retrieved and replay the graph sql.
Then users can explorer the graph data in brower.

## usage

### build 

```
make
```

### config

```
app:
    name: graph
    log_level: debug
    log_dir: logs
    key_path: ~/.ssh/id_rsa

storage:
    db_dir: .epikgraphdata
    data_dir: data

server:
    port: 8080

chains:
-   
    ssh_host: "xx.xx.xx.xx"
    ssh_port: 22
    ssh_user: ""#root/
    miner: "f0xxx"
    rpc_host: "http://xxx"
    rpc_token: "xxx"

nebula:
    address: xx.xx.xx.xx
    port: 9669
    user_name: root
    password:
```

### start service

```
./epik-graph
```

### open the graph explorer

After `epik-graph` node start, open `epik-graph-explorer/index.html` to browse graph data.