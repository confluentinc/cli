..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_byok_create:

confluent byok create
---------------------

Description
~~~~~~~~~~~

Bring your own key to Confluent Cloud for data at rest encryption (AWS and Azure only).

::

  confluent byok create <key> [flags]

Flags
~~~~~

::

      --key-vault string   The ID of the Azure Key Vault where the key is stored.
      --tenant string      The ID of the Azure Active Directory tenant that the key vault belongs to.
  -o, --output string      Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Register a new self-managed encryption key for AWS:

::

  confluent byok create "arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"

Register a new self-managed encryption key for Azure:

::

  confluent byok create "https://vault-name.vault.azure.net/keys/key-name" --tenant "00000000-0000-0000-0000-000000000000" --key-vault "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourcegroup-name/providers/Microsoft.KeyVault/vaults/vault-name"

See Also
~~~~~~~~

* :ref:`confluent_byok` - Manage your keys in Confluent Cloud.
