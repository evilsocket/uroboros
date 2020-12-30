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

<iframe src="https://www.youtube.com/embed/Kxl5F5yHi3E" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Usage

For the moment there're no binary releases and building from sources is the only way (requires Go and make):

    make uro

To monitor by pid:

    ./_build/uro -pid 1234

To search by process name:

    ./_build/uro -search test-process

Only show a subset of tabs:

    ./_build/uro -pid 1234 -tabs "cpu, mem, io"

For more options:
    
    ./_build/uro -help

Navigate tabs with left and right arrows, scroll tables with up and down arrows, lists and trees with j and k.

## License

Released under the GPL 3 license.