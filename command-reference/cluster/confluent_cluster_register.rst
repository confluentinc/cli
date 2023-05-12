..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_cluster_register:

confluent cluster register
--------------------------

Description
~~~~~~~~~~~

Register cluster with the MDS cluster registry.

::

  confluent cluster register [flags]

Flags
~~~~~

::

      --hosts strings                    REQUIRED: A comma-separated list of hosts.
      --protocol string                  REQUIRED: Security protocol.
      --cluster-name string              REQUIRED: Cluster name.
      --kafka-cluster string             REQUIRED: Kafka cluster ID.
      --schema-registry-cluster string   Schema Registry cluster ID.
      --ksql-cluster string              ksqlDB cluster ID.
      --connect-cluster string           Kafka Connect cluster ID.
      --context string                   CLI context name.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_cluster` - Retrieve metadata about Confluent Platform clusters.
