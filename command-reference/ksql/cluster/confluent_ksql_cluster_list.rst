..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_ksql_cluster_list:

confluent ksql cluster list
---------------------------

Description
~~~~~~~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      List ksqlDB clusters.
      
      ::
      
        confluent ksql cluster list [flags]
      
   .. group-tab:: On-Prem
   
      List ksqlDB clusters that are registered with the MDS cluster registry.
      
      ::
      
        confluent ksql cluster list [flags]
      
Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
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

* :ref:`confluent_ksql_cluster` - Manage ksqlDB clusters.
