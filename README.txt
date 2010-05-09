
About:
------------------------------------------------------------
This program will hopefully make it easier to compile 
source code in the go programming language.
A dependency graph is constructed from imports, this
is sorted with a topological sort to figure out legal
compile order.


Build:
------------------------------------------------------------

This should be as easy as running the script ./build.sh


Install:
------------------------------------------------------------

Copy the file: gd  somewhere it can be found ($PATH)


Try it Out:
------------------------------------------------------------

You can try to compile the same source using the generated
executable: gd


$ ./gd src          # will compile source inside src
$ ./gd -p src       # will print dependency info gathered
$ ./gd -s src       # will print legal compile order
$ ./gd -o name src  # will produce executable 'name' of
                    # source-code inside src directory



Philosophy (Babble?)
------------------------------------------------------------

Without a tool to figure out which order the source should
be compiled, Makefiles are usually the result. Makefiles
are static in nature, which make them a poor choice to handle
a dynamic problem like a changing source tree. They also make
flat structures quite common, which is far less intuitive
than a directory-tree package-structure, like; do I dare say
the word Java or C# uses :-)


-bjarneh
