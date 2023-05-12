..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_connect_cluster_list:

confluent connect cluster list
------------------------------

Description
~~~~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      List connectors.
      
      ::
      
        confluent connect cluster list [flags]
      
   .. group-tab:: On-Prem
   
      List Connect clusters that are registered with the MDS cluster registry.
      
      ::
      
        confluent connect cluster list [flags]
      
Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --cluster string       Kafka cluster ID.
            --context string       CLI context name.
            --environment string   Environment ID.
        -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")
      
   .. group-tab:: On-Prem
   
      ::
      
        -o, --output string    Specify the output format as "human", "json", or "yaml". (default "human")
            --context string   CLI context name.
      
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
   
      List connectors in the current or specified Kafka cluster context.
      
      ::
      
        confluent connect cluster list
        confluent connect cluster list --cluster lkc-123456
      
   .. group-tab:: On-Prem
   
      No examples.
      
See Also
~~~~~~~~

* :ref:`confluent_connect_cluster` - Manage Connect clusters.
