..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_ksql_cluster_delete:

confluent ksql cluster delete
-----------------------------

Description
~~~~~~~~~~~

Delete a ksqlDB cluster.

::

  confluent ksql cluster delete <id> [flags]

Flags
~~~~~

::

      --force                Skip the deletion confirmation prompt.
      --context string       CLI context name.
      --environment string   Environment ID.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_ksql_cluster` - Manage ksqlDB clusters.
