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

rmqcall <- function(msg, addr = default.addr) {
  .Call("RmqWebsocketCall", addr, msg, PACKAGE="rmq")
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
