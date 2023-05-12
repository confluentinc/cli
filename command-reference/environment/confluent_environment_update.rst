..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_environment_update:

confluent environment update
----------------------------

Description
~~~~~~~~~~~

Update an existing Confluent Cloud environment.

::

  confluent environment update <id> [flags]

Flags
~~~~~

::

      --name string      REQUIRED: New name for Confluent Cloud environment.
      --context string   CLI context name.
  -o, --output string    Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_environment` - Manage and select Confluent Cloud environments.
