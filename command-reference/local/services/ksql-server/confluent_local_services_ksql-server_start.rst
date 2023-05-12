..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_services_ksql-server_start:

confluent local services ksql-server start
------------------------------------------

.. include:: ../../../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

Start ksqlDB Server.

::

  confluent local services ksql-server start [flags]

.. include:: ../../../../includes/path-set-cli.rst

Flags
~~~~~

::

  -c, --config string   Configure ksqlDB Server with a specific properties file.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_local_services_ksql-server` - Manage ksqlDB Server.
