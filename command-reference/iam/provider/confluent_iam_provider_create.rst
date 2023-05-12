..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_provider_create:

confluent iam provider create
-----------------------------

Description
~~~~~~~~~~~

Create an identity provider.

::

  confluent iam provider create <name> [flags]

Flags
~~~~~

::

      --issuer-uri string    REQUIRED: URI of the identity provider issuer.
      --jwks-uri string      REQUIRED: JWKS (JSON Web Key Set) URI of the identity provider.
      --description string   Description of the identity provider.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Create an identity provider named "DemoIdentityProvider".

::

  confluent iam provider create DemoIdentityProvider --description "description of provider" --jwks-uri https://company.provider.com/oauth2/v1/keys --issuer-uri https://company.provider.com

See Also
~~~~~~~~

* :ref:`confluent_iam_provider` - Manage identity providers.
