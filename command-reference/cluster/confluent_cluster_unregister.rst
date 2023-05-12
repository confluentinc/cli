..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_cluster_unregister:

confluent cluster unregister
----------------------------

Description
~~~~~~~~~~~

Unregister cluster from the MDS cluster registry.

::

  confluent cluster unregister [flags]

Flags
~~~~~

::

      --cluster-name string   REQUIRED: Cluster Name.
      --context string        CLI context name.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_cluster` - Retrieve metadata about Confluent Platform clusters.
