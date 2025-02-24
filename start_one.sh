#!/bin/bash

# Check if parameters are provided
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Usage: $0 <id> <B> <mode>"
  exit 1
fi

id="$1"
B="$2"
mode="$3"

# Directory containing the config files
config_dir="$HOME/Chamael/configs"

go run ./cmd/txsMaker --id $id --shard_num 3 --tx_num 1000 --Rrate 10


config_file="$config_dir/config_$id.yaml"
# Check if the config file exists
if [ -f "$config_file" ]; then
  echo "Using config file: $config_file"
  # Run the Go program with the N and B parameters and the config file in the background
  go run ./cmd/main "$B" "$config_file" "$mode"
else
  echo "Config file $config_file not found"
fi
