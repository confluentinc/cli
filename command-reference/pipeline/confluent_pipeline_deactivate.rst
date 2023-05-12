..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_pipeline_deactivate:

confluent pipeline deactivate
-----------------------------

Description
~~~~~~~~~~~

Request to deactivate a pipeline.

::

  confluent pipeline deactivate <pipeline-id> [flags]

Flags
~~~~~

::

      --retained-topics strings   A comma-separated list of topics to be retained after deactivation.
  -o, --output string             Specify the output format as "human", "json", or "yaml". (default "human")
      --cluster string            Kafka cluster ID.
      --environment string        Environment ID.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Request to deactivate Stream Designer pipeline "pipe-12345" with 3 retained topics.

::

  confluent pipeline deactivate pipe-12345 --retained-topics topic1,topic2,topic3

See Also
~~~~~~~~

* :ref:`confluent_pipeline` - Manage Stream Designer pipelines.
