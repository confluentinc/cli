..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_context_describe:

confluent context describe
--------------------------

Description
~~~~~~~~~~~

Describe a context or a specific context field.

::

  confluent context describe [context] [flags]

Flags
~~~~~

::

      --api-key         Get the API key for a context.
      --username        Get the username for a context.
  -o, --output string   Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_context` - Manage CLI configuration contexts.
