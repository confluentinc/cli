..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_asyncapi_import:

confluent asyncapi import
-------------------------

Description
~~~~~~~~~~~

Update a Kafka cluster and Schema Registry according to an AsyncAPI specification file.

::

  confluent asyncapi import [flags]

Flags
~~~~~

::

      --file string                         REQUIRED: Input file name.
      --overwrite                           Overwrite existing topics with the same name.
      --kafka-api-key string                Kafka cluster API key.
      --schema-registry-api-key string      API key for Schema Registry.
      --schema-registry-api-secret string   API secret for Schema Registry.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Import an AsyncAPI specification file to populate an existing Kafka cluster and Schema Registry.

::

  confluent asyncapi import --file spec.yaml

See Also
~~~~~~~~

* :ref:`confluent_asyncapi` - Manage AsyncAPI document tooling.
