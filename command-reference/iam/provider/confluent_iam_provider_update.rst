..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_provider_update:

confluent iam provider update
-----------------------------

Description
~~~~~~~~~~~

Update an identity provider.

::

  confluent iam provider update <id> [flags]

Flags
~~~~~

::

      --description string   Description of the identity provider.
      --name string          Name of the identity provider.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Update the description of identity provider "op-123456".

::

  confluent iam provider update op-123456 --description "Update demo identity provider information."

See Also
~~~~~~~~

* :ref:`confluent_iam_provider` - Manage identity providers.
