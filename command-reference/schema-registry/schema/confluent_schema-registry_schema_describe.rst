..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_schema_describe:

confluent schema-registry schema describe
-----------------------------------------

Description
~~~~~~~~~~~

Get schema either by schema ID, or by subject/version.

::

  confluent schema-registry schema describe [id] [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --subject string       Subject of the schema.
            --version string       Version of the schema. Can be a specific version or "latest".
            --show-references      Display the entire schema graph, including references.
            --api-key string       API key.
            --api-secret string    API key secret.
            --context string       CLI context name.
            --environment string   Environment ID.
      
   .. group-tab:: On-Prem
   
      ::
      
            --subject string                    Subject of the schema.
            --version string                    Version of the schema. Can be a specific version or "latest".
            --show-references                   Display the entire schema graph, including references.
            --ca-location string                File or directory path to CA certificate(s) to authenticate the Schema Registry client.
            --schema-registry-endpoint string   The URL of the Schema Registry cluster.
            --context string                    CLI context name.
      
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
   
      Describe the schema string by schema ID.
      
      ::
      
        confluent schema-registry schema describe 1337
      
      Describe the schema by both subject and version.
      
      ::
      
        confluent schema-registry schema describe --subject payments --version latest
      
   .. group-tab:: On-Prem
   
      Describe the schema string by schema ID.
      
      ::
      
        confluent schema-registry schema describe 1337 --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
      Describe the schema string by both subject and version.
      
      ::
      
        confluent schema-registry schema describe --subject payments --version latest --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
See Also
~~~~~~~~

* :ref:`confluent_schema-registry_schema` - Manage Schema Registry schemas.
