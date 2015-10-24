% Generated by roxygen2 (4.1.1): do not edit by hand
% Please edit documentation in R/rmq.R
\name{from.msgpack}
\alias{from.msgpack}
\title{create an R object from raw msgpack bytes}
\usage{
from.msgpack(x)
}
\arguments{
\item{x}{An raw byte vector of msgpack formatted bytes.}
}
\value{
The R object represented by x.
}
\description{
Given a vector of raw bytes written in msgpack format, \code{from.msgpack} converts these into an R object.
}
\details{
Lists, numeric vectors, integer vectors, string vectors, and raw byte vectors are supported.
}
\seealso{
\url{http://msgpack.org}

Other rmq.functions: \code{\link{r2r.call}};
  \code{\link{r2r.server}}; \code{\link{rmq.call}};
  \code{\link{rmq.default.addr}}; \code{\link{rmq.server}};
  \code{\link{to.msgpack}}
}
