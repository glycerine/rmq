###########################################################################
## Copyright (C) 2014  Jason E. Aten                                     ##
##  rmq is licensed under the Apache 2.0. license.
##  http://www.apache.org/licenses/
###########################################################################

default.addr <- "127.0.0.1:9090"

srv <- function(x) {
    .Call("Srv", x, PACKAGE="rmq")
}

callcount <- function() {
    .Call("Callcount", PACKAGE="rmq")
}

listenAndServe <- function(handler, addr = default.addr) {
  invisible(.Call("ListenAndServe", addr, handler, new.env(), PACKAGE="rmq"))
}

rmq.call <- function(msg, addr = default.addr, timeout.msec = 5000) {
  .Call("RmqWebsocketCall", addr, msg, timeout.msec, PACKAGE="rmq")
}


to.msgpack <- function(x) {
  .Call("ToMsgpack", x)
}

from.msgpack <- function(x) {
  .Call("FromMsgpack", x)
}

block <- function() {
  .Call("BlockInSelect", PACKAGE="rmq")
}


test.server <- function() {

  handler = function(x) {
    print("handler called back with argument x = ")
    print(x)
    ##browser()
    reply=list()
    reply$hi = "there!"
    reply$yum = c(1.1, 2.3)
    reply$input = x
    reply
  }
  
  options(error=recover)
  
  r = listenAndServe(handler, addr=default.addr)
}

## test.r2.rserver demonstrates the power of
## the r2r.server and full serialization of R
## objects including functions.
##
## This server expects to receive a message x that is assembled
## like this:
##
##   x = list()
##   x$f = <function to evaluate remotely>
##   x$arg = <argument to give to function x$f>
##
## Return value: the result of doing x$f(x$arg).
##
test.r2r.server <- function() {

  handler = function(x) {
    print("handler called back with argument x = ")
    print(x)
    print("computing and returning x$f(x$arg)")    
    x$f(x$arg)
  }
    
  r = r2r.server(handler, addr=default.addr)
}

r2r.server <- function(handler, addr=default.addr) {

  ## If both client and server speak R's
  ## XDR based serialization protocol
  ## (e.g. if both ends are R sessions), then
  ## we can serialize arbitrary R objects into
  ## msgpack RAW bytes, transmit those RAW bytes,
  ## and then unserialize the XDR back into full R
  ## objects. Although not-interoperable with most other
  ## languages, this does mean that we can exchange
  ## *any* R object. The msgpack support for language
  ## interop is limited to numeric arrays, string arrays,
  ## RAW arrays, lists, and recursively nested lists.
  ## While this level of msgpack support does cover
  ## 99% of the inter-language use case, sometimes
  ## we want to serialize full R objects without
  ## restriction. For such purposes, the approach
  ## demonstrated in the r2r.server() call and the
  ## r2r.call() come in handy.
  ##
  ## Caveat: you client-server protocol can no
  ## longer be evolved by adding new fields to the
  ## msgpack.
  
  unser.handler = function(x) {
    handler(unserialize(x))
  }
  
  r = listenAndServe(unser.handler, addr)  
}

r2r.call <- function(msg, addr = default.addr) {
  ## r2r.call() is the client counter-part to r2r.server().
  rmq.call(serialize(msg, connection=NULL), addr)
}

test.r2r.call <- function() {
 x=list()
 x$arg=c(1,2,3)
 x$f = function(y) { sum(y) }
 r2r.call(x)
}
