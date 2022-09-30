#!/bin/bash
# echo 'Running Pipeline CLI Workflow'
# echo '' 
# sleep 1

echo 'confluent pipeline list' 
dist/confluent_darwin_amd64/confluent pipeline list
# sleep 3
echo '' 

echo 'confluent pipeline create --name test-name --description test-description --ksql lksqlc-kkqv3m'
pipeline=`dist/confluent_darwin_amd64/confluent pipeline create --name test-name --description test-description --ksql lksqlc-kkqv3m | cut -d ' ' -f 3`
echo 'Created pipeline:' ${pipeline}
# sleep 3
echo '' 

echo 'confluent pipeline list' 
dist/confluent_darwin_amd64/confluent pipeline list
# sleep 3
echo '' 

echo 'confluent pipeline activate --id' ${pipeline}
# dist/confluent_darwin_amd64/confluent pipeline activate --id $pipeline
# sleep 3
echo 'Activated pipeline:' ${pipeline}
echo ''

# echo 'confluent pipeline list' 
# dist/confluent_darwin_amd64/confluent pipeline list
# sleep 3
# echo '' 

echo 'confluent pipeline deactivate --id' ${pipeline} 
# dist/confluent_darwin_amd64/confluent pipeline deactivate --id $pipeline
# sleep 3
echo 'Deactivated pipeline:' ${pipeline}
echo '' 

# echo 'confluent pipeline list' 
# dist/confluent_darwin_amd64/confluent pipeline list
# sleep 3
# echo '' 

echo 'confluent pipeline delete --id' ${pipeline}
dist/confluent_darwin_amd64/confluent pipeline delete --id $pipeline
# sleep 3
echo '' 

echo 'confluent pipeline list' 
dist/confluent_darwin_amd64/confluent pipeline list
# sleep 3
echo '' 

# echo 'Completed Pipeline CLI Workflow'
# echo ''
