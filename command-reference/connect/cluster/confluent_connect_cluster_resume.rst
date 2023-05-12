..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_connect_cluster_resume:

confluent connect cluster resume
--------------------------------

Description
~~~~~~~~~~~

Resume connectors.

::

  confluent connect cluster resume <id-1> [id-2] ... [id-N] [flags]

Flags
~~~~~

::

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

Resume connectors "lcc-000001" and "lcc-000002":

::

  confluent connect cluster resume lcc-000001 lcc-000002

See Also
~~~~~~~~

* :ref:`confluent_connect_cluster` - Manage Connect clusters.
