# Go implementation of Kronos

## 1. Usage

Kronos can be run in 3 ways:
- Local
- All nodes in one Docker
- One node in one Docker

### 1.1 Local

Run on your local machine, one node is one process.

Install dependencies:
``` bash
go mod download
```

Generate node config files based on `cmd/main/config_local.yaml`:
``` bash
go run ./cmd/configMaker/configMaker.go -config_path ./cmd/main/config_local.yaml
```

Start all nodes via shell script:
``` bash
./start_all.sh 4 3 1000 1
```


### All nodes in one Docker

Build a docker including all nodes in it.

Generate node config files based on `cmd/main/config_local.yaml`:
``` bash
go run ./cmd/configMaker/configMaker.go -config_path ./cmd/main/config_local.yaml
```

Build Docker image via Dockerfile:
``` bash
docker build -t Kronos:latest .
```

Start Docker:
``` bash
docker run -it --rm Kronos:latest 4 3 1000 1
```

### One node in one Docker

Using Docker compose to run a docker each node.

Generate node config files based on `cmd/main/config_compose.yaml`:
``` bash
go run ./cmd/configMaker/configMaker.go -config_path ./cmd/main/config_compose.yaml
```

Start Docker:
``` bash
docker compose up
```
