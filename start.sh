#!/bin/bash

# Check if both parameters N and B are provided
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Usage: $0 <N> <m> <B>"
  exit 1
fi

# Assign the first argument to variable B
N="$1"
m="$2"
B="$3"

# Directory containing the config files
config_dir="/home/hiddener/Chamael/configs"

go run ./cmd/txsMaker --shard_num 3 --tx_num 1000 --Rrate 10

# Loop from 0 to (N*m)-1
for (( i=0; i<N*m; i++ ))
do
  config_file="$config_dir/config_$i.yaml"
  # Check if the config file exists
  if [ -f "$config_file" ]; then
    echo "Using config file: $config_file"
    # Run the Go program with the N and B parameters and the config file in the background
    go run ./cmd/main "$B" "$config_file" &
  else
    echo "Config file $config_file not found"
  fi
done