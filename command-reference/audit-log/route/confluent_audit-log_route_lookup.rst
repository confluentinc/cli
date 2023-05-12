..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_audit-log_route_lookup:

confluent audit-log route lookup
--------------------------------

Description
~~~~~~~~~~~

Return the single route that describes how audit log messages using this CRN would be routed, with all defaults populated.

::

  confluent audit-log route lookup <crn> [flags]

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

* :ref:`confluent_audit-log_route` - Return the audit log route rules.
