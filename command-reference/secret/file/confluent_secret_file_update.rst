..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_secret_file_update:

confluent secret file update
----------------------------

Description
~~~~~~~~~~~

This command updates the encrypted secrets from the configuration properties file.

::

  confluent secret file update [flags]

.. tip:: For examples, see :platform:`Secrets Usage Examples|security/secrets.html#secrets-examples`.

Flags
~~~~~

::

      --config-file string           REQUIRED: Path to the configuration properties file.
      --local-secrets-file string    REQUIRED: Path to the local encrypted configuration properties file.
      --remote-secrets-file string   REQUIRED: Path to the remote encrypted configuration properties file.
      --config string                REQUIRED: List of key/value pairs of configuration properties.

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent_secret_file` - Secure secrets in a configuration properties file.
