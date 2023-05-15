### Run test converage

The following command runs the tests and generates a coverage report in the `coverage` folder.
```
make test-coverage
```
You can then open the coverage.html file from the repository root in your browser to see the coverage report for each file.

### Generating Mocks
We have generated mocked as we need them. We have a make target for controllers, store and client. You can look at the mock_generator.go file as an exampple if you need to mock other things.

```
make generate-mocks
```

More info on how to generate mocks: https://github.com/golang/mock
#### Interacting with Clipboard in interactive mode

We use currently using https://github.com/atotto/clipboard to copy and write to the clipboard. It works for window, mac and linux. For the latter however, it requires 'xclip' or 'xsel' command to be installed. It doesn't seem to exist a workaround for now. One of the reasons to pick this lib, is because it's already being used in our current CLI, where this client will get integrated.

Quoted from [here](https://github.com/d-tsuji/clipboard):
Unfortunately, I don't think it's feasible for Linux to build clipboard library, since the library needs to be referenced as a daemon in order to keep the clipboard data. This approach is the same for xclip and xsel.

Go has an approach to running its own programs as external processes, such as VividCortex/godaemon and sevlyar/go-daemon. But these cannot be incorporated as a library, of course. xclip and xsel can also be achieved because they are completed as binaries, not libraries.

So it turns out that it is not possible to achieve clipboard in Linux as a library.

#### Debugging in VS Code

The current settings.json should allow you to debug the application. [You might have to install delve for that.](https://www.rookout.com/blog/golang-debugging-tutorial/#:~:text=To%20install%20Delve%20on%20VS,you%20get%20started%20with%20debugging)

#### Adding custom shortcuts to the client

An important part of a good command line tool are shortcuts. To add custom shortcuts with go-prompt, you can take a look here at:

https://stackoverflow.com/questions/6205157/how-to-set-keyboard-shortcuts-to-jump-to-beginning-end-of-line

We have a couple of examples in the codebase.

#### Inspirations for testing tview - we prob won't test this though

https://github.com/rivo/tview/pull/441
https://github.com/rivo/tview/pull/5/
