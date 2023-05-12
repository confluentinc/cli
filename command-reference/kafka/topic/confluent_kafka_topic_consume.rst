..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_topic_consume:

confluent kafka topic consume
-----------------------------

Description
~~~~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      Consume messages from a Kafka topic.
      
      Truncated message headers will be printed if they exist.
      
      ::
      
        confluent kafka topic consume <topic> [flags]
      
   .. group-tab:: On-Prem
   
      Consume messages from a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.
      
      Truncated message headers will be printed if they exist.
      
      ::
      
        confluent kafka topic consume <topic> [flags]
      
Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --group string                        Consumer group ID. (default "confluent_cli_consumer_<randomly-generated-id>")
        -b, --from-beginning                      Consume from beginning of the topic.
            --offset int                          The offset from the beginning to consume from.
            --partition int32                     The partition to consume from. (default -1)
            --value-format string                 Format of message value as "string", "avro", "jsonschema", or "protobuf". Note that schema references are not supported for avro. (default "string")
            --print-key                           Print key of the message.
            --full-header                         Print complete content of message headers.
            --delimiter string                    The delimiter separating each key and value. (default "\t")
            --timestamp                           Print message timestamp in milliseconds.
            --config strings                      A comma-separated list of configuration overrides ("key=value") for the consumer client.
            --config-file string                  The path to the configuration file (in json or avro format) for the consumer client.
            --schema-registry-context string      The Schema Registry context under which to look up schema ID.
            --schema-registry-endpoint string     Endpoint for Schema Registry cluster.
            --schema-registry-api-key string      Schema registry API key.
            --schema-registry-api-secret string   Schema registry API key secret.
            --api-key string                      API key.
            --api-secret string                   API key secret.
            --cluster string                      Kafka cluster ID.
            --context string                      CLI context name.
            --environment string                  Environment ID.
      
   .. group-tab:: On-Prem
   
      ::
      
            --bootstrap string                  REQUIRED: Comma-separated list of broker hosts, each formatted as "host" or "host:port".
            --ca-location string                REQUIRED: File or directory path to CA certificate(s) for SSL verifying the broker's key.
            --username string                   SASL_SSL username for use with PLAIN mechanism.
            --password string                   SASL_SSL password for use with PLAIN mechanism.
            --cert-location string              Path to client's public key (PEM) used for SSL authentication.
            --key-location string               Path to client's private key (PEM) used for SSL authentication.
            --key-password string               Private key passphrase for SSL authentication.
            --protocol string                   Security protocol used to communicate with brokers. (default "SSL")
            --sasl-mechanism string             SASL_SSL mechanism used for authentication. (default "PLAIN")
            --group string                      Consumer group ID.
        -b, --from-beginning                    Consume from beginning of the topic.
            --offset int                        The offset from the beginning to consume from.
            --partition int32                   The partition to consume from. (default -1)
            --value-format string               Format of message value as "string", "avro", "jsonschema", or "protobuf". Note that schema references are not supported for avro. (default "string")
            --print-key                         Print key of the message.
            --full-header                       Print complete content of message headers.
            --timestamp                         Print message timestamp in milliseconds.
            --delimiter string                  The delimiter separating each key and value. (default "\t")
            --config strings                    A comma-separated list of configuration overrides ("key=value") for the consumer client.
            --config-file string                The path to the configuration file (in json or avro format) for the consumer client.
            --schema-registry-endpoint string   The URL of the Schema Registry cluster.
        -o, --output string                     Specify the output format as "human", "json", or "yaml". (default "human")
      
Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      Consume items from the "my_topic" topic and press "Ctrl-C" to exit.
      
      ::
      
        confluent kafka topic consume -b my_topic
      
   .. group-tab:: On-Prem
   
      Consume message from topic "my_topic" with SSL protocol and SSL verification enabled (providing certificate and private key).
      
      ::
      
        confluent kafka topic consume my_topic --protocol SSL --bootstrap "localhost:19091" --ca-location my-cert.crt --cert-location client.pem --key-location client.key
      
      Consume message from topic "my_topic" with SASL_SSL/OAUTHBEARER protocol enabled (using MDS token).
      
      ::
      
        confluent kafka topic consume my_topic --protocol SASL_SSL --sasl-mechanism OAUTHBEARER --bootstrap "localhost:19091" --ca-location my-cert.crt
      
See Also
~~~~~~~~

* :ref:`confluent_kafka_topic` - Manage Kafka topics.
