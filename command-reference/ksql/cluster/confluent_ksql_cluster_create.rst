..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_ksql_cluster_create:

confluent ksql cluster create
-----------------------------

Description
~~~~~~~~~~~

Create a ksqlDB cluster.

::

  confluent ksql cluster create <name> [flags]

Flags
~~~~~

::

      --credential-identity string   REQUIRED: User account ID or service account ID to be associated with this cluster. We will create an API key associated with this identity and use it to authenticate the ksqlDB cluster with Kafka.
      --csu int32                    Number of CSUs to use in the cluster. (default 4)
      --log-exclude-rows             Exclude row data in the processing log.
      --cluster string               Kafka cluster ID.
      --context string               CLI context name.
      --environment string           Environment ID.
  -o, --output string                Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_ksql_cluster` - Manage ksqlDB clusters.
