..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_plugin:

confluent plugin
================

Description
~~~~~~~~~~~

Brief description:
	Plugins are standalone executable files that allow users to extend the 
	functionality of the confluent CLI. Plugins can be called from the confluent
	CLI as if they were built-in commands.

Making plugins discoverable by the CLI:
	The filename of a plugin makes it discoverable by the confluent CLI and 
	determines its command syntax. The filename must begin with ``confluent-``. It 
	must be executable and located in a directory on the user's $PATH. Each
	dash (-) in a plugin filename delimits a subcommand in the callable syntax.
	For example, the plugin file ``/User/me/plugins/confluent-demo-env-create`` 
	would be executed with the confluent CLI command ``confluent demo env create``.
	Additionally, the parent directory must be on the user's $PATH. For example,
	by adding ``export PATH=$PATH:/User/me/plugins`` to the user's .zshrc file.

Arguments and flags with plugins:
	Arguments and flags can be passed with plugin commands. It is the plugin's
	responsibility to validate and parse them. For example, if you run 
	``confluent demo env create arg0 --flag0 true``, the confluent CLI will first
	look for a plugin with the longest possible name: ``confluent-demo-env-create-arg0`` 
	in this case. If that is not found, it will look for a plugin with the next longest 
	possible name: ``confluent-demo-env-create`` in this case. If it finds that, it will invoke
	the confluent-demo-env-create plugin, passing along ``arg0 --flag true`` for
	the plugin's code to parse.

Naming collisions with existing CLI commands and other plugins:
	Built-in confluent CLI commands take precedence over plugins if they share 
	the same name. For example, there is a built-in ``confluent kafka cluster list``
	command. A plugin named ``confluent-kafka-cluster-list`` on the user's $PATH will 
	therefore not run. The built-in command will be run, along with a warning that the plugin 
	has been ignored.  Partial overlap between a plugin’s name and a built-in command, however, 
	is allowed. For example, a plugin named ``confluent-kafka-cluster-rebuild`` would be 
	callable with the command ``confluent kafka cluster rebuild``, since the name does not 
	exactly match a built-in command. If two or more plugins with the same name are found in the
	user's $PATH, the first one found in the $PATH is given precedence. Any subsequent plugin files 
	with the same name will be ignored.

.. toctree::
   :hidden:

   confluent_plugin_list

Subcommands
~~~~~~~~~~~

============================== ======================================
           Command                          Description              
============================== ======================================
 :ref:`confluent_plugin_list`   List Confluent CLI plugins in $PATH. 
============================== ======================================
