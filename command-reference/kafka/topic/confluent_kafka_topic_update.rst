..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_topic_update:

confluent kafka topic update
----------------------------

Description
~~~~~~~~~~~

Update a Kafka topic.

::

  confluent kafka topic update <topic> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --config strings       A comma-separated list of configuration overrides with form "key=value".
            --dry-run              Run the command without committing changes to Kafka.
            --cluster string       Kafka cluster ID.
            --context string       CLI context name.
            --environment string   Environment ID.
        -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
            --ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
            --client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
            --client-key-path string    Path to client private key, include for mTLS authentication.
            --no-authentication         Include if requests should be made without authentication headers, and user will not be prompted for credentials.
            --prompt                    Bypass use of available login credentials and prompt for Kafka Rest credentials.
            --config strings            A comma-separated list of topics configuration ("key=value") overrides for the topic being created.
        -o, --output string             Specify the output format as "human", "json", or "yaml". (default "human")
      
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
   
      Modify the "my_topic" topic to have a retention period of 3 days (259200000 milliseconds).
      
      ::
      
        confluent kafka topic update my_topic --config "retention.ms=259200000"
      
   .. group-tab:: On-Prem
   
      Modify the "my_topic" topic for the specified cluster (providing embedded Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).
      
      ::
      
        confluent kafka topic update my_topic --url http://localhost:8082 --config retention.ms=259200000
      
      Modify the "my_topic" topic for the specified cluster (providing Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).
      
      ::
      
        confluent kafka topic update my_topic --url http://localhost:8082 --config retention.ms=259200000
      
See Also
~~~~~~~~

* :ref:`confluent_kafka_topic` - Manage Kafka topics.
