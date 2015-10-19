###########################################################################
## Copyright (C) 2014  Jason E. Aten                                     ##
##  rmq is licensed under the Apache 2.0. license.
##  http://www.apache.org/licenses/
###########################################################################

srv <- function(x) {
    .Call("Srv", x, PACKAGE="rmq")
}

cli <- function(x) {
    .Call("Cli", x, PACKAGE="rmq")
}
