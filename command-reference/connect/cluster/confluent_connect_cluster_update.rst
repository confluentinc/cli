..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_connect_cluster_update:

confluent connect cluster update
--------------------------------

Description
~~~~~~~~~~~

Update a connector configuration.

::

  confluent connect cluster update <id> [flags]

Flags
~~~~~

::

      --config strings       A comma-separated list of configuration overrides ("key=value") for the connector being updated.
      --config-file string   JSON connector config file.
      --cluster string       Kafka cluster ID.
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

* :ref:`confluent_connect_cluster` - Manage Connect clusters.
