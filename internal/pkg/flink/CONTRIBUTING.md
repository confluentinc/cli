#### Code Architecture

The following diagram represents approximately the code structure.

<img src="https://user-images.githubusercontent.com/11739405/212975377-804fc1c3-4e93-45f6-86c0-6eb012312779.png" width="700" >


We have view-only dumb components and controllers to manage them. The controllers should ideally communicate through the ApplicationController to 
maintain a unidirectional data flow and avoid creating circular dependencies: view -> x-controller -> aplicationController -> x-controller (if necessary). That should make the application easier to debug and further development easier to manage without getting convoluted. This diagram might be outdated by the time you see this, but this might help you have an idea of how the whole architecture. 

#### Adding custom shortcuts to the client

An important part of a good command line tool are shortcuts. To add custom shortcuts with go-prompt, you can take a look here at:

https://stackoverflow.com/questions/6205157/how-to-set-keyboard-shortcuts-to-jump-to-beginning-end-of-line

As an example, I've added [option + arrow navigation](./main.go#L29) in the current example.

#### Inspirations for testing tview - we prob won't test this though

https://github.com/rivo/tview/pull/441
https://github.com/rivo/tview/pull/5/
