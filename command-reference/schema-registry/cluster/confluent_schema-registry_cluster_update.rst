..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_cluster_update:

confluent schema-registry cluster update
----------------------------------------

Description
~~~~~~~~~~~

Update global mode or compatibility of Schema Registry.

::

  confluent schema-registry cluster update [flags]

Flags
~~~~~

::

      --compatibility string   Can be "backward", "backward_transitive", "forward", "forward_transitive", "full", "full_transitive", or "none".
      --mode string            Can be "readwrite", "readonly", or "import".
      --api-key string         API key.
      --api-secret string      API key secret.
      --context string         CLI context name.
      --environment string     Environment ID.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Update top-level compatibility of Schema Registry.

::

  confluent schema-registry cluster update --compatibility backward

Update top-level mode of Schema Registry.

::

  confluent schema-registry cluster update --mode readwrite

See Also
~~~~~~~~

* :ref:`confluent_schema-registry_cluster` - Manage Schema Registry cluster.
