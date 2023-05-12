..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_cluster_delete:

confluent schema-registry cluster delete
----------------------------------------

Description
~~~~~~~~~~~

Delete the Schema Registry cluster for this environment.

::

  confluent schema-registry cluster delete [flags]

Flags
~~~~~

::

      --environment string   REQUIRED: Environment ID.
      --force                Skip the deletion confirmation prompt.
      --context string       CLI context name.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Delete the Schema Registry cluster for environment "env-12345".

::

  confluent schema-registry cluster delete --environment env-12345

See Also
~~~~~~~~

* :ref:`confluent_schema-registry_cluster` - Manage Schema Registry cluster.
