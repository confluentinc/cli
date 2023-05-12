..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_link_configuration_list:

confluent kafka link configuration list
---------------------------------------

Description
~~~~~~~~~~~

List cluster link configurations.

::

  confluent kafka link configuration list <link> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --cluster string       Kafka cluster ID.
            --environment string   Environment ID.
            --context string       CLI context name.
        -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
            --ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
            --client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
            --client-key-path string    Path to client private key, include for mTLS authentication.
            --no-authentication         Include if requests should be made without authentication headers, and user will not be prompted for credentials.
            --prompt                    Bypass use of available login credentials and prompt for Kafka Rest credentials.
            --context string            CLI context name.
        -o, --output string             Specify the output format as "human", "json", or "yaml". (default "human")
      
Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_kafka_link_configuration` - Manage cluster link configurations.
