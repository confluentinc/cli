..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_api-key_store:

confluent api-key store
-----------------------

Description
~~~~~~~~~~~

Use this command to register an API secret created by another
process and store it locally.

When you create an API key with the CLI, it is automatically stored locally.
However, when you create an API key using the UI, API, or with the CLI on another
machine, the secret is not available for CLI use until you "store" it. This is because
secrets are irretrievable after creation.

You must have an API secret stored locally for certain CLI commands to
work. For example, the Kafka topic consume and produce commands require an API secret.

::

  confluent api-key store [api-key] [secret] [flags]

Flags
~~~~~

::

      --resource string      The resource ID of the resource the API key is for.
  -f, --force                Force overwrite existing secret for this key.
      --context string       CLI context name.
      --environment string   Environment ID.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Pass the API key and secret as arguments

::

  confluent api-key store my-key my-secret

Get prompted for only the API secret

::

  confluent api-key store my-key

Get prompted for both the API key and secret

::

  confluent api-key store

See Also
~~~~~~~~

* :ref:`confluent_api-key` - Manage API keys.
