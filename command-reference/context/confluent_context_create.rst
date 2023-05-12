..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_context_create:

confluent context create
------------------------

Description
~~~~~~~~~~~

Create a new Cloud context with an API key.

::

  confluent context create [context] [flags]

Flags
~~~~~

::

      --bootstrap string    REQUIRED: Bootstrap URL.
      --api-key string      REQUIRED: API key.
      --api-secret string   REQUIRED: API secret. Can be specified as plaintext, as a file, starting with '@', or as stdin, starting with '-'.
  -o, --output string       Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Create a context called "new context":

::

  confluent context create "new context" --bootstrap https://example.com --api-key key --api-secret @api-secret.txt

See Also
~~~~~~~~

* :ref:`confluent_context` - Manage CLI configuration contexts.
