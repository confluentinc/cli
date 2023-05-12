..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_partition_reassignment_list:

confluent kafka partition reassignment list
-------------------------------------------

Description
~~~~~~~~~~~

List ongoing partition reassignments for a given cluster, topic, or partition via Confluent Kafka REST.

::

  confluent kafka partition reassignment list [id] [flags]

Flags
~~~~~

::

      --topic string              Topic name to search by.
      --url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
      --ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
      --client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
      --client-key-path string    Path to client private key, include for mTLS authentication.
      --no-authentication         Include if requests should be made without authentication headers, and user will not be prompted for credentials.
      --prompt                    Bypass use of available login credentials and prompt for Kafka Rest credentials.
  -o, --output string             Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

List all partition reassignments for the Kafka cluster.

::

  confluent kafka partition reassignment list

List partition reassignments for topic "my_topic".

::

  confluent kafka partition reassignment list --topic my_topic

List partition reassignments for partition "1" of topic "my_topic".

::

  confluent kafka partition reassignment list 1 --topic my_topic

See Also
~~~~~~~~

* :ref:`confluent_kafka_partition_reassignment` - Manage ongoing partition reassignments.
