..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_api-key_list:

confluent api-key list
----------------------

Description
~~~~~~~~~~~

List the API keys.

::

  confluent api-key list [flags]

Flags
~~~~~

::

      --resource string          The resource ID to filter by. Use "cloud" to show only Cloud API keys.
      --current-user             Show only API keys belonging to current user.
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

List the API keys that belong to service account "sa-123456" on cluster "lkc-123456".

::

  confluent api-key list --resource lkc-123456 --service-account sa-123456

See Also
~~~~~~~~

* :ref:`confluent_api-key` - Manage API keys.
