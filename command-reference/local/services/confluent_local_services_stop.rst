..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_services_stop:

confluent local services stop
-----------------------------

.. include:: ../../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

Stop all Confluent Platform services.

::

  confluent local services stop [flags]

.. include:: ../../../includes/path-set-cli.rst

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Stop all running services:

::

  confluent local services stop

Stop Apache Kafka® and its dependent services.

::

  confluent local services kafka stop

See Also
~~~~~~~~

* :ref:`confluent_local_services` - Manage Confluent Platform services.
