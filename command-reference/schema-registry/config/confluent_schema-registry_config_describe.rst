..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_config_describe:

confluent schema-registry config describe
-----------------------------------------

Description
~~~~~~~~~~~

Describe top-level or subject-level schema compatibility.

::

  confluent schema-registry config describe [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --subject string       Subject of the schema.
            --api-key string       API key.
            --api-secret string    API key secret.
            --context string       CLI context name.
            --environment string   Environment ID.
        -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --subject string                    Subject of the schema.
            --ca-location string                File or directory path to CA certificate(s) to authenticate the Schema Registry client.
            --schema-registry-endpoint string   The URL of the Schema Registry cluster.
            --context string                    CLI context name.
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
   
      Describe the configuration of subject "payments".
      
      ::
      
        confluent schema-registry config describe --subject payments
      
      Describe the top-level configuration.
      
      ::
      
        confluent schema-registry config describe
      
   .. group-tab:: On-Prem
   
      Describe the configuration of a subject "payments".
      
      ::
      
        confluent schema-registry config describe --subject payments --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
      Describe the top-level configuration.
      
      ::
      
        confluent schema-registry config describe --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
See Also
~~~~~~~~

* :ref:`confluent_schema-registry_config` - Manage Schema Registry configuration.
