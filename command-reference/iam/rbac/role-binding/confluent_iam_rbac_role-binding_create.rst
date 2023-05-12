..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_rbac_role-binding_create:

confluent iam rbac role-binding create
--------------------------------------

Description
~~~~~~~~~~~

Create a role binding.

::

  confluent iam rbac role-binding create [flags]

.. note:: If you need to troubleshoot when setting up role bindings, it may be helpful to view audit logs on the fly to identify authorization events for specific principals, resources, or operations. For details, refer to :platform:`Viewing audit logs on the fly|security/audit-logs/audit-logs-properties-config.html#view-audit-logs-on-the-fly`.

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --role string                      REQUIRED: Role name of the new role binding.
            --principal string                 REQUIRED: Qualified principal name for the role binding.
            --environment string               Environment ID for scope of role-binding operation.
            --current-environment              Use current environment ID for scope.
            --cloud-cluster string             Cloud cluster ID for the role binding.
            --kafka-cluster string             Kafka cluster ID for the role binding.
            --schema-registry-cluster string   Schema Registry cluster ID for the role binding.
            --ksql-cluster string              ksqlDB cluster name for the role binding.
            --resource string                  Qualified resource name for the role binding.
            --prefix                           Whether the provided resource name is treated as a prefix pattern.
        -o, --output string                    Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --role string                      REQUIRED: Role name of the new role binding.
            --principal string                 REQUIRED: Qualified principal name for the role binding.
            --kafka-cluster string             Kafka cluster ID for the role binding.
            --schema-registry-cluster string   Schema Registry cluster ID for the role binding.
            --ksql-cluster string              ksqlDB cluster ID for the role binding.
            --connect-cluster string           Kafka Connect cluster ID for the role binding.
            --cluster-name string              Cluster name to uniquely identify the cluster for role binding listings.
            --context string                   CLI context name.
            --resource string                  Qualified resource name for the role binding.
            --prefix                           Whether the provided resource name is treated as a prefix pattern.
        -o, --output string                    Specify the output format as "human", "json", or "yaml". (default "human")
      
Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      Grant the role "CloudClusterAdmin" to the principal "User:u-123456" in the environment "env-12345" for the cloud cluster "lkc-123456":
      
      ::
      
        confluent iam rbac role-binding create --principal User:u-123456 --role CloudClusterAdmin --environment env-12345 --cloud-cluster lkc-123456
      
      Grant the role "ResourceOwner" to the principal "User:u-123456", in the environment "env-12345" for the Kafka cluster "lkc-123456" on the resource "Topic:my-topic":
      
      ::
      
        confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --resource Topic:my-topic --environment env-12345 --kafka-cluster lkc-123456
      
      Grant the role "MetricsViewer" to service account "sa-123456":
      
      ::
      
        confluent iam rbac role-binding create --principal User:sa-123456 --role MetricsViewer
      
      Grant the "ResourceOwner" role to principal "User:u-123456" and all subjects for Schema Registry cluster "lsrc-123456" in environment "env-12345":
      
      ::
      
        confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-12345 --schema-registry-cluster lsrc-123456 --resource "Subject:*"
      
      Grant the "ResourceOwner" role to principal "User:u-123456" and subject "test" for the Schema Registry cluster "lsrc-123456" in the environment "env-12345":
      
      ::
      
        confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-12345 --schema-registry-cluster lsrc-123456 --resource "Subject:test"
      
      Grant the "ResourceOwner" role to principal "User:u-123456" and all subjects in schema context "schema_context" for Schema Registry cluster "lsrc-123456" in the environment "env-12345":
      
      ::
      
        confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-12345 --schema-registry-cluster lsrc-123456 --resource "Subject::.schema_context:*"
      
      Grant the "ResourceOwner" role to principal "User:u-123456" and subject "test" in schema context "schema_context" for Schema Registry "lsrc-123456" in the environment "env-12345":
      
      ::
      
        confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-12345 --schema-registry-cluster lsrc-123456 --resource "Subject::.schema_context:test"
      
   .. group-tab:: On-Prem
   
      Create a role binding for the principal permitting it produce to topic "my-topic":
      
      ::
      
        confluent iam rbac role-binding create --principal User:appSA --role DeveloperWrite --resource Topic:my-topic --kafka-cluster $KAFKA_CLUSTER_ID
      
See Also
~~~~~~~~

* :ref:`confluent_iam_rbac_role-binding` - Manage RBAC and IAM role bindings.
