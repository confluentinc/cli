..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_cluster_update:

confluent kafka cluster update
------------------------------

Description
~~~~~~~~~~~

Update a Kafka cluster.

::

  confluent kafka cluster update <id> [flags]

Flags
~~~~~

::

      --name string          Name of the Kafka cluster.
      --cku uint32           Number of Confluent Kafka Units. For Kafka clusters of type "dedicated" only. When shrinking a cluster, you must reduce capacity one CKU at a time.
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

Change a cluster's name and expand its CKU count:

::

  confluent kafka cluster update lkc-abc123 --name "Cool Cluster" --cku 3

See Also
~~~~~~~~

* :ref:`confluent_kafka_cluster` - Manage Kafka clusters.
