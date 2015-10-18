\name{rmq}
\alias{rmq}
\title{
the R message queue
}
\description{
Adds a local endpoint to the socket 's'. The endpoint can be then used by other
applications to connect to.
}

\value{
  If the function succeeds positive endpoint ID is returned.

\item{EBADF}{The provided socket is invalid.}

}

\usage{
rmq(socket)
}
\arguments{
  \item{socket}{a socket.}
}

\references{
https://github.com/glycerine/rmq
}
\author{
  Components authors noted in each individual package.
  rmq by Jason E. Aten <j.e.aten -at-sign- gmail.com>.
}


\examples{\dontrun{

library(rmq)

}}

\keyword{utilities}
