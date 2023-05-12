..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_link_configuration_update:

confluent kafka link configuration update
-----------------------------------------

Description
~~~~~~~~~~~

Update cluster link configurations.

::

  confluent kafka link configuration update <link> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --config-file string   REQUIRED: Name of the file containing link config overrides. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.
            --cluster string       Kafka cluster ID.
            --environment string   Environment ID.
            --context string       CLI context name.
      
   .. group-tab:: On-Prem
   
      ::
      
            --config-file string        REQUIRED: Name of the file containing link config overrides. Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.
            --url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
            --ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
            --client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
            --client-key-path string    Path to client private key, include for mTLS authentication.
            --no-authentication         Include if requests should be made without authentication headers, and user will not be prompted for credentials.
            --prompt                    Bypass use of available login credentials and prompt for Kafka Rest credentials.
            --context string            CLI context name.
      
Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Update configuration values for the cluster link "my-link".

::

  confluent kafka link configuration update my-link --config-file my-config.txt

See Also
~~~~~~~~

* :ref:`confluent_kafka_link_configuration` - Manage cluster link configurations.
