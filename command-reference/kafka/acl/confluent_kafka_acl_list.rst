..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_acl_list:

confluent kafka acl list
------------------------

Description
~~~~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      List Kafka ACLs for a resource.
      
      ::
      
        confluent kafka acl list [flags]
      
   .. group-tab:: On-Prem
   
      List Kafka ACLs.
      
      ::
      
        confluent kafka acl list [flags]
      
Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --cluster-scope             Modify ACLs for the cluster.
            --topic string              Modify ACLs for the specified topic resource.
            --consumer-group string     Modify ACLs for the specified consumer group resource.
            --transactional-id string   Modify ACLs for the specified TransactionalID resource.
            --prefix                    When this flag is set, the specified resource name is interpreted as a prefix.
            --cluster string            Kafka cluster ID.
            --context string            CLI context name.
            --environment string        Environment ID.
            --service-account string    Service account ID.
            --principal string          Principal for this operation, prefixed with "User:".
        -o, --output string             Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
            --ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
            --client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
            --client-key-path string    Path to client private key, include for mTLS authentication.
            --no-authentication         Include if requests should be made without authentication headers, and user will not be prompted for credentials.
            --prompt                    Bypass use of available login credentials and prompt for Kafka Rest credentials.
            --principal string          Principal for this operation with User: or Group: prefix.
            --operation string          Set ACL Operation to: (all, alter, alter-configs, cluster-action, create, delete, describe, describe-configs, idempotent-write, read, write).
            --host string               Set host for access. Only IP addresses are supported. (default "*")
            --allow                     ACL permission to allow access.
            --deny                      ACL permission to restrict access to resource.
            --cluster-scope             Set the cluster resource. With this option the ACL grants access to the provided operations on the Kafka cluster itself.
            --consumer-group string     Set the Consumer Group resource.
            --transactional-id string   Set the TransactionalID resource.
            --topic string              Set the topic resource. With this option the ACL grants the provided operations on the topics that start with that prefix, depending on whether the --prefix option was also passed.
            --prefix                    Set to match all resource names prefixed with this value.
            --context string            CLI context name.
        -o, --output string             Specify the output format as "human", "json", or "yaml". (default "human")
      
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
   
      No examples.
      
   .. group-tab:: On-Prem
   
      List all the local ACLs for the Kafka cluster (providing embedded Kafka REST Proxy endpoint).
      
      ::
      
        confluent kafka acl list --url http://localhost:8090/kafka
      
      List all the local ACLs for the Kafka cluster (providing Kafka REST Proxy endpoint).
      
      ::
      
        confluent kafka acl list --url http://localhost:8082
      
      List all the ACLs for the Kafka cluster that include allow permissions for the user "Jane":
      
      ::
      
        confluent kafka acl list --allow --principal User:Jane
      
See Also
~~~~~~~~

* :ref:`confluent_kafka_acl` - Manage Kafka ACLs.
