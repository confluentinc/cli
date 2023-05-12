..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_quota_delete:

confluent kafka quota delete
----------------------------

Description
~~~~~~~~~~~

Delete a Kafka client quota.

::

  confluent kafka quota delete <id> [flags]

Flags
~~~~~

::

      --force           Skip the deletion confirmation prompt.
  -o, --output string   Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_kafka_quota` - Manage Kafka client quotas.
