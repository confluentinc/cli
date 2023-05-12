..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_connect_cluster_describe:

confluent connect cluster describe
----------------------------------

Description
~~~~~~~~~~~

Describe a connector.

::

  confluent connect cluster describe <id> [flags]

Flags
~~~~~

::

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

Describe connector and task level details of a connector in the current or specified Kafka cluster context.

::

  confluent connect cluster describe lcc-123456
  confluent connect cluster describe lcc-123456 --cluster lkc-123456

See Also
~~~~~~~~

* :ref:`confluent_connect_cluster` - Manage Connect clusters.
