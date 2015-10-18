
all:	
	# We have to build the C code definitions separately in a library now. Despite
	# what the early CGO tutorials say, you can only have declarations inline in the
	# the .go files now. https://code.google.com/p/go/issues/detail?id=3116 and
	# the cgo command documents here: https://golang.org/doc/go1#cmd_cgo specifically:
	# In Go 1, the cgo command uses a different _cgo_export.h file, which is generated for packages containing //export lines. The _cgo_export.h file now begins with the C preamble comment, so that exported function definitions can use types defined there. This has the effect of compiling the preamble multiple times, so a package using //export must not put function definitions or variable initializations in the C preamble. 
	# This library will be called libinterface.a. The linker commands that
	# link against it are actually embedded in the rmq.go file at the top.
	# Namely: #cgo LDFLAGS: -lR ${SRCDIR}/libinterface.a
	#
	gcc -c -o interface.o cpp/interface.cpp -Iinclude/
	ar cr go/libinterface.a interface.o
	go build -buildmode=c-shared  -o rmq.so  go/rmq.go

r-install-test:
	cd ../..  &&  R CMD INSTALL rmq
	echo "require(rmq); .Call('SayHello', 'gracey', PACKAGE = 'rmq')" | R --no-save --quiet

c-test: go/rmq.go cpp/_example_main.c
	gcc -c -o interface.o cpp/interface.cpp -Iinclude/
	ar cr go/libinterface.a interface.o
	go build -buildmode=c-shared  -o example.so  go/rmq.go
	gcc -o _example_main cpp/_example_main.c example.so -I. -Iinclude/
	./_example_main

clean:
	rm -f *~ rmq.h rmq.so _example_main go/libinterface.a  *.o example.so example.h cpp/*~ go/*~

