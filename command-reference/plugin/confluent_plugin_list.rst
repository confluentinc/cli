..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_plugin_list:

confluent plugin list
---------------------

Description
~~~~~~~~~~~

List Confluent CLI plugins in $PATH. Plugins are executable files that begin with "confluent-".

::

  confluent plugin list [flags]

Flags
~~~~~

::

  -o, --output string   Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_plugin` - Manage Confluent plugins.
