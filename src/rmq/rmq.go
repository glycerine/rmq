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
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/ugorji/go/codec"
)

type Subload struct {
	A string
	B int
	F []float64
}

type Payload struct {
	Sub  Subload
	D    []string
	E    []int32
	G    []float64
	Blob []byte
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

		// todo: callbacks to R functions here not working. don't really need them if R always acts as a client instead.

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

func client_main(msg []byte) []byte {

	var err error
	if c == nil {
		u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
		fmt.Printf("connecting to %s", u.String())

		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			fmt.Println("dial error:", err)
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
		Blob: []byte{0xff, 0xf0, 0x06},
		Sub: Subload{
			A: "hi",
			B: 4611686018427387904,
			F: []float64{1.5, 3.4},
		},
		D: []string{"hello", "world"},
		E: []int32{32, 17},
		G: []float64{},
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
	m.mh.AddExt(reflect.TypeOf(time.Time{}), 1, timeEncExt, timeDecExt)
	m.mh.RawToString = true
	m.mh.WriteExt = true
	m.mh.SignedInteger = true
	m.mh.Canonical = true // sort maps before writing them

	m.initialized = true
}

var h MsgpackHelper

func decodeMsgpackToR(reply []byte) C.SEXP {

	h.init()
	var r interface{}

	decoder := codec.NewDecoderBytes(reply, &h.mh)
	err := decoder.Decode(&r)
	panicOn(err)

	VPrintf("decoded type : %T\n", r)
	VPrintf("decoded value: %v\n", r)

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
				fmt.Printf("n = %d, rmax = %d, n > rmax = %v\n", n, rmax, n > rmax)
				for i := range val {
					n := val[i].(int64)
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
		i := 0
		for k, v := range val {

			ele := decodeHelper(v, depth+1)
			C.Rf_protect(ele)
			C.SET_VECTOR_ELT(s, C.R_xlen_t(i), ele)
			C.Rf_unprotect(1) // unprotect for ele, now that it is safely inside s.

			ksexpString := C.Rf_mkString(C.CString(k))
			C.SET_VECTOR_ELT(names, C.R_xlen_t(i), ksexpString)
			i++
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
