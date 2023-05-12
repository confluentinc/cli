..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_schema_list:

confluent schema-registry schema list
-------------------------------------

Description
~~~~~~~~~~~

List schemas for a given subject prefix.

::

  confluent schema-registry schema list [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --subject-prefix string   List schemas for subjects with a given prefix.
            --all                     Include soft-deleted schemas.
            --api-key string          API key.
            --api-secret string       API key secret.
            --context string          CLI context name.
            --environment string      Environment ID.
        -o, --output string           Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --subject-prefix string             List schemas for subjects with a given prefix.
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
   
      List all schemas for subjects with prefix "my-subject".
      
      ::
      
        confluent schema-registry schema list --subject-prefix my-subject
      
      List all schemas for all subjects in context ":.mycontext:".
      
      ::
      
        confluent schema-registry schema list --subject-prefix :.mycontext:
      
      List all schemas in the default context.
      
      ::
      
        confluent schema-registry schema list
      
   .. group-tab:: On-Prem
   
      List all schemas for subjects with prefix "my-subject".
      
      ::
      
        confluent schema-registry schema list --subject-prefix my-subject --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
      List all schemas for all subjects in context ":.mycontext:".
      
      ::
      
        confluent schema-registry schema list --subject-prefix :.mycontext: --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
      List all schemas in the default context.
      
      ::
      
        confluent schema-registry schema list --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
See Also
~~~~~~~~

* :ref:`confluent_schema-registry_schema` - Manage Schema Registry schemas.
