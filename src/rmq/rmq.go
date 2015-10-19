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

//export ListenAndServe
func ListenAndServe(addr_ C.SEXP, handler_ C.SEXP, rho_ C.SEXP) C.SEXP {

	if C.TYPEOF(addr_) != C.STRSXP {
		fmt.Printf("addr is not a string (STRXSP; instead it is: %d)! addr argument to ListenAndServe() must be a string of form 'ip:port'\n", C.TYPEOF(addr_))
		return C.R_NilValue
	}

	caddr := C.R_CHAR(C.STRING_ELT(addr_, 0))
	addr := C.GoString(caddr)
	fmt.Printf("ListenAndServe listening on address '%s'...\n", addr)

	//msglen := 0
	//var evalres C.SEXP
	//var R_fcall, msg C.SEXP
	if 0 == int(C.isFunction(handler_)) { // 0 is false
		C.ReportErrorToR_NoReturn(C.CString("‘handler’ must be a function"))
		return C.R_NilValue
	}

	if 0 == int(C.isEnvironment(rho_)) { // 0 is false
		C.ReportErrorToR_NoReturn(C.CString("‘rho’ should be an environment"))
		return C.R_NilValue
	}

	/*
	   	for {
	       // blocking read
	       //REprintf("nnListenAndServe: just prior to blocking waiting for msg\n");
	       msglen = nn_recv(INTEGER(socket_)[0], &buf, NN_MSG, 0);
	       if (msglen < 0) {
	           error("error in nnRecv(): '%s'.\n", nn_strerror(nn_errno()));
	           return R_NilValue;
	       }

	       PROTECT(msg = allocVector(RAWSXP,msglen));
	       memcpy(RAW(msg),buf,msglen);
	       if (nn_freemsg(buf) < 0) {
	         error("bad buf: message pointer is invalid, in nnListenAndServer() loop.\n");
	       }

	       // put msg into env that handler_ is called with.
	       //defineVar(install("msg"), msg, rho_);

	       // evaluate
	       //REprintf("nnListenAndServe: got msg, just prior to eval.\n");
	       PROTECT(R_fcall = lang2(handler_, msg));
	       PROTECT(evalres = eval(R_fcall, rho_));
	       //REprintf("nnListenAndServe: evalres = %p.\n",evalres);
	       //REprintf("nnListenAndServe: done with eval.\n");
	       UNPROTECT(3);
	      }

	*/
	return C.R_NilValue
}

//export Srv
func Srv(str_ C.SEXP) C.SEXP {

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

	go StartServer()
	//go server_main()

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

	go StartClient()

	fmt.Printf("rmq says: after launching client_main().\n")

	return C.R_NilValue
}

func main() {
	// We need the main function to make possible
	// CGO compiler to compile the package as C shared library
}
