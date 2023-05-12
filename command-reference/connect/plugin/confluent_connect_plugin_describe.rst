..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_connect_plugin_describe:

confluent connect plugin describe
---------------------------------

Description
~~~~~~~~~~~

Describe a connector plugin.

::

  confluent connect plugin describe <plugin> [flags]

Flags
~~~~~

::

      --cluster string       Kafka cluster ID.
      --context string       CLI context name.
      --environment string   Environment ID.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Describe the required connector configuration parameters for connector plugin "MySource".

::

  confluent connect plugin describe MySource

See Also
~~~~~~~~

* :ref:`confluent_connect_plugin` - Show plugins and their configurations.
