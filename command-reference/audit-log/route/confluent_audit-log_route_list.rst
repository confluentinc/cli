..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_audit-log_route_list:

confluent audit-log route list
------------------------------

Description
~~~~~~~~~~~

List the routes that match either the queried resource or its sub-resources.

::

  confluent audit-log route list [flags]

Flags
~~~~~

::

      --resource string   REQUIRED: The Confluent resource name (CRN) that is the subject of the query.
      --context string    CLI context name.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_audit-log_route` - Return the audit log route rules.
