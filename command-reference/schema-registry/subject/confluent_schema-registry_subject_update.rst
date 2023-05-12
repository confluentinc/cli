..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_subject_update:

confluent schema-registry subject update
----------------------------------------

Description
~~~~~~~~~~~

Update subject compatibility or mode.

::

  confluent schema-registry subject update <subject> [flags]

Flags
~~~~~

.. tabs::

   .. group-tab:: Cloud
   
      ::
      
            --compatibility string   Can be "backward", "backward_transitive", "forward", "forward_transitive", "full", "full_transitive", or "none".
            --mode string            Can be "readwrite", "readonly", or "import".
            --api-key string         API key.
            --api-secret string      API key secret.
            --context string         CLI context name.
            --environment string     Environment ID.
      
   .. group-tab:: On-Prem
   
      ::
      
            --compatibility string              Can be "backward", "backward_transitive", "forward", "forward_transitive", "full", "full_transitive", or "none".
            --mode string                       Can be "readwrite", "readonly", or "import".
            --ca-location string                File or directory path to CA certificate(s) to authenticate the Schema Registry client.
            --schema-registry-endpoint string   The URL of the Schema Registry cluster.
            --context string                    CLI context name.
      
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
   
      Update subject-level compatibility of subject "payments".
      
      ::
      
        confluent schema-registry subject update payments --compatibility backward
      
      Update subject-level mode of subject "payments".
      
      ::
      
        confluent schema-registry subject update payments --mode readwrite
      
   .. group-tab:: On-Prem
   
      Update subject-level compatibility of subject "payments".
      
      ::
      
        confluent schema-registry subject update payments --compatibility backward --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
      Update subject-level mode of subject "payments".
      
      ::
      
        confluent schema-registry subject update payments --mode readwrite --ca-location <ca-file-location> --schema-registry-endpoint <schema-registry-endpoint>
      
See Also
~~~~~~~~

* :ref:`confluent_schema-registry_subject` - Manage Schema Registry subjects.
