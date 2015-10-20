// package name: rmq
package main

/*
#cgo LDFLAGS: -lR ${SRCDIR}/libinterface.a
#cgo CFLAGS: -I${SRCDIR}/../include
#include <string.h>
#include "interface.h"
*/
import "C"

//go:generate msgp

import (
	"fmt"
	"net/http"
	"net/url"
	"unsafe"

	"github.com/gorilla/websocket"
)

type Subload struct {
	A string
	B int
}

type Payload struct {
	Sub Subload
	D   string
	E   int
}

var addr = "localhost:8081"

var R_serialize_fun C.SEXP

//export ListenAndServe
func ListenAndServe(addr_ C.SEXP, handler_ C.SEXP, rho_ C.SEXP) C.SEXP {

	if C.TYPEOF(addr_) != C.STRSXP {
		fmt.Printf("addr is not a string (STRXSP; instead it is: %d)! addr argument to ListenAndServe() must be a string of form 'ip:port'\n", C.TYPEOF(addr_))
		return C.R_NilValue
	}

	//msglen := 0
	if 0 == int(C.isFunction(handler_)) { // 0 is false
		C.ReportErrorToR_NoReturn(C.CString("‘handler’ must be a function"))
		return C.R_NilValue
	}

	if rho_ != nil && rho_ != C.R_NilValue {
		if 0 == int(C.isEnvironment(rho_)) { // 0 is false
			C.ReportErrorToR_NoReturn(C.CString("‘rho’ should be an environment"))
			return C.R_NilValue
		}
	}

	caddr := C.R_CHAR(C.STRING_ELT(addr_, 0))
	addr := C.GoString(caddr)
	fmt.Printf("ListenAndServe listening on address '%s'...\n", addr)

	webSockHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", 404)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "Method not allowed, only GET allowed.", 405)
			return
		}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Print("websocket handler upgrade error:", err)
			return
		}
		defer c.Close()

		mt, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read error: ", err)
			return
		}

		// make the call, and get a response
		msglen := len(message)
		rawmsg := C.allocVector(C.RAWSXP, C.R_xlen_t(msglen))
		C.Rf_protect(rawmsg)
		C.memcpy(unsafe.Pointer(C.RAW(rawmsg)), unsafe.Pointer(&message[0]), C.size_t(msglen))

		// put msg into env that handler_ is called with.
		C.defineVar(C.install(C.CString("msg")), rawmsg, rho_)

		R_serialize_fun = C.findVar(C.install(C.CString("serialize")), C.R_GlobalEnv)

		// evaluate
		C.PrintToR(C.CString("listenAndServe: stuffed msg into env rho_.\n"))
		//R_fcall := C.lang3(handler_, rawmsg, C.R_NilValue)
		R_fcall := C.lang3(R_serialize_fun, rawmsg, C.R_NilValue)
		C.Rf_protect(R_fcall)
		C.PrintToR(C.CString("listenAndServe: got msg, just prior to eval.\n"))
		evalres := C.eval(R_fcall, rho_)
		C.Rf_protect(evalres)

		C.PrintToR(C.CString("listenAndServe: after eval.\n"))
		/*
			var s, t C.SEXP
			s = C.allocList(3)
			t = s
			C.Rf_protect(t)
			C.SetTypeToLANGSXP(&s)
			//C.SETCAR(t, R_fcall)
			C.SETCAR(t, handler_)
			t = C.CDR(t)
			C.SETCAR(t, rawmsg)

			evalres := C.eval(s, rho_)
			C.Rf_protect(evalres)
		*/
		C.PrintToR(C.CString("nnListenAndServe: done with eval.\n"))

		if C.TYPEOF(evalres) != C.RAWSXP {
			fmt.Printf("rats! handler result was not RAWSXP raw bytes!\n")
		} else {

			//fmt.Printf("recv: %s\n", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				fmt.Println("write error: ", err)
			}
		}
		C.Rf_unprotect(3)

	} // end handler func

	http.HandleFunc("/", webSockHandler)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}

	return C.R_NilValue
}

//export Srv
func Srv(str_ C.SEXP) C.SEXP {

	if C.TYPEOF(str_) != C.STRSXP {
		fmt.Printf("not a STRXSP! instead: %d, argument to rmq() must be a string to be decoded to its integer constant value in the rmq pkg.\n", C.TYPEOF(str_))
		return C.R_NilValue
	}

	name := C.R_CHAR(C.STRING_ELT(str_, 0))
	gname := C.GoString(name)
	fmt.Printf("rmq says: Hello '%s'!\n", gname)

	//go StartServer()
	go server_main()

	fmt.Printf("\n  after gorilla webserver on '%s' launched.\n", addr)

	return C.R_NilValue
}

//export Cli
func Cli(str_ C.SEXP) C.SEXP {

	if C.TYPEOF(str_) != C.STRSXP {
		fmt.Printf("not a STRXSP! instead: %d, argument to rmq() must be a string to be decoded to its integer constant value in the rmq pkg.\n", C.TYPEOF(str_))
		return C.R_NilValue
	}

	name := C.R_CHAR(C.STRING_ELT(str_, 0))
	msg := C.GoString(name)

	//fmt.Printf("rmq says: client sees '%s'.\n", msg)

	reply := client_main([]byte(msg))

	fmt.Printf("rmq says: after client_main().\n")

	if len(reply) == 0 {
		return C.R_NilValue
	}

	if len(reply) > 0 {
		rawmsg := C.allocVector(C.RAWSXP, C.R_xlen_t(len(reply)))
		C.Rf_protect(rawmsg)
		C.memcpy(unsafe.Pointer(C.RAW(rawmsg)), unsafe.Pointer(&reply[0]), C.size_t(len(reply)))
		C.Rf_unprotect(1)

		return rawmsg
	}
	return C.R_NilValue
}

//export Callcount
func Callcount() C.SEXP {
	fmt.Printf("count = %d\n", count)
	return C.R_NilValue
}

func main() {
	// We need the main function to make possible
	// CGO compiler to compile the package as C shared library
}

var c *websocket.Conn
var count int

func client_main(msg []byte) []byte {

	var err error
	if c == nil {
		u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
		fmt.Printf("connecting to %s", u.String())

		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			fmt.Println("dial:", err)
			c = nil
			return []byte{}
		}
	}

	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		fmt.Println("write err:", err)
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		fmt.Println("read err:", err)
	}
	if false {
		fmt.Printf("recv: %s\n", message)
	}
	count++

	return message
}

// server

var upgrader = websocket.Upgrader{} // use default options

var data Payload
var dataBytes []byte

func echo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed, only GET allowed.", 405)
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read error: ", err)
			break
		}
		if false {
			fmt.Printf("recv: %s\n", message)
		}
		err = c.WriteMessage(mt, dataBytes)
		if err != nil {
			fmt.Println("write error: ", err)
			break
		}
	}
}

func server_main() {

	data = Payload{
		Sub: Subload{
			A: "hi",
			B: 43,
		},
		D: "hello",
		E: 32,
	}

	var err error
	dataBytes, err = data.MarshalMsg(nil)
	panicOn(err)

	fmt.Printf("data = %#v\n", data)

	http.HandleFunc("/", echo)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}
