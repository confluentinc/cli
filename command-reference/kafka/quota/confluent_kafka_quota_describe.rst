..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_quota_describe:

confluent kafka quota describe
------------------------------

Description
~~~~~~~~~~~

Describe a Kafka client quota.

::

  confluent kafka quota describe <id> [flags]

Flags
~~~~~

::

      --cluster string       Kafka cluster ID.
      --environment string   Environment ID.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_kafka_quota` - Manage Kafka client quotas.
