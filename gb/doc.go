/* 
   Copyright 2011 John Asmuth

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

/*
gb is nearly configuration-free project builder for the go language.

Jumpstart

For a single target which produces a command "X", make a directory called
X and put all your source files in it. Have no subdirectories with other
source. From within this directory you can run gb and it will build a
binary named X.

For a multiple target project, first create a workspace directory W (call
it whatever you want, but it will be referred to as W here). In the
directory W/x/y/z you can put the source files for a package that will be
imported with the path "x/y/z" by any other target built within W. In the
directory W/anything/c you can put the source files for a command that will
be named c. To build everything, cd into W and run gb.

Overview

With gb, one only has to run the simple command line, gb, in order to bring
all binaries up to date. To clean, gb -c. To install, gb -i. There are a
few other options to do other tasks, but never should one have to write 
scripts that, for instance, specify lists of which source files should be 
used. gb figures that out by analyzing the directory structure and source.

It works on a "one target per directory" rule to discover which source
files are to be compiled with one another. The idea is to obviate the need 
for any sort of build script.

gb will compile a set of "relevant targets" that exist within directories
listed on the command line, and any other targets that these have an import
depedence on. Note that a target in directory a/b/c will be included in this
set if the directory a/b is listed in the command line.

The most effective way to use gb is to have a top-level workspace directory
for all your code, similar to the concept of a workspace with eclipse.
Within the root directory, each target (whether it be a cmd or pkg) must
reside entirely within its own subdirectory. Nested targets are allowed; 
source in a directory a/b will not be compiled with source in a directory a.

When gb is run, it first recursively scans all subdirectories from its 
working directory to identify targets. Any directory that contains .go 
files or .c files will be identified as a target. Each target's name is then 
determined by first looking at its relative path, but can be overridden by 
either a gb.cfg file containing a new name, or a //target:<name> comment 
in one of the source files, before the package statement. This renaming of 
targets is primarily useful for projects that are intended to be installed 
with goinstall, which requires that the target name match a URL.

If gb is run within a directory that has a valid target, the target's name
will be taken from the containing directory, rather than the relative path,
".".

gb will match target names with import statements found in the source to 
determine the workspace dependency structure. It will use this structure to 
do incremental building correctly.

Packages are all built to the _obj directory in the root, and commands are 
built to the bin directory in the root. If -i is on, they will be copied to 
$GOROOT/pkg/$GOOS_$GOOARCH and $GOROOT/bin.


gb.cfg

If a directory has a file named "gb.cfg", gb will examine it for special
settings. They are entered each on their own line, in the form "key=val".
Currently valid keys are as follows.

workspace=<relative path>
  Running gb in the current directory will pretend the working directory
  is the one specified by the relative path.
target=<string>
  Set the package's import path or the binaries name.
makefile=true
  Always build this target with a local makefile.
ignore=true
  Never try to build a package in this directory.
ignoreall=true
  Never try to build a package in this directory or any of its
  subdirectories.
gcflags=<flag1> <flag2>...
  Include these flags on the compile line.


Tips

If your root contains a few packages and a few commands, but you only want 
to install the packages, run gb -Pi.

You can encode some information in file names. If a common value for $GOOS 
or $GOARCH appears in the file name in the form of *_VALUE*.go, that file 
will only be included if it matches $GOOS or $GOARCH. The flag *_unix*.go 
will match any of the unix-based $GOOS options.

Quickly check the build status of any target with gb -s. It will print out 
a list of targets, and will tell you if they are up to date or installed 
(if a target is installed, it is also up to date).

Tell gb to use goinstall to download packages with gb -g.

To build a simple one-target package or command, you can run gb from 
within its directory if you use either target.gb or a //target:<name> 
comment.

If you are working on something in $GOROOT/src and something outside at the 
same time, you can run gb -R to build dependencies in $GOROOT.

gb passes any command line arguments that begin with "-test." to testing
binaries, when you run gb -t.


Options:
 -i		Install build pkgs and cmds to $GOROOT/pkg/$GOOS_$GOARCH and
		$GOROOT/bin, respectively.

 -c		Remove all intermediate binaries.

 -N     Remove all installed binaries.

 -b		Definitely try to build. Useful when used as "-cb", to tell gb to
		first clean and then build.

 -g		Tell gb to use goinstall to build remote packages available at
		one of the following websites: googlecode.com, github.com,
		bitbucket.org and launchpad.net.

 -G		Same as -g, except goinstall fetches new code from the
        repository.

 -p		Attempt to build a package immediately once its dependencies are
		met and a processor is free.

 -s		List all targets that are relevant to the current build plan. If
		no directories are listed on the command line, all targets found
		will be listed. Otherwise, only targets that need to be up to
		date in order to bring the listed targets up to date will be
		listed.

 -S		Same as "-s", except import dependencies are also printed.

 -t		Run all tests contained in *_test.go source for the relevant
		targets. Behaves similarly to "make test". All additional
		command line arguments beginning with "-test." are passed to the
		test binary (see http://golang.org/cmd/gotest for details).

 -e		Exclusive target list. Do not attempt to build any packages that
		aren't in the directories listed on the command line.

 -v		Verbose. Print out all build instructions used.

 -m		Use makefiles. If this flag is set, and a target contains a
		makefile, that makefile will be used to build.

 -f		For use with "-M", force overwriting of makefiles. Otherwise
		you will be prompted when attempting to create a makefile for
		a target that already has one.

 -P		Build/clean/install only packages. Useful if you have a set of
		helper commands to test your packages, but don't want to install
		them to $GOROOT.

 -C		The same as "-P", but for commands.

 -R		Add targets in $GOROOT/src to those that gb can build. They will
		not be built automatically, but if a local target has an import
		dependence on a target in $GOROOT/src, it will be brought up to
		date. This works with the "-s" option. Using "-Rs" will list
		any targets in $GOROOT/src that the local targets depend on.

 --makefiles
 		Generate makefiles and a build script. In each relevant target,
		create a makefile that supports incremental building with the
		rest of the targets. The build script invokes each of the
		makefiles in a topological order, ensuring that running "./build"
		will always result in a correct build.

 --workspace
 		Create workspace.gb files for all listed targets. Doing this
 		allows you to run gb from within the target directories as if
 		you were running gb from the directory you ran --workspace in.

 --gofmt
 		Run gofmt on all source for relevant targets.

 --testargs
 		All command line arguments that follow --testargs will be
 		passed on to the test binaries, and otherwise ignored.

*/
package documentation
