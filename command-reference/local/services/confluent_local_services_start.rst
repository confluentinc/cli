..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_services_start:

confluent local services start
------------------------------

.. include:: ../../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

Start all Confluent Platform services.

::

  confluent local services start [flags]

.. include:: ../../../includes/path-set-cli.rst

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Start all available services:

::

  confluent local services start

Start Apache Kafka® and ZooKeeper as its dependency:

::

  confluent local services kafka start

See Also
~~~~~~~~

* :ref:`confluent_local_services` - Manage Confluent Platform services.
