..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_rbac_role_describe:

confluent iam rbac role describe
--------------------------------

Description
~~~~~~~~~~~

Describe the resources and operations allowed for a role.

::

  confluent iam rbac role describe <name> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
        -o, --output string   Specify the output format as "human", "json", or "yaml". (default "human")
      
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

See Also
~~~~~~~~

* :ref:`confluent_iam_rbac_role` - Manage RBAC and IAM roles.
