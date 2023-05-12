..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_secret_master-key_generate:

confluent secret master-key generate
------------------------------------

Description
~~~~~~~~~~~

This command generates a master key. This key is used for encryption and decryption of configuration values.

::

  confluent secret master-key generate [flags]

.. tip:: For examples, see :platform:`Secrets Usage Examples|security/secrets.html#secrets-examples`.

Flags
~~~~~

::

      --local-secrets-file string   REQUIRED: Path to the local encrypted configuration properties file.
      --passphrase string           The key passphrase. To pipe from stdin use "-", e.g. "--passphrase -". To read from a file use "@<path-to-file>", e.g. "--passphrase @/User/bob/secret.properties".

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_secret_master-key` - Manage the master key for Confluent Platform.
