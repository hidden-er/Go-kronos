#!/bin/bash

# Check if parameters are provided
if [ -z "$1" ]; then
  echo "Usage: $0 <N>"
  exit 1
fi

N="$1"

config_dir="$HOME/Chamael/configs"

for (( i=0; i<N; i++ ))
do
  config_file="$config_dir/config_$i.yaml"
  echo "Using config file: $config_file"
  go run ./cmd/globalBftTest/globalBftTest.go $config_file "Debug" &
done