..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_provider_delete:

confluent iam provider delete
-----------------------------

Description
~~~~~~~~~~~

Delete an identity provider.

::

  confluent iam provider delete <id> [flags]

Flags
~~~~~

::

      --force   Skip the deletion confirmation prompt.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Delete identity provider "op-12345":

::

  confluent iam provider delete op-12345

See Also
~~~~~~~~

* :ref:`confluent_iam_provider` - Manage identity providers.
