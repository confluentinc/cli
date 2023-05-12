..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_byok_list:

confluent byok list
-------------------

Description
~~~~~~~~~~~

List self-managed keys registered in Confluent Cloud.

::

  confluent byok list [flags]

Flags
~~~~~

::

      --provider string   Specify the provider as "aws" or "azure".
      --state string      Specify the state as "in-use" or "available".
  -o, --output string     Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_byok` - Manage your keys in Confluent Cloud.
