# rmq: R Messaging and Queuing

## installation

pre-reqs:

a) make sure you have a working `go` (golang) installation. 

b) make sure you have a working gcc (C compiler) installed.

c) make sure you have R installed.

d) download `rmq`

~~~
$ go get -d github.com/glycerine/rmq
~~~

e) tell the C compiler where to find your R headers.

Locate your 'R.h' header, wherever it lives in your
filesystem (e.g. `find / -name 'R.h'` if you have to),
and add a '-I' flag to the src/Makefile in these
places, to reflect your local location of the 'R.h' and
related headers.

https://github.com/glycerine/rmq/blob/master/src/Makefile#L15

https://github.com/glycerine/rmq/blob/master/src/Makefile#L26

For example, if your R.h header has been installed in /usr/local/lib/R/include, then
you would have the Makefile lines look like this:
~~~
gcc -fPIC -O2 -c -o interface.o cpp/interface.cpp -Iinclude/ -I/usr/local/lib/R/include
~~~

f) then build:

~~~
$ cd $GOPATH/src/github.com/glycerine/rmq
$ make install
$ R
> require(rmq)
>
~~~

this should also work, once you have done `make install` to build vendor.tar.gz:

~~~
$ cd $GOPATH/src/github.com/glycerine
$ R CMD INSTALL rmq
~~~


## docs

[pdf in R documentation format](https://github.com/glycerine/rmq/blob/master/rmq.pdf)

[godoc API reference](https://godoc.org/github.com/glycerine/rmq)

[![GoDoc](https://godoc.org/github.com/glycerine/rmq?status.svg)](https://godoc.org/github.com/glycerine/rmq)

### RMQ Or: How to utilize Go libraries from R.

The much anticipated Go 1.5 release brought strong support for building C-style shared libraries (.so files) from Go source code and libraries. 

*This is huge*. It opens up many exciting new possibilities. In this proof-of-concept project (rmq), we explore using this new capability to extend R with Go libraries.

Package rmq provides messaging based on msgpack and websockets. It demonstrates calling from R into Golang (Go) libraries to extend R with functionality available in Go.

We use the Go library https://github.com/ugorji/go codec for msgpack encoding and decoding. This is a high performance implementation. We use it in a mode where it only supports the updated msgpack 2 (current) spec. This is critical for interoperability with other compiled languages that distiguish between utf8 strings and binary blobs (otherwise embedded '\0' zeros in blobs cause problems).

For websockets, we use the terrific https://github.com/gorilla/websocket library. As time permits in the future, we may extend more features aiming towards message queuing as well. The gorilla library supports securing your communication with TLS certs.

##Status

Excellent. Tested on OSX and Linux. Documentation has been written and is available. The package is functionally complete for the RPC over websockets and msgpack based serialization.  After interactive usage, I added SIGINT handling so that the web-server can be stopped during development with a simple ctrl-c at the R console. The client side will be blocked during calls (it does not poll back to R while waiting on the network) but has a configurable timeout (default 5 seconds), that allows easy client-side error handling.



## example R session, showing the msgpack library at work

~~~
> library(rmq)
> input = list() # make an R object to serialize.
> input$Blob = as.raw(c(0xff,0xf0,0x06))
> input$D = c("hello","world")
> input$E = c(32, 17)
> o=to.msgpack(input)
> o # look at the raw bytes in msgpack format
 [1] 83 a4 42 6c 6f 62 c4 03 ff f0 06 a1 44 92 a5 68 65 6c 6c 6f a5 77 6f 72 6c
[26] 64 a1 45 92 cb 40 40 00 00 00 00 00 00 cb 40 31 00 00 00 00 00 00
> from.msgpack(o) # now the inverse
$Blob
[1] ff f0 06

$D
[1] "hello" "world"

$E
[1] 32 17

> 
~~~

###sample session showing web-socket based RPC, from both the client and the server side:

server-side:
~~~
> require(rmq) 
> handler <- function(x) {
        print("handler called back with argument x = ")
        print(x)
        reply = list()
        reply$hi = "there!"
        reply$yum = c(1.1, 2.3)
        reply$input = x
        reply
    }
+ + + + + + + + > > 
> listenAndServe(handler, addr = "127.0.0.1:9090")
ListenAndServe listening on address '127.0.0.1:9090'...
[1] "handler called back with argument x = "
$hello
[1] "cran"  "this"  "is"    "great"

  [give Ctrl-c to stop the web-server]
>
~~~

client-side:
~~~
> require(rmq)
> my.message=list()
> my.message$hello =c("cran","this","is","great")
> rmq.call(addr = "127.0.0.1:9090", my.message)
$hi
[1] "there!"
$input
$input$hello
[1] "cran"  "this"  "is"    "great"
$yum
[1] 1.1 2.3
> 
~~~

## additional example scripts

I've added two example scripts, `example-server.R` and `example-client.R`. These live in the top level of the repo. Run `example-server.R` first. Then in a different window, run `example-client.R`. These are simultaneously bash scripts and R source-able scripts; you can run them straight from the shell if 'R' is on your PATH.


# And the reverse: embedding R inside your Golang program

In addition to using a Golang library under R, one can alternatively embed R as a library inside a Go executable. This is equally easy
using the SexpToIface() function. Here is an example. This is taken from the main source file [src/rmq/rmq.go](https://github.com/glycerine/rmq/blob/master/src/rmq/rmq.go). 

~~~
func main() {
        // Give an example also of how to embed R in a Go program.

        // Introduction to embedding R:
        //
        // While RMQ is mainly designed to embed Go under R, it
        // defines functions that make embedding R in Go
        // quite easy too. We use SexpToIface() to generate
        // a go inteface{} value. For simple uses, this may be
        // more than enough.
        //
        // If you wish to turn results into
        // a pre-defined Go structure, the interface{} value could
        // transformed into msgpack (as in encodeRIntoMsgpack())
        // and from there automatically parsed into Go structures
        // if you define the Go structures and use
        // https://github.com/tinylib/msgp to generate the
        // go struct <-> msgpack encoding/decoding boilerplate.
        // The tinylib/msgp library uses go generate and is
        // blazing fast. This also avoids maintaining a separate
        // IDL file. Your Go source code becomes the defining document
        // for your data structures.

        var iface interface{}
        C.callInitEmbeddedR()
        myRScript := "rnorm(100)" // generate 100 Gaussian(0,1) samples
        var evalErrorOccurred C.int
        r := C.callParseEval(C.CString(myRScript), &evalErrorOccurred)
        if evalErrorOccurred == 0 && r != C.R_NilValue {
                C.Rf_protect(r)
                iface = SexpToIface(r)
                fmt.Printf("\n Embedding R in Golang example: I got back from evaluating myRScript:\n")
                goon.Dump(iface)
                C.Rf_unprotect(1) // unprotect r
        }
        C.callEndEmbeddedR()
}
~~~


### copyright and license

Copyright 2015 Jason E. Aten, Ph.D.

License: Apache 2.0 for the top level RMQ code and integration. Individual vendored library components include their own licenses which are Apache2, MIT, or BSD style. See the src/vendor subdirectories for details.

Requires: Go 1.5.1 for GO15VENDOREXPERIMENT=1
