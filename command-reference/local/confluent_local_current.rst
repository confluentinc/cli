..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_current:

confluent local current
-----------------------

.. include:: ../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

Print the filesystem path of the data and logs of the services managed by the current ``confluent local`` command. If such a path does not exist, it will be created.

::

  confluent local current [flags]

.. include:: ../../includes/path-set-cli.rst

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

In Linux, running ``confluent local current`` should resemble the following:

::

  /tmp/confluent.SpBP4fQi

In macOS, running ``confluent local current`` should resemble the following:

::

  /var/folders/cs/1rndf6593qb3kb6r89h50vgr0000gp/T/confluent.000000

See Also
~~~~~~~~

* :ref:`confluent_local` - Manage a local Confluent Platform development environment.
