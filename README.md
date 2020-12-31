Uroboros is A GNU/Linux monitoring tool focused on single processes. 

**WORK IN PROGRESS, BETA STAGE, EXPECT CRASHES AND WEIRD STUFF**

While 
utilities like top, ps and htop provide great overall details, they often lack useful temporal representation for 
specific processes, such visual representation of the process data points can be used to profile, debug and 
generally monitor the good health of a process. There are also tools like psrecord that can record some of the 
activity of a process, but then some graphical server is required for rendering, and they are not realtime.

Uroboros aims to fill this gap by providing a single tool to record, replay and render in realtime process 
runtime information in the terminal, without affecting the process performances like more invasive ptrace based 
solutions 
would do.

[![click here for a video](https://asciinema.org/a/382003.png)](https://asciinema.org/a/382003)

## Usage

For the moment there're no binary releases and building from sources is the only way (requires Go and make):

    sudo make install

To monitor by pid:

    sudo uro -pid 1234

To search by process name:

    sudo uro -search test-process

Only show a subset of tabs:

    sudo uro -pid 1234 -tabs "cpu, mem, io"

To save a recording on disk:

    sudo uro -pid 1234 -record /tmp/process-activity.dat

To play a recording from disk:

    uro -replay /tmp/process-activity.dat

For more options:
    
    ./_build/uro -help

### UI Navigation

* Left and right arrows to navigate tabs.
* Up and down arrows to scroll tables.
* `j` and `k` to navigate lists, enter to select an element.
* In replay mode, use `p` to pause, `f` to fast forward.
* Use `q` or `C-c` to quit.

## License

Released under the GPL 3 license.