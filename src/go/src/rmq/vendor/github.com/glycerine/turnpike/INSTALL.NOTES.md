# ncurses dependencies

+ centos7: yum install ncurses-devel

+ OSX:

Courtesy of Michael R. Cook's blog post, here is how to fix
the ncurses install problem on OSX.

http://mrcook.uk/how-to-install-go-ncurses-on-mac-osx

How to Install Go NCurses on Mac OS X

If you’ve been trying to use the goncurses library in your Go project, but kept encountering various ncurses related errors - such as missing .pc files (ncurses.pc, form.pc, menu.pc, panel.pc) - then this post may help you get your Mac OS X, ncurses and golang installation running smoothly.

A couple of weekends ago I decided that porting the classic UMoria game could teach me alot about Go. It all went swimmingly well - if somewhat boring at times - and finally last night I was able to get the code to compile without errors, Woot! But when trying to install goncurses I kept getting ncurses errors. Boo!

I’d experimented a little with goncurses a couple of months back and somehow (exactly how I don’t remember) go it working - it was fun making my @ character run around the screen! - but when I updated to Go 1.5, I’d reset my golang dev environment, which included deleting old sources.

After several hours down the rabbit hole, I finally figured out what the issue was, and as I’m sure others have encountered these same problems, I wanted to share the solution here.

My current software versions are;

Mac OS X 10.9.5 (Mavericks)
Homebrew 0.9.5
Go 1.5
Golang NCurses support with goncurses

To use ncurses in your Go app you’ll likely be using the goncurses library, which itself depends on the pkg-config utility and the C-ncurses package to be installed on your system. With OS X we can use Homebrew to install these;

$ brew install pkg-config
$ brew install ncurses
At this point I’m presuming you already have Go installed and set up correctly (is your $GOPATH set?), so that we can use go get to install the goncurses library.

$ go get -u github.com/rthornton128/goncurses
On running that command you’ll likely encounter the following error;

# pkg-config --cflags ncurses form menu ncurses ncurses panel
Package ncurses was not found in the pkg-config search path.
Perhaps you should add the directory containing `ncurses.pc'
to the PKG_CONFIG_PATH environment variable
No package 'ncurses' found
Package form was not found in the pkg-config search path.
Perhaps you should add the directory containing `form.pc'
to the PKG_CONFIG_PATH environment variable
No package 'form' found
Package menu was not found in the pkg-config search path.
Perhaps you should add the directory containing `menu.pc'
to the PKG_CONFIG_PATH environment variable
No package 'menu' found
Package ncurses was not found in the pkg-config search path.
Perhaps you should add the directory containing `ncurses.pc'
to the PKG_CONFIG_PATH environment variable
No package 'ncurses' found
Package ncurses was not found in the pkg-config search path.
Perhaps you should add the directory containing `ncurses.pc'
to the PKG_CONFIG_PATH environment variable
No package 'ncurses' found
Package panel was not found in the pkg-config search path.
Perhaps you should add the directory containing `panel.pc'
to the PKG_CONFIG_PATH environment variable
No package 'panel' found
pkg-config: exit status 1
There are two different issues at work here.

PKG_CONFIG_PATH has not been set to point to the .pc files
Homebrew uses NCurses 6.0, which does not contain any .pc files.
We’ll get to the PKG_CONFIG_PATH in a moment, but first we need to fix those missing .pc files.

Install and Configure NCurses on Mac OS X

As of August 12, 2015, Homebrew started shipping Ncures 6.0, and it seems that that version doesn’t include the .pc files we need (5.9 does). Hopefully goncurses will be updated soon to work with the latest release, but for now we need to create those files manually.

I’ll be keeping mine in my home directory under; ~/ncurses-pc-files, but you can put yours in /usr/local/lib/pkgconfig if you desire. So go ahead and create the following;

# ./ncurses.pc

prefix=/usr/local/opt/ncurses
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include/ncurses
major_version=6
version=6.0.20150808

Name: ncurses
Description: ncurses 6.0 library
Version: ${version}
Requires:
Libs: -L${libdir} -lncurses
Cflags: -I${includedir} -I${includedir}/..


# ./form.pc

prefix=/usr/local/opt/ncurses
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include/ncurses
major_version=6
version=6.0.20150808

Name: form
Description: ncurses 6.0 add-on library
Version: ${version}
Requires: ncurses
Libs: -L${libdir} -lform
Cflags: -I${includedir} -I${includedir}/..


# ./menu.pc

prefix=/usr/local/opt/ncurses
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include/ncurses
major_version=6
version=6.0.20150808

Name: menu
Description: ncurses 6.0 add-on library
Version: ${version}
Requires: ncurses
Libs: -L${libdir} -lmenu
Cflags: -I${includedir} -I${includedir}/..


# ./panel.pc

prefix=/usr/local/opt/ncurses
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include/ncurses
major_version=6
version=6.0.20150808

Name: panel
Description: ncurses 6.0 add-on library
Version: ${version}
Requires: ncurses
Libs: -L${libdir} -lpanel
Cflags: -I${includedir} -I${includedir}/..
Note the Cflags setting. If you happen to have copied these files from ncurses v5.9 you’ll likely encounter the following error when trying to get the goncurses library;

In file included from ~/projects/src/github.com/rthornton128/goncurses/defs.go:8:
/opt/ncurses/include/ncurses/curses.h:60:10: fatal error: 'ncurses/ncurses_dll.h' file not found
#include <ncurses/ncurses_dll.h>
Adding -I${includedir}/.. will solve that.

Configuring the PKG_CONFIG_PATH

Now that we have our required .pc files we can go ahead and configure pkg-config. Add the following export to your bash/zsh profile config;

export PKG_CONFIG_PATH=/path/to/ncurses-pc-files
NOTE: make sure to change that path to point to the location of the files you just created, and restart your terminal.

Installing the goncurses library

With luck we’ll be able to install goncurses without any further issues;

$ go get -u github.com/rthornton128/goncurses
You should be presented with the command-prompt, without any errors or additional information.

I’ve not looked into how goncurses itself works so I’m not sure what will be needed to get it working with NCurses 6.0, but considering Homebrew now only provides that version for install, something will need to be changed. For now though, the above solution should work well enough to let you start experimenting with Go and ncurses on your Mac OS X.

