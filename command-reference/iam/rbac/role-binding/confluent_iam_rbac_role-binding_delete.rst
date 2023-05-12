..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_rbac_role-binding_delete:

confluent iam rbac role-binding delete
--------------------------------------

Description
~~~~~~~~~~~

Delete a role binding.

::

  confluent iam rbac role-binding delete [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --role string                      REQUIRED: Role name of the existing role binding.
            --principal string                 REQUIRED: Qualified principal name associated with the role binding.
            --force                            Skip the deletion confirmation prompt.
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
      
            --role string                      REQUIRED: Role name of the existing role binding.
            --principal string                 REQUIRED: Qualified principal name associated with the role binding.
            --force                            Skip the deletion confirmation prompt.
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
   
      Delete the role "ResourceOwner" for the resource "Topic:my-topic" on the Kafka cluster "lkc-123456":
      
      ::
      
        confluent iam rbac role-binding delete --principal User:u-123456 --role ResourceOwner --environment env-12345 --kafka-cluster lkc-123456 --resource Topic:my-topic
      
   .. group-tab:: On-Prem
   
      No examples.
      
See Also
~~~~~~~~

* :ref:`confluent_iam_rbac_role-binding` - Manage RBAC and IAM role bindings.
