..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_pool_update:

confluent iam pool update
-------------------------

Description
~~~~~~~~~~~

Update an identity pool.

::

  confluent iam pool update <id> [flags]

Flags
~~~~~

::

      --provider string         REQUIRED: ID of this pool's identity provider.
      --name string             Name of the identity pool.
      --description string      Description of the identity pool.
      --filter string           Filter which identities can authenticate with the identity pool.
      --identity-claim string   Claim specifying the external identity using this identity pool.
  -o, --output string           Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Update the description of identity pool "pool-123456":

::

  confluent iam pool update pool-123456 --provider op-12345 --description "New description."

See Also
~~~~~~~~

* :ref:`confluent_iam_pool` - Manage identity pools.
