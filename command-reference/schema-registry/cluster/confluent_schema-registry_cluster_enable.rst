..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_cluster_enable:

confluent schema-registry cluster enable
----------------------------------------

Description
~~~~~~~~~~~

Enable Schema Registry for this environment.

::

  confluent schema-registry cluster enable [flags]

Flags
~~~~~

::

      --cloud string         REQUIRED: Specify the cloud provider as "aws", "azure", or "gcp".
      --geo string           REQUIRED: Specify the geo as "us", "eu", or "apac".
      --package string       Specify the type of Stream Governance package as "essentials" or "advanced". (default "essentials")
      --context string       CLI context name.
      --environment string   Environment ID.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Enable Schema Registry, using Google Cloud Platform in the US with the "advanced" package.

::

  confluent schema-registry cluster enable --cloud gcp --geo us --package advanced

See Also
~~~~~~~~

* :ref:`confluent_schema-registry_cluster` - Manage Schema Registry cluster.
