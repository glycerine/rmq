require(rmq)
require(testthat)

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

r = .Call("ListenAndServe", "127.0.0.1:9090", handler, new.env(), package="rmq")

# in client terminal
require(rmq)
options(error=recover)
res=rmqcall("127.0.0.1:9090", c(4,5,2))
