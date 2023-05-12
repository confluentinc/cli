..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_kafka_quota_update:

confluent kafka quota update
----------------------------

Description
~~~~~~~~~~~

Update a Kafka client quota.

::

  confluent kafka quota update <id> [flags]

Flags
~~~~~

::

      --ingress string              Update ingress limit for quota.
      --egress string               Update egress limit for quota.
      --add-principals strings      A comma-separated list of service accounts to add to the quota.
      --remove-principals strings   A comma-separated list of service accounts to remove from the quota.
      --description string          Update description.
      --name string                 Update display name.
      --context string              CLI context name.
  -o, --output string               Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Add "sa-12345" to an existing quota and remove "sa-67890".

::

  confluent kafka quota update cq-123ab --add-principals sa-12345 --remove-principals sa-67890

See Also
~~~~~~~~

* :ref:`confluent_kafka_quota` - Manage Kafka client quotas.
