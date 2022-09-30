CREATE SOURCE CONNECTOR "DatagenSourceConnector_jv-lineage" WITH (
  "connector.class"='DatagenSource',
  "kafka.api.key"='*****************',
  "kafka.api.secret"='*****************',
  "kafka.auth.mode"='KAFKA_API_KEY',
  "kafka.topic"='jv-lineage-topic',
  "output.data.format"='JSON_SR',
  "quickstart"='ORDERS',
  "tasks.max"='1'
);

CREATE OR REPLACE STREAM "jv-lineage-topic" 
  WITH (kafka_topic='jv-lineage-topic', partitions=1, value_format='JSON_SR');

CREATE OR REPLACE STREAM "jv-filtered-lineage"
  WITH (kafka_topic='jv-filtered-lineage', partitions=1, value_format='JSON_SR')
  AS SELECT * FROM "jv-lineage-topic" WHERE orderunits>7;