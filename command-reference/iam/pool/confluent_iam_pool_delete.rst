..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_pool_delete:

confluent iam pool delete
-------------------------

Description
~~~~~~~~~~~

Delete an identity pool.

::

  confluent iam pool delete <id> [flags]

Flags
~~~~~

::

      --provider string   REQUIRED: ID of this pool's identity provider.
      --force             Skip the deletion confirmation prompt.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Delete identity pool "pool-12345":

::

  confluent iam pool delete pool-12345 --provider op-12345

See Also
~~~~~~~~

* :ref:`confluent_iam_pool` - Manage identity pools.
