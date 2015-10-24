###########################################################################
## Copyright (C) 2014  Jason E. Aten                                     ##
##  rmq is licensed under the Apache 2.0. license.
##  http://www.apache.org/licenses/
###########################################################################

srv <- function(x) {
    .Call("Srv", x, PACKAGE="rmq")
}

callcount <- function() {
    .Call("Callcount", PACKAGE="rmq")
}

listenAndServe <- function(addr, handler) {
  invisible(.Call("ListenAndServe", addr, handler, new.env(), PACKAGE="rmq"))
}

rmqcall <- function(addr, msg) {
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
    print(paste("handler called back with argument x = ", paste(collapse=" ",sep=" ",x)))
    ##browser()
    reply=list()
    reply$hi = "there!"
    reply$yum = c(1.1, 2.3)
    reply$input = x
    reply
  }
  
  options(error=recover)
  
  r = listenAndServe("127.0.0.1:9090", handler)
}
