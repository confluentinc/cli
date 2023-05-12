..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_services_connect_connector_unload:

confluent local services connect connector unload
-------------------------------------------------

.. include:: ../../../../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

Unload a connector.

::

  confluent local services connect connector unload <connector-name> [flags]

.. include:: ../../../../../includes/path-set-cli.rst

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Unload a predefined connector called ``s3-sink``:

::

  confluent local unload s3-sink

See Also
~~~~~~~~

* :ref:`confluent_local_services_connect_connector` - Manage connectors.
