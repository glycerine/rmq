# rmq: R Messaging and Queuing

### Or: How to utilize Go libraries from R.

Package rmq provides messaging based on msgpack and websockets. It demonstrates calling from R into Golang (Go) libraries to extend R with functionality available in Go.

 We use the Go library https://github.com/ugorji/go/codec for msgpack encoding and decoding. This is a high performance implementation. We use it in a mode where it only supports the updated msgpack 2 (current) spec.

For websockets, we use the terrific https://github.com/gorilla/websocket library.

Copyright 2015 Jason E. Aten, Ph.D.

License: Apache 2.0. Individual vendored components include their own licenses, which may differ.
