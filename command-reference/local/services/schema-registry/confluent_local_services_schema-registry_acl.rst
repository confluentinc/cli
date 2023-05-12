..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_local_services_schema-registry_acl:

confluent local services schema-registry acl
--------------------------------------------

.. include:: ../../../../includes/cli.rst
  :end-before: cli_limitations_end
  :start-after: cli_limitations_start

Description
~~~~~~~~~~~

Specify an ACL for Schema Registry.

::

  confluent local services schema-registry acl [flags]

.. include:: ../../../../includes/path-set-cli.rst

Flags
~~~~~

::

      --add                Indicates you are trying to add ACLs.
      --list               List all the current ACLs.
      --remove             Indicates you are trying to remove ACLs.
  -o, --operation string   Operation that is being authorized. Valid operation names are SUBJECT_READ, SUBJECT_WRITE, SUBJECT_DELETE, SUBJECT_COMPATIBILITY_READ, SUBJECT_COMPATIBILITY_WRITE, GLOBAL_COMPATIBILITY_READ, GLOBAL_COMPATIBILITY_WRITE, and GLOBAL_SUBJECTS_READ.
  -p, --principal string   Principal to which the ACL is being applied to. Use * to apply to all principals.
  -s, --subject string     Subject to which the ACL is being applied to. Only applicable for SUBJECT operations. Use * to apply to all subjects.
  -t, --topic string       Topic to which the ACL is being applied to. The corresponding subjects would be topic-key and topic-value. Only applicable for SUBJECT operations. Use * to apply to all subjects.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_local_services_schema-registry` - Manage Schema Registry.
