#!/bin/bash

# run sed inside this script to get around difficulties in running it directly inside a post hook
sed "s|path/to/file|$1|" scripts/gon_confluent.hcl > dist/gon_confluent_$2.hcl
