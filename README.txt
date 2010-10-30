
About:
------------------------------------------------------------
This program will hopefully make it easier to compile 
source code in the go programming language.
A dependency graph is constructed from imports, this
is sorted with a topological sort to figure out legal
compile order. [865]g and [865]l is used to compile and
link program. Testing and formatting is also automated.


Install:
------------------------------------------------------------

Run the script ./build.sh

NOTE: It will try to copy the file 'gd' to $HOME/bin
      and if that directory is not present, an error
      message will be displayed, it's not the most
      advanced install script.. :-)



Try it Out:
------------------------------------------------------------

You can try to compile the same source using the generated
executable: gd


$ gd src          # will compile source inside src
$ gd -p src       # will print dependency info gathered
$ gd -s src       # will print legal compile order
$ gd src -test    # will run unit-tests
$ gd src -fmt     # will format (gofmt) the source-code
$ gd src -o gd    # will compile and link executable


Philosophy (Babble?)
------------------------------------------------------------

Without a tool to figure out which order the source should
be compiled, Makefiles are usually the result. Makefiles
are static in nature, which make them a poor choice to handle
a dynamic problem like a changing source tree. They also make
flat structures quite common, since this usually simplifies
the Makefiles, but makes organisation far less intuitive than
a directory-tree package-structure.


Completion
------------------------------------------------------------

A small completion script for bash is placed in util/


Logo
------------------------------------------------------------

The logo was made with LaTeX and tikz, it's basically just
an upside down g filled with yellow..

=start LaTeX

\documentclass[12pt]{article}
\usepackage{tikz}
\usepackage{nopageno}


\begin{document}
\begin{tikzpicture}[remember picture,overlay]
  \node [scale=75,fill=black!100,opacity=.8, rounded corners]
   at (current page.center) {}; 
  \node [rotate=180,scale=63,text opacity=0.9,yellow]
   at (current page.center) {g};
\end{tikzpicture}
\end{document}


=end LaTeX



-bjarneh
