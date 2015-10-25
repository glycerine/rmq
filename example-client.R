## file: example-client.R

## RMQ is, for the moment, a simple client/server system.
## This file example-client.R sends requests to
## a server which should be run before the client
## code so that the client can find it. So start
## two R sessions, run example-server.R from #1,
## and then turn to #2 to run the client code
## below.

## make some data and a function to have evaluated
## remotely.
##

require(rmq)

request=list()
request$f = function(x) { diff(x) }
request$arg = c(4,5,1)

## not run:
r = r2r.call(request)
r
