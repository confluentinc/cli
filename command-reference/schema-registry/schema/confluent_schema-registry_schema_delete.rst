..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_schema_delete:

confluent schema-registry schema delete
---------------------------------------

Description
~~~~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      Delete one or more schema versions. This command should only be used if absolutely necessary.
      
      ::
      
        confluent schema-registry schema delete [flags]
      
   .. group-tab:: On-Prem
   
      Delete one or more schemas. This command should only be used if absolutely necessary.
      
      ::
      
        confluent schema-registry schema delete [flags]
      
Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --subject string       REQUIRED: Subject of the schema.
            --version string       REQUIRED: Version of the schema. Can be a specific version, "all", or "latest".
            --permanent            Permanently delete the schema.
            --api-key string       API key.
            --api-secret string    API key secret.
            --force                Skip the deletion confirmation prompt.
            --context string       CLI context name.
            --environment string   Environment ID.
      
   .. group-tab:: On-Prem
   
      ::
      
            --subject string                    REQUIRED: Subject of the schema.
            --version string                    REQUIRED: Version of the schema. Can be a specific version, "all", or "latest".
            --permanent                         Permanently delete the schema.
            --ca-location string                File or directory path to CA certificate(s) to authenticate the Schema Registry client.
            --schema-registry-endpoint string   The URL of the Schema Registry cluster.
            --force                             Skip the deletion confirmation prompt.
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
   
      Soft delete the latest version of subject "payments".
      
      ::
      
        confluent schema-registry schema delete --subject payments --version latest
      
   .. group-tab:: On-Prem
   
      Soft delete the latest version of subject "payments".
      
      ::
      
        confluent schema-registry schema delete --subject payments --version latest --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
See Also
~~~~~~~~

* :ref:`confluent_schema-registry_schema` - Manage Schema Registry schemas.
