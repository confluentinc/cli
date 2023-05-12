..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_pool_create:

confluent iam pool create
-------------------------

Description
~~~~~~~~~~~

Create an identity pool.

::

  confluent iam pool create <name> [flags]

Flags
~~~~~

::

      --filter string           REQUIRED: Filter which identities can authenticate with the identity pool.
      --identity-claim string   REQUIRED: Claim specifying the external identity using this identity pool.
      --provider string         REQUIRED: ID of this pool's identity provider.
      --description string      Description of the identity pool.
  -o, --output string           Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Create an identity pool named "DemoIdentityPool" with provider "op-12345":

::

  confluent iam pool create DemoIdentityPool --provider op-12345 --description new-description --identity-claim claims.sub --filter 'claims.iss=="https://my.issuer.com"'

See Also
~~~~~~~~

* :ref:`confluent_iam_pool` - Manage identity pools.
