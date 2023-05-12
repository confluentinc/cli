..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_cluster_describe:

confluent schema-registry cluster describe
------------------------------------------

Description
~~~~~~~~~~~

Describe the Schema Registry cluster for this environment.

::

  confluent schema-registry cluster describe [flags]

Flags
~~~~~

::

      --api-key string       API key.
      --api-secret string    API key secret.
      --context string       CLI context name.
      --environment string   Environment ID.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_schema-registry_cluster` - Manage Schema Registry cluster.
