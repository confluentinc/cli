#!/bin/bash

# run sed inside this script to get around difficulties in running it directly inside a post hook
sed "s|path/to/confluent|$1|" scripts/gon_confluent_darwin.hcl > dist/gon_confluent_darwin_$2.hcl
