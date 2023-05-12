..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_pipeline_update:

confluent pipeline update
-------------------------

Description
~~~~~~~~~~~

Update an existing pipeline.

::

  confluent pipeline update <pipeline-id> [flags]

Flags
~~~~~

::

      --name string              Name of the pipeline.
      --description string       Description of the pipeline.
      --ksql-cluster string      KSQL cluster for the pipeline.
      --activation-privilege     Grant or revoke the privilege to activate this pipeline. (default true)
      --update-schema-registry   Update the pipeline with the latest Schema Registry cluster.
  -o, --output string            Specify the output format as "human", "json", or "yaml". (default "human")
      --cluster string           Kafka cluster ID.
      --environment string       Environment ID.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Request to update Stream Designer pipeline "pipe-12345", with new name and new description.

::

  confluent pipeline update pipe-12345 --name test-pipeline --description "Description of the pipeline"

Grant privilege to activate Stream Designer pipeline "pipe-12345".

::

  confluent pipeline update pipe-12345 --activation-privilege true

Revoke privilege to activate Stream Designer pipeline "pipe-12345".

::

  confluent pipeline update pipe-12345 --activation-privilege false

Update Stream Designer pipeline "pipe-12345" with KSQL cluster ID "lksqlc-123456".

::

  confluent pipeline update pipe-12345 --ksql-cluster lksqlc-123456

Update Stream Designer pipeline "pipe-12345" with new Schema Registry cluster ID.

::

  confluent pipeline update pipe-12345 --update-schema-registry

See Also
~~~~~~~~

* :ref:`confluent_pipeline` - Manage Stream Designer pipelines.
