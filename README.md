# rmq: R Messaging and Queuing


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

See also the test.r2r.call() and test.r2r.server() examples which demonstrate transporting full blown R objects over the RMQ transport layer.

### copyright and license

Copyright 2015 Jason E. Aten, Ph.D.

License: Apache 2.0 for the top level RMQ code and integration. Individual vendored library components include their own licenses which are Apache2, MIT, or BSD style. See the src/vendor subdirectories for details.

Requires: Go 1.5.1 for GO15VENDOREXPERIMENT=1
