..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_service-quota_list:

confluent service-quota list
----------------------------

Description
~~~~~~~~~~~

List Confluent Cloud service quota values by a scope (organization, environment, network, kafka_cluster, service_account, or user_account).

::

  confluent service-quota list <quota-scope> [flags]

Flags
~~~~~

::

      --environment string   Environment ID.
      --cluster string       Kafka cluster ID.
      --quota-code string    Filter the result by quota code.
      --network string       Filter the result by network id.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

List Confluent Cloud service quota values for scope "organization".

::

  confluent service-quota list organization

See Also
~~~~~~~~

* :ref:`confluent_service-quota` - Look up Confluent Cloud service quota limits.
