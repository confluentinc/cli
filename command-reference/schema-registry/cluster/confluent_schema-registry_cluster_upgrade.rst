..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_cluster_upgrade:

confluent schema-registry cluster upgrade
-----------------------------------------

Description
~~~~~~~~~~~

Upgrade the Schema Registry package for this environment.

::

  confluent schema-registry cluster upgrade [flags]

Flags
~~~~~

::

      --package string       REQUIRED: Specify the type of Stream Governance package as "essentials" or "advanced".
      --context string       CLI context name.
      --environment string   Environment ID.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Upgrade Schema Registry to the "advanced" package for environment "env-12345".

::

  confluent schema-registry cluster upgrade --package advanced --environment env-12345

See Also
~~~~~~~~

* :ref:`confluent_schema-registry_cluster` - Manage Schema Registry cluster.
