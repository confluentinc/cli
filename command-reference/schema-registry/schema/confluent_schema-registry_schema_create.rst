..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_schema_create:

confluent schema-registry schema create
---------------------------------------

Description
~~~~~~~~~~~

Create a schema.

::

  confluent schema-registry schema create [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --schema string        REQUIRED: The path to the schema file.
            --subject string       REQUIRED: Subject of the schema.
            --type string          Specify the schema type as "avro", "json", or "protobuf".
            --references string    The path to the references file.
            --api-key string       API key.
            --api-secret string    API key secret.
            --context string       CLI context name.
            --environment string   Environment ID.
        -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --schema string                     REQUIRED: The path to the schema file.
            --subject string                    REQUIRED: Subject of the schema.
            --type string                       Specify the schema type as "avro", "json", or "protobuf".
            --references string                 The path to the references file.
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
   
      Register a new schema.
      
      ::
      
        confluent schema-registry schema create --subject payments --schema payments.avro --type avro
      
      Where ``schemafilepath`` may include these contents:
      
      ::
      
        {
        	"type" : "record",
        	"namespace" : "Example",
        	"name" : "Employee",
        	"fields" : [
        		{ "name" : "Name" , "type" : "string" },
        		{ "name" : "Age" , "type" : "int" }
        	]
        }
      
      For more information on schema types, see https://docs.confluent.io/current/schema-registry/serdes-develop/index.html.
      
      For more information on schema references, see https://docs.confluent.io/current/schema-registry/serdes-develop/index.html#schema-references.
      
   .. group-tab:: On-Prem
   
      Register a new schema.
      
      ::
      
        confluent schema-registry schema create --subject payments --schema payments.avro --type avro --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
See Also
~~~~~~~~

* :ref:`confluent_schema-registry_schema` - Manage Schema Registry schemas.
