..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_api-key_create:

confluent api-key create
------------------------

Description
~~~~~~~~~~~

Create API keys for a given resource. A resource is some Confluent product or service for which an API key can be created, for example ksqlDB application ID, or "cloud" to create a Cloud API key.

::

  confluent api-key create [flags]

Flags
~~~~~

::

      --resource string          REQUIRED: The resource ID. Use "cloud" to create a Cloud API key.
      --description string       Description of API key.
      --use                      Use the created API key for the provided resource.
      --context string           CLI context name.
      --environment string       Environment ID.
      --service-account string   Service account ID.
  -o, --output string            Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Create an API key with full access to cluster "lkc-123456":

::

  confluent api-key create --resource lkc-123456

Create an API key for cluster "lkc-123456" and service account "sa-123456":

::

  confluent api-key create --resource lkc-123456 --service-account sa-123456

See Also
~~~~~~~~

* :ref:`confluent_api-key` - Manage API keys.
