..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_pipeline_create:

confluent pipeline create
-------------------------

Description
~~~~~~~~~~~

Create a new pipeline.

::

  confluent pipeline create [flags]

Flags
~~~~~

::

      --name string           REQUIRED: Name of the pipeline.
      --description string    Description of the pipeline.
      --ksql-cluster string   KSQL cluster for the pipeline.
  -o, --output string         Specify the output format as "human", "json", or "yaml". (default "human")
      --cluster string        Kafka cluster ID.
      --environment string    Environment ID.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Create a new Stream Designer pipeline

::

  confluent pipeline create --name test-pipeline --ksql-cluster lksqlc-12345 --description "this is a test pipeline"

See Also
~~~~~~~~

* :ref:`confluent_pipeline` - Manage Stream Designer pipelines.
