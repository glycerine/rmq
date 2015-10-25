###########################################################################
## Copyright (C) 2014  Jason E. Aten                                     ##
##  rmq is licensed under the Apache 2.0. license.
##  http://www.apache.org/licenses/
###########################################################################

#' R Messaging and Queuing: msgpack2 serialization and RPC over websockets
#'
#' RMQ lets you do msgpack2 encoding and decoding, and provides a
#' websocket based remote procedure call (RPC) mechanism.
#'
#' The basic server and client functions are \code{\link{rmq.server}}
#' and \code{\link{rmq.call}}. The client and server communicate
#' internally by encoding and decoding to msgpack2 bytes on the wire.
#' Msgpack2 is the upgraded msgpack spec that distinguishes between
#' blobs and utf8 strings.
#'
#' Client and server use the websocket protocol which means the 
#' server can be accessed from the broswer-based javascript, and
#' the calls will go through firewalls without issue. The gorilla websocket 
#' implementation supports TLS certificates for security. A user
#' supplied R function is invoked by the server to handle each
#' incoming client connection.
#' 
#' You can also make use of \code{\link{to.msgpack}} and
#'  \code{\link{from.msgpack}} diretly for situations that do not
#' require remote procedure call or websockets.
#'
#' @references \url{https://github.com/glycerine/rmq}, \url{http://msgpack.org}
#' @docType package
#' @name rmq
NULL

#' The default address bound by \code{rmq.server}.
#' @family rmq functions
rmq.default.addr <- "127.0.0.1:9090"

#' Start an RMQ server, listening on specified IP and port.
#'
#' @param handler A handler R function taking a single argument
#' @param addr A string of "IP:port" format. The server will bind \code{addr}, and it must be available. Defaults to \code{rmq.default.addr}, which is "127.0.0.1:9090".
#' @return No return value. Blocks forever listening and calling the handler function when a request arrives. Ctrl-c will interrupt the server and shut it down. Call \code{rms.server()} again to re-start the server.
#' @examples
#' \dontrun{ 
#'    ## a) the simplest echo server - in R session #1.
#'    rmq.server(handler=function(msg) {msg}, addr="10.0.0.1:7777")
#'
#'   ## b) This second example is a simple handler
#'   ## that echos the input it receives, and adds a few other things.
#'   ## This would also be in R session #1, as an alternative
#'   ## to a) above.
#'  handler = function(x) {
#'    print("handler called back with argument x = ")
#'    print(x)
#'    reply=list()
#'    reply$hi = "there!"
#'    reply$yum = c(1.1, 2.3)
#'    reply$input = x
#'    reply
#'  }
#'  r = rmq.server(handler, addr=rmq.default.addr)
#'
#'  ## c) lastly the client call. In R session #2. You'll
#'  ## always need to run c) after first starting the 
#'  ## the server using a) or b) above in a separate
#'  ## R session.
#'  rmq.call("hello rmq!")
#'
#'  ## d) illustrate how the client call can pass complex
#'  ## nested list structured data.
#'  monster=list()
#'  eyes=list()
#'  eyes$description = c("red","glowing")
#'  monster$eyes = eyes
#'  monster$measurements = c(34, 22, 33)
#'
#'  ## finally, send the monster to the server.
#'  rmq.call(monster)
#'
#' }
#' 
#'
#' @family rmq functions
#'
rmq.server <- function(handler, addr = rmq.default.addr) {
   try(.Call("ListenAndServe", addr, handler, new.env(), PACKAGE="rmq"))
}


#' Send a message to a listening RMQ server.
#'
#' @param msg An R object. Can be a list. Internally this will be converted into msgpack and sent to the server.
#' @param addr A string of "IP:port" format. The server will bind \code{addr}, and it must be available. Defaults to \code{rmq.default.addr}, which is "127.0.0.1:9090".
#' @param timeout.msec A timeout value in milliseconds. A value of 0 means wait forever for a reply. It is recommended to use a small finite timeout such as the 5 second default, because there is no other way to interrupt the \code{rmq.call()} while it is waiting on the network. Issuing ctrl-c (SIGINT) in particular will not interrupt the \code{rmq.call()} in progress.
#' @return The return value is the response from the rmq server to the given msg.
#' @examples
#' \dontrun{ 
#'   rmq.call(msg, addr="10.0.0.1:7777", timeout.msec = 1000)
#' }
#'
#' @family rmq functions
#'
rmq.call <- function(msg, addr = rmq.default.addr, timeout.msec = 5000) {
  try(.Call("RmqWebsocketCall", addr, msg, timeout.msec, PACKAGE="rmq"))
}

#' serialize an R object to raw msgpack bytes
#'
#' Given an R object, \code{to.msgpack} will convert that object to a vector of raw bytes written in msgpack format.
#'
#' Lists, numeric vectors, integer vectors, string vectors, and raw byte vectors are supported.
#'
#' @param x An R object to be serialized. Lists, numeric vectors, raw vectors, and string vectors are supported.
#' @return A raw byte vector containing the msgpack serialized object.
#'
#' @examples
#' 
#   \dontrun{
#'    x=list()
#'    x$hello = "rmq"
#'    raw=to.msgpack(x)
#'    y=from.msgpack(raw)
#'    ## y and x should be equal
#'
#'
#' @family rmq functions
#'
#' @seealso \url{http://msgpack.org}
#'
to.msgpack <- function(x) {
  .Call("ToMsgpack", x)
}

#' create an R object from raw msgpack bytes
#'
#' Given a vector of raw bytes written in msgpack format, \code{from.msgpack} converts these into an R object.
#'
#' Lists, numeric vectors, integer vectors, string vectors, and raw byte vectors are supported.
#'
#' @param x An raw byte vector of msgpack formatted bytes.
#' @return The R object represented by x.
#'
#' @family rmq functions
#'
#' @seealso \url{http://msgpack.org}
#'
from.msgpack <- function(x) {
  .Call("FromMsgpack", x)
}


#' Start a server expecting serialized then msgpacked R objects.
#'
#' \code{r2r.call} calls on R's native \code{serialize()} function,
#' the encodes those bytes in msgpack and sends them over
#' to a waiting \code{r2r.server}, which turns them back
#' into R objects before passing them to the handler.
#'
#'  @details
#'  This is an example of how to use \code{rmq.server} to
#'  good effect. While \code{rmq.server} is designed to allow
#'  cross-language messaging, it may also be the case
#'  that only R sessions wish to communicate.
#'  If both client and server speak R's
#'  XDR based serialization protocol
#'  (e.g. if both ends are R sessions), then
#'  we can \code{serialize()} arbitrary R objects into
#'  msgpack RAW bytes, transmit those RAW bytes,
#'  and then \code{unserialize()} the XDR back into full R
#'  objects. Although not-interoperable with most other
#'  languages, this does mean that we can exchange
#'  *any* R object. The msgpack support for language
#'  interop is limited to numeric arrays, string arrays,
#'  RAW arrays, integer arrays, lists, and 
#'  recursively nested lists.
#'  While this level of msgpack support does cover
#'  most of the inter-language use cases, sometimes
#'  we want to serialize full R objects without
#'  restriction. For such purposes, the approach
#'  demonstrated in the r2r.server() call and the
#'  r2r.call() come in handy.
#' 
#'  Caveat: you client-server protocol can no
#'  longer be evolved by adding new fields to the
#'  msgpack. If you want to be able to evolve your
#'  cluster gracefully over time, you may be
#'  better sticking to msgpack.
#'
#' @examples
#' \dontrun{
#'
#'  ## R session 1 - start the server, giving it
#'  ##  a handler to call on arrival of each new message.
#'
#'  handler = function(x) {
#'    print("handler called back with argument x = ")
#'    print(x)
#'    print("computing and returning x$f(x$arg)")    
#'    x$f(x$arg)
#'  }
#'  r = r2r.server(handler)
#'
#'  ## lastly the client call - in R session #2
#'   x=list()
#'   x$arg=c(1,2,3)
#'   x$f = function(y) { sum(y) }
#'   r2r.call(x)
#' }
#' 
#' @family rmq functions
#'
r2r.server <- function(handler, addr=rmq.default.addr) {  
  unser.handler = function(x) {
    handler(unserialize(x))
  }
  
  r = rmq.server(unser.handler, addr)
}

#' Send an R object to a listening RMQ server.
#'
#' \code{r2r.call()} is the client counter-part to \code{r2r.server()}
#'
#' @examples
#'
#' \dontrun{ 
#'   x=list()
#'   x$arg=c(1,2,3)
#'   x$f = function(y) { sum(y) }
#'   r2r.call(x)
#' }
#' @family rmq functions
#'
r2r.call <- function(msg, addr = rmq.default.addr) {
  rmq.call(serialize(msg, connection=NULL), addr)
}

