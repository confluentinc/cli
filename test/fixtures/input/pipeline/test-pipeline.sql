CREATE STREAM `upstream` (id INTEGER, name STRING) WITH (kafka_topic = 'topic', partitions=1, value_format='JSON');

CREATE STREAM `downstream` AS SELECT * FROM upstream;