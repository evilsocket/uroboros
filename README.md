Uroboros is a GNU/Linux monitoring tool focused on single processes. 

![Build and Test](https://github.com/evilsocket/uroboros/workflows/Build%20and%20Test/badge.svg)

While 
utilities like top, ps and htop provide great overall details, they often lack useful temporal representation for 
specific processes, such visual representation of the process data points can be used to profile, debug and 
generally monitor its good health. There are tools like psrecord that can record some of the 
activity of a process, but some graphical server is required for rendering, and it's neither complete nor realtime.

Uroboros aims to fill this gap by providing a single tool to record, replay and render in realtime process 
runtime information in the terminal, without affecting the process performances like more invasive ptrace based 
solutions 
would do.

[![click here for a video](https://asciinema.org/a/382091.png)](https://asciinema.org/a/382091)

**Work in progress**

## Usage

For the moment there are no binary releases and building from sources is the only way (requires the go compiler, 
will install the binary in $GOPATH/bin):

    # make sure go modules are used
    GO111MODULE=on go get github.com/evilsocket/uroboros/cmd/uro

To monitor by pid:

    sudo uro -pid 1234

To search by process name:

    sudo uro -search test-process

Only show a subset of tabs:

    sudo uro -pid 1234 -tabs "cpu, mem, io"

To save a recording on disk:

    sudo uro -pid 1234 -record /tmp/process-activity.dat

To play a recording from disk (works on any OS and does not require sudo):

    uro -replay /tmp/process-activity.dat

For more options:
    
    uro -help

### Keybindings

|           Key            | Action                                                     |
| :----------------------: | ---------------------------------------------------------- |
| <kbd>&lt;Right&gt;</kbd> | Show the next tab view. |
| <kbd>&lt;Left&gt;</kbd>  | Show the previous tab view. |
| <kbd>&lt;Down&gt;</kbd>  | Scroll down tables. |
| <kbd>&lt;Up&gt;</kbd>    | Scroll up tables. |
|       <kbd>j</kbd>       | Scroll down lists. |
|       <kbd>k</kbd>       | Scroll up lists. |
| <kbd>&lt;Enter&gt;</kbd> | Select list elements. |
|       <kbd>p</kbd>       | Pause (default and replay modes). |
|       <kbd>f</kbd>       | Fast forward (replay mode). |
|       <kbd>q</kbd> / <kbd>&lt;C-c&gt;</kbd> | Quit uro. |

## License

Released under the GPL3 license.
