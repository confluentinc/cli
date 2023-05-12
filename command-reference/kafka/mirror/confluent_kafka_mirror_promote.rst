..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_mirror_promote:

confluent kafka mirror promote
------------------------------

Description
~~~~~~~~~~~

Promote mirror topics.

::

  confluent kafka mirror promote <destination-topic-1> [destination-topic-2] ... [destination-topic-N] [flags]

Flags
~~~~~

::

      --link string          REQUIRED: Name of cluster link.
      --dry-run              If set, does not actually create the link, but simply validates it.
      --cluster string       Kafka cluster ID.
      --context string       CLI context name.
      --environment string   Environment ID.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Promote mirror topics "my-topic-1" and "my-topic-2":

::

  confluent kafka mirror promote my-topic-1 my-topic-2 --link my-link

See Also
~~~~~~~~

* :ref:`confluent_kafka_mirror` - Manage cluster linking mirror topics.
