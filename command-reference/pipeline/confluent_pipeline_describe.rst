..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_pipeline_describe:

confluent pipeline describe
---------------------------

Description
~~~~~~~~~~~

Describe a Stream Designer pipeline.

::

  confluent pipeline describe <pipeline-id> [flags]

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

Examples
~~~~~~~~

Describe Stream Designer pipeline "pipe-12345".

::

  confluent pipeline describe pipe-12345

See Also
~~~~~~~~

* :ref:`confluent_pipeline` - Manage Stream Designer pipelines.
