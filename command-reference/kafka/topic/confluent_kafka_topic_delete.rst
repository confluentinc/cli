..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_topic_delete:

confluent kafka topic delete
----------------------------

Description
~~~~~~~~~~~

Delete a Kafka topic.

::

  confluent kafka topic delete <topic> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --force                Skip the deletion confirmation prompt.
            --cluster string       Kafka cluster ID.
            --context string       CLI context name.
            --environment string   Environment ID.
      
   .. group-tab:: On-Prem
   
      ::
      
            --url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
            --ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
            --client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
            --client-key-path string    Path to client private key, include for mTLS authentication.
            --no-authentication         Include if requests should be made without authentication headers, and user will not be prompted for credentials.
            --prompt                    Bypass use of available login credentials and prompt for Kafka Rest credentials.
            --force                     Skip the deletion confirmation prompt.
      
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
   
      Delete the topics "my_topic" and "my_topic_avro". Use this command carefully as data loss can occur.
      
      ::
      
        confluent kafka topic delete my_topic
        confluent kafka topic delete my_topic_avro
      
   .. group-tab:: On-Prem
   
      Delete the topic "my_topic" for the specified cluster (providing embedded Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.
      
      ::
      
        confluent kafka topic delete my_topic --url http://localhost:8090/kafka
      
      Delete the topic "my_topic" for the specified cluster (providing Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.
      
      ::
      
        confluent kafka topic delete my_topic --url http://localhost:8082
      
See Also
~~~~~~~~

* :ref:`confluent_kafka_topic` - Manage Kafka topics.
