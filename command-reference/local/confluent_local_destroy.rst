..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_destroy:

confluent local destroy
-----------------------

.. include:: ../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

Delete an existing Confluent Platform run. All running services are stopped and the data and the log files of all services are deleted.

::

  confluent local destroy [flags]

.. include:: ../../includes/path-set-cli.rst

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

If you run the ``confluent local destroy`` command, your output will confirm that every service is stopped and the deleted filesystem path is printed:

::

  confluent local destroy

See Also
~~~~~~~~

* :ref:`confluent_local` - Manage a local Confluent Platform development environment.
