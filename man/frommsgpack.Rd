\name{from.msgpack}
\alias{from.msgpack}
\title{
decode msgpack into R objects [in package rmq]
}
\description{
  from.msgpack(raw)
}

\value{
  from.msgpack(raw) converts a RAW byte vector into an R object.
}

\usage{
from.msgpack(raw)
}
\arguments{
  \item{raw}{an R raw byte vector}
}

\references{
https://github.com/glycerine/rmq
}
\author{
  Components authors noted in each individual package.
  Jason E. Aten <j.e.aten -at-sign- gmail.com>.
}


\examples{\dontrun{

library(rmq)

# make sample input, to be serialized into msgpack format.
input = list()
input$Blob = as.raw(c(0xff,0xf0,0x06))
input$D = c("hello","world")
input$E = c(32, 17)
input$G = c(-10.5)
sub = list()
sub$A = "hi"
sub$B = 4611686018427387904
sub$F = c(1.5, 3.4)
input$Sub = sub

# encode it in msgpack 2 (the updated spec).
o= to.msgpack(input)

# now decode it
r = from.msgpack(o)

# compare. r and input are identical
> str(r)
List of 5
 $ Blob: raw [1:3] ff f0 06
 $ D   : chr [1:2] "hello" "world"
 $ E   : num [1:2] 32 17
 $ G   : num -10.5
 $ Sub :List of 3
  ..$ A: chr "hi"
  ..$ B: num 4.61e+18
  ..$ F: num [1:2] 1.5 3.4

> str(input)
List of 5
 $ Blob: raw [1:3] ff f0 06
 $ D   : chr [1:2] "hello" "world"
 $ E   : num [1:2] 32 17
 $ G   : num -10.5
 $ Sub :List of 3
  ..$ A: chr "hi"
  ..$ B: num 4.61e+18
  ..$ F: num [1:2] 1.5 3.4
> 

}}

\keyword{utilities}
