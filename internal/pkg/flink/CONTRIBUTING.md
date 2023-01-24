#### Code Architecture

The following diagram represents approximately the code structure.

<img src="https://user-images.githubusercontent.com/11739405/212975377-804fc1c3-4e93-45f6-86c0-6eb012312779.png" width="700" >

We have view-only dumb components and controllers to manage them. The controllers should ideally communicate through the ApplicationController to
maintain a unidirectional data flow and avoid creating circular dependencies: view -> x-controller -> aplicationController -> x-controller (if necessary). That should make the application easier to debug and further development easier to manage without getting convoluted. This diagram might be outdated by the time you see this, but this might help you have an idea of how the whole architecture.

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

As an example, I've added [option + arrow navigation](./main.go#L29) in the current example.

#### Inspirations for testing tview - we prob won't test this though

https://github.com/rivo/tview/pull/441
https://github.com/rivo/tview/pull/5/
