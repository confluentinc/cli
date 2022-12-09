CREATE STREAM `upstream` (id integer, name string) WITH (kafka_topic = 'topic', partitions=1, value_format='JSON');

CREATE STREAM `downstream` AS SELECT * FROM upstream;