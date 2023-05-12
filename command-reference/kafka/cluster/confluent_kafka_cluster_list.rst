..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_cluster_list:

confluent kafka cluster list
----------------------------

Description
~~~~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      List Kafka clusters.
      
      ::
      
        confluent kafka cluster list [flags]
      
   .. group-tab:: On-Prem
   
      List Kafka clusters that are registered with the MDS cluster registry.
      
      ::
      
        confluent kafka cluster list [flags]
      
Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --all                  List clusters across all environments.
            --context string       CLI context name.
            --environment string   Environment ID.
        -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
            --context string   CLI context name.
        -o, --output string    Specify the output format as "human", "json", or "yaml". (default "human")
      
Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_kafka_cluster` - Manage Kafka clusters.
