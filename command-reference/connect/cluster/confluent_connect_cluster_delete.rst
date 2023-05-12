..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_connect_cluster_delete:

confluent connect cluster delete
--------------------------------

Description
~~~~~~~~~~~

Delete a connector.

::

  confluent connect cluster delete <id> [flags]

Flags
~~~~~

::

      --force                Skip the deletion confirmation prompt.
      --cluster string       Kafka cluster ID.
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

Delete a connector in the current or specified Kafka cluster context.

::

  confluent connect cluster delete
  confluent connect cluster delete --cluster lkc-123456

See Also
~~~~~~~~

* :ref:`confluent_connect_cluster` - Manage Connect clusters.
