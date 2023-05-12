..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_services_kafka-rest_log:

confluent local services kafka-rest log
---------------------------------------

.. include:: ../../../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

Print logs showing Kafka REST output.

::

  confluent local services kafka-rest log [flags]

.. include:: ../../../../includes/path-set-cli.rst

Flags
~~~~~

::

  -f, --follow   Log additional output until the command is interrupted.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_local_services_kafka-rest` - Manage Kafka REST.
