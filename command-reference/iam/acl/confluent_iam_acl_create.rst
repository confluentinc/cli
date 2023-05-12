..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_acl_create:

confluent iam acl create
------------------------

Description
~~~~~~~~~~~

Create a centralized ACL.

::

  confluent iam acl create [flags]

Flags
~~~~~

::

      --kafka-cluster string      REQUIRED: Kafka cluster ID for scope of ACL commands.
      --principal string          REQUIRED: Principal for this operation with User: or Group: prefix.
      --operation string          REQUIRED: Set ACL Operation to: (all, alter, alter-configs, cluster-action, create, delete, describe, describe-configs, idempotent-write, read, write).
      --host string               Set host for access. Only IP addresses are supported. (default "*")
      --allow                     ACL permission to allow access.
      --deny                      ACL permission to restrict access to resource.
      --cluster-scope             Set the cluster resource. With this option the ACL grants
                                  access to the provided operations on the Kafka cluster itself.
      --consumer-group string     Set the Consumer Group resource.
      --transactional-id string   Set the TransactionalID resource.
      --topic string              Set the topic resource. With this option the ACL grants the provided
                                  operations on the topics that start with that prefix, depending on whether
                                  the --prefix option was also passed.
      --prefix                    Set to match all resource names prefixed with this value.
      --context string            CLI context name.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Create an ACL that grants the specified user "read" permission to the specified consumer group in the specified Kafka cluster:

::

  confluent iam acl create --allow --principal User:User1 --operation read --consumer-group java_example_group_1 --kafka-cluster <kafka-cluster-id>

Create an ACL that grants the specified user "write" permission on all topics in the specified Kafka cluster:

::

  confluent iam acl create --allow --principal User:User1 --operation write --topic "*" --kafka-cluster <kafka-cluster-id>

Create an ACL that assigns a group "read" access to all topics that use the specified prefix in the specified Kafka cluster:

::

  confluent iam acl create --allow --principal Group:Finance --operation read --topic financial --prefix --kafka-cluster <kafka-cluster-id>

See Also
~~~~~~~~

* :ref:`confluent_iam_acl` - Manage centralized ACLs.
