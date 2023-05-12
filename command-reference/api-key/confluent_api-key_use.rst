..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_api-key_use:

confluent api-key use
---------------------

Description
~~~~~~~~~~~

Set the active API key for use in any command which supports passing an API key with the ``--api-key`` flag.

::

  confluent api-key use <api-key> [flags]

Flags
~~~~~

::

      --resource string   REQUIRED: The resource ID.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_api-key` - Manage API keys.
