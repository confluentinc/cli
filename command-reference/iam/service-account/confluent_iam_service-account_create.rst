..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_iam_service-account_create:

confluent iam service-account create
------------------------------------

Description
~~~~~~~~~~~

Create a service account.

::

  confluent iam service-account create <name> [flags]

Flags
~~~~~

::

      --description string   REQUIRED: Description of the service account.
  -o, --output string        Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

Create a service account named ``DemoServiceAccount``.

::

  confluent iam service-account create DemoServiceAccount --description "This is a demo service account."

See Also
~~~~~~~~

* :ref:`confluent_iam_service-account` - Manage service accounts.
