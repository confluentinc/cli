..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_topic_describe:

confluent kafka topic describe
------------------------------

Description
~~~~~~~~~~~

Describe a Kafka topic.

::

  confluent kafka topic describe <topic> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
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
   
      Describe the "my_topic" topic.
      
      ::
      
        confluent kafka topic describe my_topic
      
   .. group-tab:: On-Prem
   
      Describe the "my_topic" topic for the specified cluster (providing embedded Kafka REST Proxy endpoint).
      
      ::
      
        confluent kafka topic describe my_topic --url http://localhost:8090/kafka
      
      Describe the "my_topic" topic for the specified cluster (providing Kafka REST Proxy endpoint).
      
      ::
      
        confluent kafka topic describe my_topic --url http://localhost:8082
      
See Also
~~~~~~~~

* :ref:`confluent_kafka_topic` - Manage Kafka topics.
