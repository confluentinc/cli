..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_topic_create:

confluent kafka topic create
----------------------------

Description
~~~~~~~~~~~

Create a Kafka topic.

::

  confluent kafka topic create <topic> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --partitions uint32    Number of topic partitions.
            --config strings       A comma-separated list of configuration overrides ("key=value") for the topic being created.
            --dry-run              Run the command without committing changes to Kafka.
            --if-not-exists        Exit gracefully if topic already exists.
            --cluster string       Kafka cluster ID.
            --context string       CLI context name.
            --environment string   Environment ID.
      
   .. group-tab:: On-Prem
   
      ::
      
            --url string                  Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
            --ca-cert-path string         Path to a PEM-encoded CA to verify the Confluent REST Proxy.
            --client-cert-path string     Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
            --client-key-path string      Path to client private key, include for mTLS authentication.
            --no-authentication           Include if requests should be made without authentication headers, and user will not be prompted for credentials.
            --prompt                      Bypass use of available login credentials and prompt for Kafka Rest credentials.
            --partitions uint32           Number of topic partitions.
            --replication-factor uint32   Number of replicas.
            --config strings              A comma-separated list of topic configuration ("key=value") overrides for the topic being created.
            --if-not-exists               Exit gracefully if topic already exists.
      
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
   
      Create a topic named "my_topic" with default options.
      
      ::
      
        confluent kafka topic create my_topic
      
   .. group-tab:: On-Prem
   
      Create a topic named "my_topic" with default options for the current specified cluster (providing embedded Kafka REST Proxy endpoint).
      
      ::
      
        confluent kafka topic create my_topic --url http://localhost:8090/kafka
      
      Create a topic named "my_topic" with default options at specified cluster (providing Kafka REST Proxy endpoint).
      
      ::
      
        confluent kafka topic create my_topic --url http://localhost:8082
      
      Create a topic named "my_topic_2" with specified configuration parameters.
      
      ::
      
        confluent kafka topic create my_topic_2 --url http://localhost:8082 --config cleanup.policy=compact,compression.type=gzip
      
See Also
~~~~~~~~

* :ref:`confluent_kafka_topic` - Manage Kafka topics.
