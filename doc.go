/*
Sumary: rmq passes msgpack2 messages over websockets between Golang and the R stats language. It is an R package.

# rmq: R Messaging and Queuing

### Or: How to utilize Go libraries from R.

The much anticipated Go 1.5 release brought strong support for building C-style shared libraries (.so files) from Go source code and libraries.

*This is huge*. It opens up many exciting new possibilities. In this project (rmq), we explore using this new capability to extend R with Go libraries.

Package rmq provides messaging based on msgpack and websockets. It demonstrates calling from R into Golang (Go) libraries to extend R with functionality available in Go.

We use the Go library https://github.com/ugorji/go codec for msgpack encoding and decoding. This is a high performance implementation. We use it in a mode where it only supports the updated msgpack 2 (current) spec. This is critical for interoperability with other compiled languages that distiguish between utf8 strings and binary blobs (otherwise embedded '\0' zeros in blobs cause problems).

For websockets, we use the terrific https://github.com/gorilla/websocket library. As time permits in the future, we may extend more features aiming towards message queuing as well. The gorilla library supports securing your communication with TLS certs.

##Status

Excellent. Tested on OSX and Linux. Documentation has been written and is available. The package is functionally complete for the RPC over websockets and msgpack based serialization.  After interactive usage, I added SIGINT handling so that the web-server can be stopped during development with a simple ctrl-c at the R console. The client side will be blocked during calls (it does not poll back to R while waiting on the network) but has a configurable timeout (default 5 seconds), that allows easy client-side error handling.

## structure of this repo

This repository is mainly structured as an R package. It is
designed to be built and installed into an R (staticial environment)
installation, using the standard tools for R.

For ease of godoc indexing, we've also provided symlinks to the
go files at the top level which should make it appear go-gettable
to godoc. It's not really meant to be standalone go-gettable,
as it doesn't directly create a re-usable go library. Instead
we target a c-shared library (rmq.so) that will install
into R using 'make install' or 'make build' followed by
doing `install.packages('./rmq_1.0.1.tar.gz', repos=NULL)`
from inside R (assuming the package is in your current directory;
if not then adjust the ./ part of the package path).

The code also serves as an example of how to
use golang inside R, so we want godoc.org to find
and index it. This may be somewhat tricky, but we
will try.

*/
package main
