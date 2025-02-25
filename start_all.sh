#!/bin/bash

# Check if parameters are provided
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ] || [ -z "$4" ]; then
  echo "Usage: $0 <N> <m> <B> <mode>"
  exit 1
fi

N="$1"
m="$2"
B="$3"
mode="$4"

for (( i=0; i<N*m; i++ ))
do
  ./start_one.sh $i $B $mode 0 &
done

wait

go run ./cmd/performance/performanceCal.go