..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_client-config_create_python:

confluent kafka client-config create python
-------------------------------------------

Description
~~~~~~~~~~~

Create a Python client configuration file, of which the client configuration file is printed to stdout and the warnings are printed to stderr. Please see our examples on how to redirect the command output.

::

  confluent kafka client-config create python [flags]

Flags
~~~~~

::

      --context string                      CLI context name.
      --environment string                  Environment ID.
      --cluster string                      Kafka cluster ID.
      --api-key string                      API key.
      --api-secret string                   API key secret.
      --schema-registry-api-key string      Schema registry API key.
      --schema-registry-api-secret string   Schema registry API key secret.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Create a Python client configuration file.

::

  confluent kafka client-config create python --schema-registry-api-key my-sr-key --schema-registry-api-secret my-sr-secret

Create a Python client configuration file with arguments passed via flags.

::

  confluent kafka client-config create python --environment env-123 --cluster lkc-123456 --api-key my-key --api-secret my-secret --schema-registry-api-key my-sr-key --schema-registry-api-secret my-sr-secret

Create a Python client configuration file, redirecting the configuration to a file and the warnings to a separate file.

::

  confluent kafka client-config create python --schema-registry-api-key my-sr-key --schema-registry-api-secret my-sr-secret 1> my-client-config-file.config 2> my-warnings-file

Create a Python client configuration file, redirecting the configuration to a file and keeping the warnings in the console.

::

  confluent kafka client-config create python --schema-registry-api-key my-sr-key --schema-registry-api-secret my-sr-secret 1> my-client-config-file.config 2>&1

See Also
~~~~~~~~

* :ref:`confluent_kafka_client-config_create` - Create a Kafka client configuration file.
