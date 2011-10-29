cd $GOROOT/src

#!/usr/bin/env bash
# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e
if [ ! -f env.bash ]; then
	echo 'make.bash must be run from $GOROOT/src' 1>&2
	exit 1
fi
. ./env.bash

if ld --version 2>&1 | grep 'gold.* 2\.20' >/dev/null; then
	echo 'ERROR: Your system has gold 2.20 installed.'
	echo 'This version is shipped by Ubuntu even though'
	echo 'it is known not to work on Ubuntu.'
	echo 'Binaries built with this linker are likely to fail in mysterious ways.'
	echo
	echo 'Run sudo apt-get remove binutils-gold.'
	echo
	exit 1
fi

# Create target directories
if [ "$GOBIN" = "$GOROOT/bin" ]; then
	mkdir -p "$GOROOT/bin"
fi
mkdir -p "$GOROOT/pkg"

GOROOT_FINAL=${GOROOT_FINAL:-$GOROOT}

MAKEFLAGS=${MAKEFLAGS:-"-j4"}
export MAKEFLAGS
unset CDPATH	# in case user has it set

rm -f "$GOBIN"/quietgcc
CC=${CC:-gcc}
export CC
sed -e "s|@CC@|$CC|" < "$GOROOT"/src/quietgcc.bash > "$GOBIN"/quietgcc
chmod +x "$GOBIN"/quietgcc

rm -f "$GOBIN"/gomake
(
	echo '#!/bin/sh'
	echo 'export GOROOT=${GOROOT:-'$GOROOT_FINAL'}'
	echo 'exec '$MAKE' "$@"'
) >"$GOBIN"/gomake
chmod +x "$GOBIN"/gomake

# TODO(brainman): delete this after 01/01/2012.
rm -f "$GOBIN"/gotest	# remove old bash version of gotest on Windows

if [ -d /selinux -a -f /selinux/booleans/allow_execstack -a -x /usr/sbin/selinuxenabled ] && /usr/sbin/selinuxenabled; then
	if ! cat /selinux/booleans/allow_execstack | grep -c '^1 1$' >> /dev/null ; then
		echo "WARNING: the default SELinux policy on, at least, Fedora 12 breaks "
		echo "Go. You can enable the features that Go needs via the following "
		echo "command (as root):"
		echo "  # setsebool -P allow_execstack 1"
		echo
		echo "Note that this affects your system globally! "
		echo
		echo "The build will continue in five seconds in case we "
		echo "misdiagnosed the issue..."

		sleep 5
	fi
fi

(
	cd "$GOROOT"/src/pkg;
	bash deps.bash	# do this here so clean.bash will work in the pkg directory
) || exit 1
bash "$GOROOT"/src/clean.bash

for i in lib9 libbio libmach cmd
do
	echo; echo; echo %%%% making $i %%%%; echo
	gomake -C $i install
done

cd -

echo
echo
echo "Now you can build the rest of Go using gb (assuming gb is already built) by running it from $GOROOT/src"
echo
echo "Or you can build your own targets using gb -R, and gb will only build the core packages that you need"