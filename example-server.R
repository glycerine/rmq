## file: example-server.R

## RMQ is, for the moment, a simple client/server system.
## This file example-server.R starts a server,
## which should be run in R session #1 first.

## The companion file example-client.R should
## then be run in a separate R session (#2), and
## it will talk to the server in session #1.

## To run the code in the R file, either copy and
## paste it into the terminal, or use
## source("example-server.R") from the R command
## prompt.

## In R session #1, start the server first. In
## order to start the server, we must define an
## R function to be called back to upon receipt
## of any message. This is handler:

require(rmq)

handler = function(x) {
    print("handler called back with argument x = ")
    print(x)
    print("computing and returning x$f(x$arg)")    
    result=list()
    result$f.evaluated.with.arg = x$f(x$arg)
    result$a.closure.returned = function() { paste(x) }
}

## this will block, server requests, until the user
## issues ctrl-c (or sends the R process SIGINT).

r = r2r.server(handler)

