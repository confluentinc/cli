..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_exporter_list:

confluent schema-registry exporter list
---------------------------------------

Description
~~~~~~~~~~~

List all schema exporters.

::

  confluent schema-registry exporter list [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --api-key string       API key.
            --api-secret string    API key secret.
            --context string       CLI context name.
            --environment string   Environment ID.
        -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
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

See Also
~~~~~~~~

* :ref:`confluent_schema-registry_exporter` - Manage Schema Registry exporters.
