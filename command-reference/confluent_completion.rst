..
   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.

.. _confluent_completion:

confluent completion
--------------------

Description
~~~~~~~~~~~

Use this command to print the shell completion
code for the specified shell (Bash/Zsh only). The shell code must be evaluated to provide
interactive completion of ``confluent`` commands.

Install Bash completions on macOS:
  #. Install Homebrew (https://brew.sh/).

  #. Install Bash completions using the ``brew`` command:
  
     ::
     
        brew install bash-completion
  
  #. Update your Bash profile:
  
     ::
     
       echo '[[ -r "$(brew --prefix)/etc/profile.d/bash_completion.sh" ]] && . "$(brew --prefix)/etc/profile.d/bash_completion.sh"' >> ~/.bash_profile
  
  #. Run the following command to install auto completion:
  
     ::
     
       confluent completion bash > $(brew --prefix)/etc/bash_completion.d/confluent

Install Bash completions on Linux:
  #.  Install Bash completion:

      ::

        sudo apt-get install bash-completion

  #. Set the ``confluent completion`` code for Bash to a file that's sourced on login:
  
     ::
     
        confluent completion bash > /etc/bash_completion.d/confluent

  #. Load the ``confluent completion`` code for Bash into the current shell:
  
     ::
  
        source /etc/bash_completion.d/confluent

  #. Add the source command above to your ``~/.bashrc`` or ``~/.bash_profile`` to enable completions for new terminals.

Install Zsh completions:
  Zsh looks for completion functions in the directories listed in the ``fpath`` shell variable.

  #. Put the ``confluent completion`` code for Zsh into a file in one the ``fpath`` directories, preferably one of the functions directories. For example:

     ::

        confluent completion zsh > ${fpath[1]}/_confluent

  #. Enable Zsh completions:
  
     ::
     
        autoload -U compinit && compinit

  #. Add the autoload command in your ``~/.zshrc`` to enable completions for new terminals. If you encounter error messages about insecure files, you can resolve by running the ``chown`` command to change the ``_confluent`` file to the same ``user:group`` as the other files in ``${fpath[1]}/``.


::

  confluent completion <bash|zsh> [flags]

Global Flags
~~~~~~~~~~~~

::

  -h, --help            Show help for this command.
      --unsafe-trace    Equivalent to -vvvv, but also log HTTP requests and responses which may contain plaintext secrets.
  -v, --verbose count   Increase verbosity (-v for warn, -vv for info, -vvv for debug, -vvvv for trace).

See Also
~~~~~~~~

* :ref:`confluent-ref` - Confluent CLI.
