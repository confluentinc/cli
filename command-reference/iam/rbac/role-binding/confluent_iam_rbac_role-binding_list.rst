..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_rbac_role-binding_list:

confluent iam rbac role-binding list
------------------------------------

Description
~~~~~~~~~~~

List the role bindings for a particular principal and/or role, and a particular scope.

::

  confluent iam rbac role-binding list [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --principal string                 Principal whose role bindings should be listed.
            --current-user                     Show role bindings belonging to the current user.
            --role string                      List role bindings under a specific role given to a principal. Or if no principal is specified, list principals with the role.
            --environment string               Environment ID for scope of role binding listings.
            --current-environment              Use current environment ID for scope.
            --cloud-cluster string             Cloud cluster ID for scope of role binding listings.
            --kafka-cluster string             Kafka cluster ID for scope of role binding listings.
            --schema-registry-cluster string   Schema Registry cluster ID for the role binding listings.
            --ksql-cluster string              ksqlDB cluster name for the role binding listings.
            --resource string                  If specified with a role and no principals, list principals with role bindings to the role for this qualified resource.
            --inclusive                        List all role bindings in a specific scope and its nested scopes.
        -o, --output string                    Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --principal string                 Principal whose role bindings should be listed.
            --current-user                     Show role bindings belonging to the current user.
            --role string                      List role bindings under a specific role given to a principal. Or if no principal is specified, list principals with the role.
            --kafka-cluster string             Kafka cluster ID for scope of role binding listings.
            --schema-registry-cluster string   Schema Registry cluster ID for scope of role binding listings.
            --ksql-cluster string              ksqlDB cluster ID for scope of role binding listings.
            --connect-cluster string           Kafka Connect cluster ID for scope of role binding listings.
            --cluster-name string              Cluster name to uniquely identify the cluster for role binding listings.
            --context string                   CLI context name.
            --resource string                  If specified with a role and no principals, list principals with role bindings to the role for this qualified resource.
            --inclusive                        List all role bindings in a specific scope and its nested scopes.
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
   
      List the role bindings for the current user:
      
      ::
      
        confluent iam rbac role-binding list --current-user
      
      List the role bindings for user "u-123456":
      
      ::
      
        confluent iam rbac role-binding list --principal User:u-123456
      
      List the role bindings for principals with role "CloudClusterAdmin":
      
      ::
      
        confluent iam rbac role-binding list --role CloudClusterAdmin --current-environment --cloud-cluster lkc-123456
      
      List the role bindings for user "u-123456" with role "CloudClusterAdmin":
      
      ::
      
        confluent iam rbac role-binding list --principal User:u-123456 --role CloudClusterAdmin --environment env-12345 --cloud-cluster lkc-123456
      
      List the role bindings for user "u-123456" for all scopes:
      
      ::
      
        confluent iam rbac role-binding list --principal User:u-123456 --inclusive
      
      List the role bindings for the current user at the environment scope and its nested scopes:
      
      ::
      
        confluent iam rbac role-binding list --current-user --environment env-12345 --inclusive
      
   .. group-tab:: On-Prem
   
      Only use the ``--resource`` flag when specifying a ``--role`` with no ``--principal`` specified. If specifying a ``--principal``, then the ``--resource`` flag is ignored. To list role bindings for a specific role on an identified resource:
      
      ::
      
        confluent iam rbac role-binding list --kafka-cluster $KAFKA_CLUSTER_ID --role DeveloperRead --resource Topic
      
      List the role bindings for a specific principal:
      
      ::
      
        confluent iam rbac role-binding list --kafka-cluster $KAFKA_CLUSTER_ID --principal User:my-user
      
      List the role bindings for a specific principal, filtered to a specific role:
      
      ::
      
        confluent iam rbac role-binding list --kafka-cluster $KAFKA_CLUSTER_ID --principal User:my-user --role DeveloperRead
      
      List the principals bound to a specific role:
      
      ::
      
        confluent iam rbac role-binding list --kafka-cluster $KAFKA_CLUSTER_ID --role DeveloperWrite
      
      List the principals bound to a specific resource with a specific role:
      
      ::
      
        confluent iam rbac role-binding list --kafka-cluster $KAFKA_CLUSTER_ID --role DeveloperWrite --resource Topic:my-topic
      
See Also
~~~~~~~~

* :ref:`confluent_iam_rbac_role-binding` - Manage RBAC and IAM role bindings.
