..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_pipeline_list:

confluent pipeline list
-----------------------

Description
~~~~~~~~~~~

Display pipelines in the current environment and cluster.

::

  confluent pipeline list [flags]

Flags
~~~~~

::

  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")
      --cluster string       Kafka cluster ID.
      --environment string   Environment ID.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_pipeline` - Manage Stream Designer pipelines.
