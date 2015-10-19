// package name: rmq
package main

/*
#cgo LDFLAGS: -lR ${SRCDIR}/libinterface.a
#cgo CFLAGS: -I${SRCDIR}/../include
#include <string.h>
#include "interface.h"
*/
import "C"
import "fmt"
import "github.com/glycerine/especial"

//export Srv
func Srv(str_ C.SEXP) C.SEXP {

	if especial.Esp() != "Esp() called" {
		panic("especial not linked!")
	}

	if C.TYPEOF(str_) != C.STRSXP {
		fmt.Printf("not a STRXSP! instead: %d, argument to rmq() must be a string to be decoded to its integer constant value in the rmq pkg.\n", C.TYPEOF(str_))
		return C.R_NilValue
	}

	name := C.R_CHAR(C.STRING_ELT(str_, 0))
	namelen := C.strlen(name)
	gname := C.GoString(name)
	fmt.Printf("namelen=%d   length %d. rmq says: Hello '%s'!\n", namelen, len(gname), gname)
	back := C.JasonsLinkeMe()
	fmt.Printf("YYEEEEEE-HAW  rmq JasonsLinkMe() resulted in: %v!\n", back)

	go server_main()

	return C.R_NilValue
}

//export Cli
func Cli(str_ C.SEXP) C.SEXP {

	if C.TYPEOF(str_) != C.STRSXP {
		fmt.Printf("not a STRXSP! instead: %d, argument to rmq() must be a string to be decoded to its integer constant value in the rmq pkg.\n", C.TYPEOF(str_))
		return C.R_NilValue
	}

	name := C.R_CHAR(C.STRING_ELT(str_, 0))
	gname := C.GoString(name)

	fmt.Printf("rmq says: SayBye() sees '%s'.\n", gname)

	go client_main()

	fmt.Printf("rmq says: after launching client_main().\n")

	return C.R_NilValue
}

func main() {
	// We need the main function to make possible
	// CGO compiler to compile the package as C shared library
}
