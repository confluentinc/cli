..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_audit-log_config_edit:

confluent audit-log config edit
-------------------------------

Description
~~~~~~~~~~~

Edit the audit-log config spec object interactively, using the $EDITOR specified in your environment (for example, vim).

::

  confluent audit-log config edit [flags]

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
