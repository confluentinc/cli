..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_services_connect_connector_config:

confluent local services connect connector config
-------------------------------------------------

.. include:: ../../../../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

View or set connector configurations.

::

  confluent local services connect connector config <connector-name> [flags]

.. include:: ../../../../../includes/path-set-cli.rst

Flags
~~~~~

::

  -c, --config string   Configuration file for a connector.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Print the current configuration of a connector named ``s3-sink``:

::

  confluent local services connect connector config s3-sink

Configure a connector named ``wikipedia-file-source`` by passing its configuration properties in JSON format.

::

  confluent local services connect connector config wikipedia-file-source --config <path-to-connector>/wikipedia-file-source.json

Configure a connector named ``wikipedia-file-source`` by passing its configuration properties as Java properties.

::

  confluent local services connect connector config wikipedia-file-source --config <path-to-connector>/wikipedia-file-source.properties

See Also
~~~~~~~~

* :ref:`confluent_local_services_connect_connector` - Manage connectors.
