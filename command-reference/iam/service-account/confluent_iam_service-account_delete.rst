..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_service-account_delete:

confluent iam service-account delete
------------------------------------

Description
~~~~~~~~~~~

Delete a service account.

::

  confluent iam service-account delete <id> [flags]

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

Delete service account "sa-123456".

::

  confluent iam service-account delete sa-123456

See Also
~~~~~~~~

* :ref:`confluent_iam_service-account` - Manage service accounts.
