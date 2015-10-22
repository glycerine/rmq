/* package name: rmq
//
// Copyright 2015 Jason E. Aten <j.e.aten -a-t- g-m-a-i-l dot c-o-m>
// License: Apache 2.0
*/
package main

/*
#cgo LDFLAGS: -L/usr/local/lib64/R/lib -lm -lR ${SRCDIR}/libinterface.a
#cgo CFLAGS: -I${SRCDIR}/../include
#include <string.h>
#include "interface.h"
*/
import "C"

//go:generate msgp

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/shurcooL/go-goon"
	"github.com/ugorji/go/codec"
)

// inside test struct for checking serialization
type Subload struct {
	A string
	B int
	F []float64
}

// outside test struct for checking serialization
type Payload struct {
	Sub  Subload
	D    []string
	E    []int32
	G    []float64
	Blob []byte
}

var DefaultAddr = "localhost:8081"

var R_serialize_fun C.SEXP

func getAddr(addr_ C.SEXP) (*net.TCPAddr, error) {

	if C.TYPEOF(addr_) != C.STRSXP {
		fmt.Printf("addr is not a string (STRXSP; instead it is: %d)! addr argument to ListenAndServe() must be a string of form 'ip:port'\n", C.TYPEOF(addr_))
		return nil, fmt.Errorf("addr is not string type")
	}

	caddr := C.R_CHAR(C.STRING_ELT(addr_, 0))
	addr := C.GoString(caddr)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil || tcpAddr == nil {
		return nil, fmt.Errorf("getAddr() error: address '%s' could not be parsed by net.ResolveTCPAddr(): error: '%s'", addr, err)
	}
	return tcpAddr, nil
}

//export ListenAndServe
func ListenAndServe(addr_ C.SEXP, handler_ C.SEXP, rho_ C.SEXP) C.SEXP {

	addr, err := getAddr(addr_)

	if err != nil {
		C.ReportErrorToR_NoReturn(C.CString(err.Error()))
		return C.R_NilValue
	}

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

	fmt.Printf("ListenAndServe listening on address '%s'...\n", addr)

	// one problem  when acting as a web server:
	// webSockHandler will be run on a separate goroutine and this will surely
	// be a separate thread--distinct from the R callback thread. This is a
	// problem because if we call back into R from the goroutine thread
	// instead of R's thread, R will see the small stack and freak out.
	//
	// So: we'll use a channel to send the request to the main C/R thread
	// for call back into R. The *[]byte passed on these channels represent
	// msgpack serialized R objects.
	requestToRCh := make(chan *[]byte)
	replyFromRCh := make(chan *[]byte)
	reqStopCh := make(chan bool)
	doneCh := make(chan bool)

	ctrlC_Chan := make(chan os.Signal, 1)
	//	signal.Notify(ctrlC_Chan, os.Interrupt)

	webSockHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", 404)
			return
		}

		fmt.Printf("\n\n  webSockHandler() sees request r = \n")
		goon.Dump(r)
		fmt.Printf("\n\n")

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

		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read error: ", err)
			return
		}

		requestToRCh <- &message
		reply := <-replyFromRCh

		err = c.WriteMessage(websocket.BinaryMessage, *reply)
		if err != nil {
			fmt.Println("write error: ", err)
		}
	} // end webSockHandler

	http.HandleFunc("/", webSockHandler)
	go func() {
		err := http.ListenAndServe(addr.String(), nil)
		if err != nil {
			fmt.Println("ListenAndServe error: ", err)
		}
	}()

	for {
		select {
		case msgpackRequest := <-requestToRCh:

			rRequest := decodeMsgpackToR(*msgpackRequest)
			C.Rf_protect(rRequest)

			// Call into the R handler_ function, and get its reply.
			R_fcall := C.lang2(handler_, rRequest)
			C.Rf_protect(R_fcall)
			C.PrintToR(C.CString("listenAndServe: got msg, just prior to eval.\n"))
			evalres := C.eval(R_fcall, rho_)
			C.Rf_protect(evalres)
			C.PrintToR(C.CString("listenAndServe: after eval.\n"))

			// send back the reply, first converting to msgpack
			reply := encodeRIntoMsgpack(evalres)
			C.Rf_unprotect(3)
			replyFromRCh <- &reply

		case <-ctrlC_Chan:
			// ctrl-c pressed, return to user.
			close(doneCh)
			return C.R_NilValue

		case <-reqStopCh:
			// possibly fired by ctrl-c? not sure who should close(reqStopCh).
			close(doneCh)
			return C.R_NilValue
		}
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

	fmt.Printf("\n  after gorilla webserver launched.\n")

	return C.R_NilValue
}

// RmqWebsocketCall() is the client part that talks to
// the server part waiting in ListenAndServe().

//export RmqWebsocketCall
func RmqWebsocketCall(addr_ C.SEXP, msg_ C.SEXP) C.SEXP {

	addr, err := getAddr(addr_)

	if err != nil {
		C.ReportErrorToR_NoReturn(C.CString(err.Error()))
		return C.R_NilValue
	}

	// marshall msg_ into msgpack []byte
	msgpackRequest := encodeRIntoMsgpack(msg_)

	reply := client_main(addr, msgpackRequest)
	VPrintf("rmq says: after client_main().\n")

	if len(reply) == 0 {
		return C.R_NilValue
	}
	if len(reply) > 0 {
		return decodeMsgpackToR(reply)
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

func client_main(addr *net.TCPAddr, msg []byte) []byte {

	var err error
	if c == nil {
		u := url.URL{Scheme: "ws", Host: addr.String(), Path: "/"}
		fmt.Printf("connecting to %s", u.String())

		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			fmt.Println("dial error:", err)
			c = nil
			return []byte{}
		}
	}

	err = c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		fmt.Println("write err:", err)
	}

	_, replyBytes, err := c.ReadMessage()
	if err != nil {
		fmt.Println("read err:", err)
	}
	if false {
		// debug
		fmt.Printf("recv: %s\n", replyBytes)
	}
	count++

	return replyBytes
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
		Blob: []byte{0xff, 0xf0, 0x06},
		Sub: Subload{
			A: "hi",
			B: 4611686018427387904,
			F: []float64{1.5, 3.4},
		},
		D: []string{"hello", "world"},
		E: []int32{32, 17},
		G: []float64{-10.5},
	}

	// we have to fix the encoding we return, as go will
	// randomize out map ordering access on passing through a map.
	dataBytes = []byte{0x85, 0xa4, 0x42, 0x6c, 0x6f, 0x62, 0xc4, 0x03, 0xff, 0xf0, 0x06, 0xa1, 0x44, 0x92, 0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xa5, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0xa1, 0x45, 0x92, 0xcb, 0x40, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xcb, 0x40, 0x31, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa1, 0x47, 0x91, 0xcb, 0xc0, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa3, 0x53, 0x75, 0x62, 0x83, 0xa1, 0x41, 0x91, 0xa2, 0x68, 0x69, 0xa1, 0x42, 0x91, 0xcb, 0x43, 0xd0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa1, 0x46, 0x92, 0xcb, 0x3f, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xcb, 0x40, 0x0b, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}

	//	var err error
	//	dataBytes, err = data.MarshalMsg(nil)
	//	panicOn(err)

	fmt.Printf("data = %#v\n", data)
	fmt.Printf("dataBytes = %#v\n", dataBytes)

	// create a fresh mux to avoid errors from reuse.
	mux := http.NewServeMux()
	mux.HandleFunc("/", echo)
	server := &http.Server{Addr: DefaultAddr, Handler: mux}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}

type MsgpackHelper struct {
	initialized bool
	mh          codec.MsgpackHandle
}

func (m *MsgpackHelper) init() {
	if m.initialized {
		return
	}

	m.mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	// configure extensions
	// e.g. for msgpack, define functions and enable Time support for tag 1
	//does this make a differenece? m.mh.AddExt(reflect.TypeOf(time.Time{}), 1, timeEncExt, timeDecExt)
	m.mh.RawToString = true
	m.mh.WriteExt = true
	m.mh.SignedInteger = true
	m.mh.Canonical = true // sort maps before writing them

	m.initialized = true
}

var h MsgpackHelper

func init() {
	h.init()
}

func decodeMsgpackToR(reply []byte) C.SEXP {

	h.init()
	var r interface{}

	decoder := codec.NewDecoderBytes(reply, &h.mh)
	err := decoder.Decode(&r)
	panicOn(err)

	VPrintf("decoded type : %T\n", r)
	VPrintf("decoded value: %#v\n", r)

	s := decodeHelper(r, 0)
	if s != C.R_NilValue {
		C.Rf_unprotect(1) // unprotect s before returning it
	}
	return s
}

const FLT_RADIX = 2
const DBL_MANT_DIG = 53

func decodeHelper(r interface{}, depth int) (s C.SEXP) {

	VPrintf("decodeHelper() at depth %d, decoded type is %T\n", depth, r)
	switch val := r.(type) {
	case string:
		VPrintf("depth %d found string case: val = %#v\n", depth, val)
		return C.Rf_mkString(C.CString(val))

	case int:
		VPrintf("depth %d found int case: val = %#v\n", depth, val)
		return C.Rf_ScalarReal(C.double(float64(val)))

	case int32:
		VPrintf("depth %d found int32 case: val = %#v\n", depth, val)
		return C.Rf_ScalarReal(C.double(float64(val)))

	case int64:
		VPrintf("depth %d found int64 case: val = %#v\n", depth, val)
		return C.Rf_ScalarReal(C.double(float64(val)))

	case []interface{}:
		VPrintf("depth %d found []interface{} case: val = %#v\n", depth, val)

		var sxpTy C.SEXPTYPE = C.VECSXP

		lenval := len(val)
		if lenval == 0 {
			emptyvec := C.allocVector(C.NILSXP, C.R_xlen_t(0))

			if depth == 0 {
				C.Rf_protect(emptyvec)
			}
			return emptyvec
		}

		if lenval > 0 {
			first := val[0]
			VPrintf(" ... also at depth %d,   ---> first has type '%T' and value '%v'\n", depth, first, first)

			switch first.(type) {
			case string:
				sxpTy = C.STRSXP

				stringSlice := C.allocVector(sxpTy, C.R_xlen_t(lenval))
				C.Rf_protect(stringSlice)
				for i := range val {
					C.SET_STRING_ELT(stringSlice, C.R_xlen_t(i), C.mkChar(C.CString(val[i].(string))))
				}
				if depth != 0 {
					C.Rf_unprotect(1) // unprotect for stringSlice, now that we are returning it
				}
				return stringSlice

			case int64:
				// we can only realistically hope to preserve 53 bits worth here.
				// todo? unless... can we require bit64 package be available somehow?
				sxpTy = C.REALSXP

				numSlice := C.allocVector(sxpTy, C.R_xlen_t(lenval))
				C.Rf_protect(numSlice)
				size := unsafe.Sizeof(C.double(0))
				naflag := false
				rmax := int64(C.pow(FLT_RADIX, DBL_MANT_DIG) - 1)
				//VPrintf("rmax = %v\n", rmax) //  rmax = 9007199254740991
				rmin := -rmax
				ptrNumSlice := unsafe.Pointer(C.REAL(numSlice))
				var ui uintptr
				var rhs C.double
				for i := range val {
					n := val[i].(int64)
					fmt.Printf("n = %d, rmax = %d, n > rmax = %v\n", n, rmax, n > rmax)

					if n < rmin || n > rmax {
						naflag = true
					}

					ui = uintptr(i)
					rhs = C.double(float64(n))
					// Try to avoid any gc activity by avoiding conversions
					// and hence do pointer arithmetic all at once in one expression. See
					// https://github.com/golang/go/issues/8994 for discussion.
					*((*C.double)(unsafe.Pointer(uintptr(ptrNumSlice) + size*ui))) = rhs
				}
				if naflag {
					C.WarnAndContinue(C.CString("integer precision lost while converting to double"))
				}

				if depth != 0 {
					C.Rf_unprotect(1) // unprotect for numSlice, now that we are returning it
				}
				return numSlice

			case float64:
				sxpTy = C.REALSXP

				numSlice := C.allocVector(sxpTy, C.R_xlen_t(lenval))
				C.Rf_protect(numSlice)
				size := unsafe.Sizeof(C.double(0))

				// unfortunately C.memmove() doesn't work here (I tried). I speculate this is because val[i] is
				// really wrapped in an interface{} rather than being a actual float64. val *is* an
				// []interface{} after all.
				var rhs C.double
				ptrNumSlice := unsafe.Pointer(C.REAL(numSlice))
				for i := range val {
					rhs = C.double(val[i].(float64))
					*((*C.double)(unsafe.Pointer(uintptr(ptrNumSlice) + size*uintptr(i)))) = rhs
				}
				if depth != 0 {
					C.Rf_unprotect(1) // unprotect for numSlice, now that we are returning it
				}
				return numSlice

			}
		}

		intslice := C.allocVector(sxpTy, C.R_xlen_t(lenval))
		C.Rf_protect(intslice)
		for i := range val {
			C.SET_VECTOR_ELT(intslice, C.R_xlen_t(i), decodeHelper(val[i], depth+1))
		}
		if depth != 0 {
			C.Rf_unprotect(1) // unprotect for intslice, now that we are returning it
		}
		return intslice

	case map[string]interface{}:

		s = C.allocVector(C.VECSXP, C.R_xlen_t(len(val)))
		if depth == 0 {
			// only protect the top parent of the returned value, recursively
			// geneated are transitively protected by their parent.
			C.Rf_protect(s)
		}
		names := C.allocVector(C.VECSXP, C.R_xlen_t(len(val)))
		C.Rf_protect(names)

		VPrintf("depth %d found map[string]interface case: val = %#v\n", depth, val)
		sortedMapKey, sortedMapVal := makeSortedSlicesFromMap(val)
		for i := range sortedMapKey {

			ele := decodeHelper(sortedMapVal[i], depth+1)
			C.Rf_protect(ele)
			C.SET_VECTOR_ELT(s, C.R_xlen_t(i), ele)
			C.Rf_unprotect(1) // unprotect for ele, now that it is safely inside s.

			ksexpString := C.Rf_mkString(C.CString(sortedMapKey[i]))
			C.SET_VECTOR_ELT(names, C.R_xlen_t(i), ksexpString)
		}
		C.setAttrib(s, C.R_NamesSymbol, names)
		C.Rf_unprotect(1) // unprotect for names, now that it is attached to s.

	case []byte:
		VPrintf("depth %d found []byte case: val = %#v\n", depth, val)

		rawmsg := C.allocVector(C.RAWSXP, C.R_xlen_t(len(val)))

		if depth == 0 {
			C.Rf_protect(rawmsg)
		}
		C.memcpy(unsafe.Pointer(C.RAW(rawmsg)), unsafe.Pointer(&val[0]), C.size_t(len(val)))
		return rawmsg

	case nil:
		return C.R_NilValue

	default:
		fmt.Printf("unknown type in type switch, val = %#v.  type = %T.\n", val, val)
	}

	return s
}

//export FromMsgpack
func FromMsgpack(s C.SEXP) C.SEXP {
	// starting from within R, we convert a raw byte vector into R structures.

	// s must be a RAWSXP
	if C.TYPEOF(s) != C.RAWSXP {
		C.ReportErrorToR_NoReturn(C.CString("from.msgpack(x) requires x be a RAW vector of bytes."))
	}

	n := int(C.Rf_xlength(s))
	if n == 0 {
		return C.R_NilValue
	}
	bytes := make([]byte, n)

	C.memcpy(unsafe.Pointer(&bytes[0]), unsafe.Pointer(C.RAW(s)), C.size_t(n))

	return decodeMsgpackToR(bytes)
}

//export ToMsgpack
func ToMsgpack(s C.SEXP) C.SEXP {
	byteSlice := encodeRIntoMsgpack(s)

	if len(byteSlice) == 0 {
		return C.R_NilValue
	}
	rawmsg := C.allocVector(C.RAWSXP, C.R_xlen_t(len(byteSlice)))
	C.Rf_protect(rawmsg)
	C.memcpy(unsafe.Pointer(C.RAW(rawmsg)), unsafe.Pointer(&byteSlice[0]), C.size_t(len(byteSlice)))
	C.Rf_unprotect(1)

	return rawmsg
}

func encodeRIntoMsgpack(s C.SEXP) []byte {
	iface := toIface(s)

	VPrintf("toIface returned: '%#v'\n", iface)

	if iface == nil {
		return []byte{}
	}

	h.init()

	var w bytes.Buffer
	enc := codec.NewEncoder(&w, &h.mh)
	err := enc.Encode(&iface)
	panicOn(err)

	return w.Bytes()
}

func toIface(s C.SEXP) interface{} {
	// generate a go map or slice or scalar value, then encode it

	n := int(C.Rf_xlength(s))
	if n == 0 {
		return nil // drops type info. Meh.
	}

	switch C.TYPEOF(s) {
	case C.VECSXP:
		// an R generic vector; e.g list()
		VPrintf("encodeRIntoMsgpack sees VECSXP\n")

		// could be a map or a slice. Check out the names.
		rnames := C.Rf_getAttrib(s, C.R_NamesSymbol)
		rnamesLen := int(C.Rf_xlength(rnames))
		VPrintf("namesLen = %d\n", rnamesLen)
		if rnamesLen > 0 {
			myMap := map[string]interface{}{}
			for i := 0; i < rnamesLen; i++ {
				myMap[C.GoString(C.get_string_elt(rnames, C.int(i)))] = toIface(C.VECTOR_ELT(s, C.R_xlen_t(i)))
			}
			VPrintf("VECSXP myMap = '%#v'\n", myMap)
			return myMap
		} else {
			// else: no names, so we treat it as an array instead of as a map
			mySlice := make([]interface{}, n)
			for i := 0; i < n; i++ {
				mySlice[i] = toIface(C.VECTOR_ELT(s, C.R_xlen_t(i)))
			}
			VPrintf("VECSXP mySlice = '%#v'\n", mySlice)
			return mySlice
		}

	case C.REALSXP:
		// a vector of float64 (numeric)
		VPrintf("encodeRIntoMsgpack sees REALSXP\n")
		mySlice := make([]float64, n)
		for i := 0; i < n; i++ {
			mySlice[i] = float64(C.get_real_elt(s, C.int(i)))
		}
		VPrintf("VECSXP mySlice = '%#v'\n", mySlice)
		return mySlice

	case C.INTSXP:
		// a vector of int32
		VPrintf("encodeRIntoMsgpack sees INTSXP\n")
		mySlice := make([]int, n)
		for i := 0; i < n; i++ {
			mySlice[i] = int(C.get_int_elt(s, C.int(i)))
		}
		VPrintf("INTSXP mySlice = '%#v'\n", mySlice)
		return mySlice

	case C.RAWSXP:
		VPrintf("encodeRIntoMsgpack sees RAWSXP\n")
		mySlice := make([]byte, n)
		C.memcpy(unsafe.Pointer(&mySlice[0]), unsafe.Pointer(C.RAW(s)), C.size_t(n))
		VPrintf("RAWSXP mySlice = '%#v'\n", mySlice)
		return mySlice

	case C.STRSXP:
		// a vector of string (pointers to charsxp that are interned)
		VPrintf("encodeRIntoMsgpack sees STRSXP\n")
		mySlice := make([]string, n)
		for i := 0; i < n; i++ {
			mySlice[i] = C.GoString(C.get_string_elt(s, C.int(i)))
		}
		VPrintf("STRSXP mySlice = '%#v'\n", mySlice)
		return mySlice

	case C.NILSXP:
		// c(); an empty vector
		VPrintf("encodeRIntoMsgpack sees NILSXP\n")
		return nil

	case C.CHARSXP:
		// a single string, interned in a global pool for reuse by STRSXP.
		VPrintf("encodeRIntoMsgpack sees CHARSXP\n")
	case C.SYMSXP:
		VPrintf("encodeRIntoMsgpack sees SYMSXP\n")
	case C.LISTSXP:
		VPrintf("encodeRIntoMsgpack sees LISTSXP\n")
	case C.CLOSXP:
		VPrintf("encodeRIntoMsgpack sees CLOSXP\n")
	case C.ENVSXP:
		VPrintf("encodeRIntoMsgpack sees ENVSXP\n")
	case C.PROMSXP:
		VPrintf("encodeRIntoMsgpack sees PROMSXP\n")
	case C.LANGSXP:
		VPrintf("encodeRIntoMsgpack sees LANGSXP\n")
	case C.SPECIALSXP:
		VPrintf("encodeRIntoMsgpack sees SPECIALSXP\n")
	case C.BUILTINSXP:
		VPrintf("encodeRIntoMsgpack sees BUILTINSXP\n")
	case C.LGLSXP:
		VPrintf("encodeRIntoMsgpack sees LGLSXP\n")
	case C.CPLXSXP:
		VPrintf("encodeRIntoMsgpack sees CPLXSXP\n")
	case C.DOTSXP:
		VPrintf("encodeRIntoMsgpack sees DOTSXP\n")
	case C.ANYSXP:
		VPrintf("encodeRIntoMsgpack sees ANYSXP\n")
	case C.EXPRSXP:
		VPrintf("encodeRIntoMsgpack sees EXPRSXP\n")
	case C.BCODESXP:
		VPrintf("encodeRIntoMsgpack sees BCODESXP\n")
	case C.EXTPTRSXP:
		VPrintf("encodeRIntoMsgpack sees EXTPTRSXP\n")
	case C.WEAKREFSXP:
		VPrintf("encodeRIntoMsgpack sees WEAKREFSXP\n")
	case C.S4SXP:
		VPrintf("encodeRIntoMsgpack sees S4SXP\n")
	default:
		VPrintf("encodeRIntoMsgpack sees <unknown>\n")
	}
	VPrintf("... warning: encodeRIntoMsgpack() ignoring this input.\n")

	return nil
}

//msgp:ignore mapsorter KiSlice

type mapsorter struct {
	key   string
	iface interface{}
}

type KiSlice []*mapsorter

func (a KiSlice) Len() int           { return len(a) }
func (a KiSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a KiSlice) Less(i, j int) bool { return a[i].key < a[j].key }

func makeSortedSlicesFromMap(m map[string]interface{}) ([]string, []interface{}) {
	key := make([]string, len(m))
	val := make([]interface{}, len(m))
	so := make(KiSlice, 0)
	for k, i := range m {
		so = append(so, &mapsorter{key: k, iface: i})
	}
	sort.Sort(so)
	for i := range so {
		key[i] = so[i].key
		val[i] = so[i].iface
	}
	return key, val
}
