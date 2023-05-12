..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_schema-registry_region_list:

confluent schema-registry region list
-------------------------------------

Description
~~~~~~~~~~~

List Schema Registry cloud regions.

::

  confluent schema-registry region list [flags]

Flags
~~~~~

::

      --cloud string     Specify the cloud provider as "aws", "azure", or "gcp".
      --package string   Specify the type of Stream Governance package as "essentials" or "advanced".
  -o, --output string    Specify the output format as "human", "json", or "yaml". (default "human")

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

Examples
~~~~~~~~

List the Schema Registry cloud regions for AWS in the "advanced" package.

::

  confluent schema-registry region list --cloud aws --package advanced

See Also
~~~~~~~~

* :ref:`confluent_schema-registry_region` - Manage Schema Registry cloud regions.
