..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_secret_file_rotate:

confluent secret file rotate
----------------------------

Description
~~~~~~~~~~~

This command rotates either the master or data key. To rotate the master key, specify the current master key passphrase flag (``--passphrase``) followed by the new master key passphrase flag (``--passphrase-new``). To rotate the data key, specify the current master key passphrase flag (``--passphrase``).

::

  confluent secret file rotate [flags]

.. tip:: For examples, see :platform:`Secrets Usage Examples|security/secrets.html#secrets-examples`.

Flags
~~~~~

::

      --local-secrets-file string   REQUIRED: Path to the encrypted configuration properties file.
      --master-key                  Rotate the master key. Generates a new master key and re-encrypts with the new key.
      --data-key                    Rotate data key. Generates a new data key and re-encrypts the file with the new key.
      --passphrase string           Master key passphrase. You can use dash ("-") to pipe from stdin or @file.txt to read from file.
      --passphrase-new string       New master key passphrase. You can use dash ("-") to pipe from stdin or @file.txt to read from file.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_secret_file` - Secure secrets in a configuration properties file.
