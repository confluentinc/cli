..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_audit-log_config_describe:

confluent audit-log config describe
-----------------------------------

Description
~~~~~~~~~~~

Prints the audit log configuration spec object, where "spec" refers to the JSON blob that describes audit log routing rules.

::

  confluent audit-log config describe [flags]

Flags
~~~~~

::

      --context string   CLI context name.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_audit-log_config` - Manage the audit log configuration specification.
