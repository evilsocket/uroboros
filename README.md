Uroboros is a GNU/Linux monitoring tool focused on single processes. 

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

    go get github.com/evilsocket/uroboros/cmd/uro

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

### UI Navigation

* Left and right arrows to navigate tabs.
* Up and down arrows to scroll tables.
* `j` and `k` to navigate lists, enter to select an element.
* Use `p` to pause, `f` to fast forward in replay mode.
* Use `q` or `C-c` to quit.

## License

Released under the GPL3 license.