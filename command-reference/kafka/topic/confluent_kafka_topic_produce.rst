..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_topic_produce:

confluent kafka topic produce
-----------------------------

Description
~~~~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      Produce messages to a Kafka topic.
      
      When using this command, you cannot modify the message header, and the message header will not be printed out.
      
      ::
      
        confluent kafka topic produce <topic> [flags]
      
   .. group-tab:: On-Prem
   
      Produce messages to a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.
      
      When using this command, you cannot modify the message header, and the message header will not be printed out.
      
      ::
      
        confluent kafka topic produce <topic> [flags]
      
Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --schema string                       The path to the schema file.
            --schema-id int32                     The ID of the schema.
            --value-format string                 Format of message value as "string", "avro", "jsonschema", or "protobuf". Note that schema references are not supported for avro. (default "string")
            --references string                   The path to the references file.
            --parse-key                           Parse key from the message.
            --delimiter string                    The delimiter separating each key and value. (default ":")
            --config strings                      A comma-separated list of configuration overrides ("key=value") for the producer client.
            --config-file string                  The path to the configuration file (in json or avro format) for the producer client.
            --schema-registry-endpoint string     Endpoint for Schema Registry cluster.
            --schema-registry-api-key string      Schema registry API key.
            --schema-registry-api-secret string   Schema registry API key secret.
            --api-key string                      API key.
            --api-secret string                   API key secret.
            --cluster string                      Kafka cluster ID.
            --context string                      CLI context name.
            --environment string                  Environment ID.
        -o, --output string                       Specify the output format as "human", "json", or "yaml". (default "human")
      
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
            --schema string                     The path to the local schema file.
            --value-format string               Format of message value as "string", "avro", "jsonschema", or "protobuf". Note that schema references are not supported for avro. (default "string")
            --references string                 The path to the references file.
            --parse-key                         Parse key from the message.
            --delimiter string                  The delimiter separating each key and value. (default ":")
            --config strings                    A comma-separated list of configuration overrides ("key=value") for the producer client.
            --config-file string                The path to the configuration file (in json or avro format) for the producer client.
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
   
      No examples.
      
   .. group-tab:: On-Prem
   
      Produce message to topic "my_topic" with SASL_SSL/PLAIN protocol (providing username and password).
      
      ::
      
        confluent kafka topic produce my_topic --protocol SASL_SSL --sasl-mechanism PLAIN --bootstrap "localhost:19091" --username user --password secret --ca-location my-cert.crt
      
      Produce message to topic "my_topic" with SSL protocol, and SSL verification enabled.
      
      ::
      
        confluent kafka topic produce my_topic --protocol SSL --bootstrap "localhost:18091" --ca-location my-cert.crt
      
See Also
~~~~~~~~

* :ref:`confluent_kafka_topic` - Manage Kafka topics.
