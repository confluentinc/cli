..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_exporter_create:

confluent schema-registry exporter create
-----------------------------------------

Description
~~~~~~~~~~~

Create a new schema exporter.

::

  confluent schema-registry exporter create <name> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --config-file string      REQUIRED: Exporter config file.
            --subjects strings        A comma-separated list of exporter subjects. (default [*])
            --subject-format string   Exporter subject rename format. The format string can contain ${subject}, which will be replaced with default subject name. (default "${subject}")
            --context-type string     Exporter context type. One of "AUTO", "CUSTOM" or "NONE". (default "AUTO")
            --context-name string     Exporter context name.
            --api-key string          API key.
            --api-secret string       API key secret.
            --context string          CLI context name.
            --environment string      Environment ID.
        -o, --output string           Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --config-file string                REQUIRED: Exporter config file.
            --subjects strings                  A comma-separated list of exporter subjects. (default [*])
            --subject-format string             Exporter subject rename format. The format string can contain ${subject}, which will be replaced with default subject name. (default "${subject}")
            --context-type string               Exporter context type. One of "AUTO", "CUSTOM" or "NONE". (default "AUTO")
            --context-name string               Exporter context name.
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
   
      Create a new schema exporter.
      
      ::
      
        confluent schema-registry exporter create my-exporter --config-file config.txt --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type CUSTOM --context-name my-context
      
   .. group-tab:: On-Prem
   
      Create a new schema exporter.
      
      ::
      
        confluent schema-registry exporter create my-exporter --config-file config.txt --subjects my-subject1,my-subject2 --subject-format my-\${subject} --context-type CUSTOM --context-name my-context --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
See Also
~~~~~~~~

* :ref:`confluent_schema-registry_exporter` - Manage Schema Registry exporters.
