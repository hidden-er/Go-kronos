# Chamael

Go implementation of Kronos

## Usage

generate config files:
``` bash
go run ./cmd/configMaker/configMaker.go
```

start nodes:
``` bash
./start_all.sh 4 3 1000 1
```

calculate performance:
``` bash
go run ./cmd/performance/performanceCal.go
```