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
	"encoding/binary"
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

	fmt.Printf("rmq says: after client_main().\n")

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

func timeEncExt(rv reflect.Value) (bs []byte, err error) {
	switch v2 := rv.Interface().(type) {
	case time.Time:
		bs = EncodeTime(v2)
	case *time.Time:
		bs = EncodeTime(*v2)
	default:
		err = fmt.Errorf("unsupported format for time conversion: expecting time.Time; got %T", v2)
	}
	return
}

func timeDecExt(rv reflect.Value, bs []byte) error {
	tt, err := DecodeTime(bs)
	if err == nil {
		*(rv.Interface().(*time.Time)) = tt
	}
	return err
}

// from github.com/ugorji/go/codec/, but exporting the necessary
// DecodeTime and EncodeTime functions which are inexplicable private.
//

//
// Copyright (c) 2012-2015 Ugorji Nwoke. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

var (
	timeDigits = [...]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
)

// EncodeTime encodes a time.Time as a []byte, including
// information on the instant in time and UTC offset.
//
// Format Description
//
//   A timestamp is composed of 3 components:
//
//   - secs: signed integer representing seconds since unix epoch
//   - nsces: unsigned integer representing fractional seconds as a
//     nanosecond offset within secs, in the range 0 <= nsecs < 1e9
//   - tz: signed integer representing timezone offset in minutes east of UTC,
//     and a dst (daylight savings time) flag
//
//   When encoding a timestamp, the first byte is the descriptor, which
//   defines which components are encoded and how many bytes are used to
//   encode secs and nsecs components. *If secs/nsecs is 0 or tz is UTC, it
//   is not encoded in the byte array explicitly*.
//
//       Descriptor 8 bits are of the form `A B C DDD EE`:
//           A:   Is secs component encoded? 1 = true
//           B:   Is nsecs component encoded? 1 = true
//           C:   Is tz component encoded? 1 = true
//           DDD: Number of extra bytes for secs (range 0-7).
//                If A = 1, secs encoded in DDD+1 bytes.
//                    If A = 0, secs is not encoded, and is assumed to be 0.
//                    If A = 1, then we need at least 1 byte to encode secs.
//                    DDD says the number of extra bytes beyond that 1.
//                    E.g. if DDD=0, then secs is represented in 1 byte.
//                         if DDD=2, then secs is represented in 3 bytes.
//           EE:  Number of extra bytes for nsecs (range 0-3).
//                If B = 1, nsecs encoded in EE+1 bytes (similar to secs/DDD above)
//
//   Following the descriptor bytes, subsequent bytes are:
//
//       secs component encoded in `DDD + 1` bytes (if A == 1)
//       nsecs component encoded in `EE + 1` bytes (if B == 1)
//       tz component encoded in 2 bytes (if C == 1)
//
//   secs and nsecs components are integers encoded in a BigEndian
//   2-complement encoding format.
//
//   tz component is encoded as 2 bytes (16 bits). Most significant bit 15 to
//   Least significant bit 0 are described below:
//
//       Timezone offset has a range of -12:00 to +14:00 (ie -720 to +840 minutes).
//       Bit 15 = have\_dst: set to 1 if we set the dst flag.
//       Bit 14 = dst\_on: set to 1 if dst is in effect at the time, or 0 if not.
//       Bits 13..0 = timezone offset in minutes. It is a signed integer in Big Endian format.
//
func EncodeTime(t time.Time) []byte {
	//t := rv.Interface().(time.Time)
	tsecs, tnsecs := t.Unix(), t.Nanosecond()
	var (
		bd   byte
		btmp [8]byte
		bs   [16]byte
		i    int = 1
	)
	l := t.Location()
	if l == time.UTC {
		l = nil
	}
	if tsecs != 0 {
		bd = bd | 0x80
		binary.BigEndian.PutUint64(btmp[:], uint64(tsecs))
		f := pruneSignExt(btmp[:], tsecs >= 0)
		bd = bd | (byte(7-f) << 2)
		copy(bs[i:], btmp[f:])
		i = i + (8 - f)
	}
	if tnsecs != 0 {
		bd = bd | 0x40
		binary.BigEndian.PutUint32(btmp[:4], uint32(tnsecs))
		f := pruneSignExt(btmp[:4], true)
		bd = bd | byte(3-f)
		copy(bs[i:], btmp[f:4])
		i = i + (4 - f)
	}
	if l != nil {
		bd = bd | 0x20
		// Note that Go Libs do not give access to dst flag.
		_, zoneOffset := t.Zone()
		//zoneName, zoneOffset := t.Zone()
		zoneOffset /= 60
		z := uint16(zoneOffset)
		binary.BigEndian.PutUint16(btmp[:2], z)
		// clear dst flags
		bs[i] = btmp[0] & 0x3f
		bs[i+1] = btmp[1]
		i = i + 2
	}
	bs[0] = bd
	return bs[0:i]
}

// DecodeTime decodes a []byte into a time.Time.
func DecodeTime(bs []byte) (tt time.Time, err error) {
	bd := bs[0]
	var (
		tsec  int64
		tnsec uint32
		tz    uint16
		i     byte = 1
		i2    byte
		n     byte
	)
	if bd&(1<<7) != 0 {
		var btmp [8]byte
		n = ((bd >> 2) & 0x7) + 1
		i2 = i + n
		copy(btmp[8-n:], bs[i:i2])
		//if first bit of bs[i] is set, then fill btmp[0..8-n] with 0xff (ie sign extend it)
		if bs[i]&(1<<7) != 0 {
			copy(btmp[0:8-n], bsAll0xff)
			//for j,k := byte(0), 8-n; j < k; j++ {	btmp[j] = 0xff }
		}
		i = i2
		tsec = int64(binary.BigEndian.Uint64(btmp[:]))
	}
	if bd&(1<<6) != 0 {
		var btmp [4]byte
		n = (bd & 0x3) + 1
		i2 = i + n
		copy(btmp[4-n:], bs[i:i2])
		i = i2
		tnsec = binary.BigEndian.Uint32(btmp[:])
	}
	if bd&(1<<5) == 0 {
		tt = time.Unix(tsec, int64(tnsec)).UTC()
		return
	}
	// In stdlib time.Parse, when a date is parsed without a zone name, it uses "" as zone name.
	// However, we need name here, so it can be shown when time is printed.
	// Zone name is in form: UTC-08:00.
	// Note that Go Libs do not give access to dst flag, so we ignore dst bits

	i2 = i + 2
	tz = binary.BigEndian.Uint16(bs[i:i2])
	i = i2
	// sign extend sign bit into top 2 MSB (which were dst bits):
	if tz&(1<<13) == 0 { // positive
		tz = tz & 0x3fff //clear 2 MSBs: dst bits
	} else { // negative
		tz = tz | 0xc000 //set 2 MSBs: dst bits
		//tzname[3] = '-' (TODO: verify. this works here)
	}
	tzint := int16(tz)
	if tzint == 0 {
		tt = time.Unix(tsec, int64(tnsec)).UTC()
	} else {
		// For Go Time, do not use a descriptive timezone.
		// It's unnecessary, and makes it harder to do a reflect.DeepEqual.
		// The Offset already tells what the offset should be, if not on UTC and unknown zone name.
		// var zoneName = timeLocUTCName(tzint)
		tt = time.Unix(tsec, int64(tnsec)).In(time.FixedZone("", int(tzint)*60))
	}
	return
}

func timeLocUTCName(tzint int16) string {
	if tzint == 0 {
		return "UTC"
	}
	var tzname = []byte("UTC+00:00")
	//tzname := fmt.Sprintf("UTC%s%02d:%02d", tzsign, tz/60, tz%60) //perf issue using Sprintf. inline below.
	//tzhr, tzmin := tz/60, tz%60 //faster if u convert to int first
	var tzhr, tzmin int16
	if tzint < 0 {
		tzname[3] = '-' // (TODO: verify. this works here)
		tzhr, tzmin = -tzint/60, (-tzint)%60
	} else {
		tzhr, tzmin = tzint/60, tzint%60
	}
	tzname[4] = timeDigits[tzhr/10]
	tzname[5] = timeDigits[tzhr%10]
	tzname[7] = timeDigits[tzmin/10]
	tzname[8] = timeDigits[tzmin%10]
	return string(tzname)
	//return time.FixedZone(string(tzname), int(tzint)*60)
}

func pruneSignExt(v []byte, pos bool) (n int) {
	if len(v) < 2 {
	} else if pos && v[0] == 0 {
		for ; v[n] == 0 && n+1 < len(v) && (v[n+1]&(1<<7) == 0); n++ {
		}
	} else if !pos && v[0] == 0xff {
		for ; v[n] == 0xff && n+1 < len(v) && (v[n+1]&(1<<7) != 0); n++ {
		}
	}
	return
}

var bsAll0x00 = []byte{0, 0, 0, 0, 0, 0, 0, 0}
var bsAll0xff = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

// end of stuff extracted for ugorji/codec

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
	C.Rf_unprotect(1) // unprotect s before returning it
	return s
}

const FLT_RADIX = 2
const DBL_MANT_DIG = 53

// this is suggested: go:nosplit, because we are doing unsafe pointer arithmetic
// http://stackoverflow.com/questions/32700999/pointer-arithmetic-in-go

//go:nosplit
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

		if len(val) == 0 {
			emptyvec := C.allocVector(C.NILSXP, C.R_xlen_t(0))

			if depth == 0 {
				C.Rf_protect(emptyvec)
			}
			return emptyvec
		}

		if len(val) > 0 {
			first := val[0]
			VPrintf(" ... also at depth %d,   ---> first has type '%T' and value '%v'\n", depth, first, first)

			switch first.(type) {
			case string:
				sxpTy = C.STRSXP

				stringSlice := C.allocVector(sxpTy, C.R_xlen_t(len(val)))
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

				numSlice := C.allocVector(sxpTy, C.R_xlen_t(len(val)))
				C.Rf_protect(numSlice)
				size := unsafe.Sizeof(C.double(0))
				naflag := false
				rmax := int64(C.pow(FLT_RADIX, DBL_MANT_DIG) - 1)
				//VPrintf("rmax = %v\n", rmax) //  rmax = 9007199254740991
				rmin := -rmax
				for i := range val {
					n := val[i].(int64)
					if n < rmin || n > rmax {
						naflag = true
					}

					// this pointer arithmetic wants the //go:nosplit annotation at the top of this function
					*((*C.double)(unsafe.Pointer(uintptr(unsafe.Pointer(C.REAL(numSlice))) + size*uintptr(i)))) = C.double(float64(n))
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

				numSlice := C.allocVector(sxpTy, C.R_xlen_t(len(val)))
				C.Rf_protect(numSlice)
				size := unsafe.Sizeof(C.double(0))

				// unfortunately C.memmove() doesn't work here (I tried). I speculate this is because val[i] is
				// really wrapped in an interface{} rather than being a actual float64. val is an
				// []interface{} after all.
				for i := range val {
					// this pointer arithmetic wants the //go:nosplit annotation at the top of this function
					*((*C.double)(unsafe.Pointer(uintptr(unsafe.Pointer(C.REAL(numSlice))) + size*uintptr(i)))) = C.double(val[i].(float64))
				}
				if depth != 0 {
					C.Rf_unprotect(1) // unprotect for numSlice, now that we are returning it
				}
				return numSlice

			}
		}

		intslice := C.allocVector(sxpTy, C.R_xlen_t(len(val)))
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

	default:
		fmt.Printf("unknown type in type switch, val = %#v\n", val)
	}

	return s
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

	switch C.TYPEOF(s) {
	case C.VECSXP:
		// an R generic vector; e.g list()
		fmt.Printf("encodeRIntoMsgpack sees VECSXP\n")

		// could be a map or a slice. Check out the names.
		rnames := C.Rf_getAttrib(s, C.R_NamesSymbol)
		rnamesLen := int(C.Rf_xlength(rnames))
		fmt.Printf("namesLen = %d\n", rnamesLen)
		myNames := map[string]interface{}{}
		for i := 0; i < rnamesLen; i++ {
			myNames[C.GoString(C.get_string_elt(rnames, C.int(i)))] = toIface(C.VECTOR_ELT(s, C.R_xlen_t(i)))
		}
		fmt.Printf("myNames = '%#v'\n", myNames)

	case C.INTSXP:
		// a vector of int32
		fmt.Printf("encodeRIntoMsgpack sees INTSXP\n")
		// asInteger(x)
	case C.REALSXP:
		// a vector of float64 (numeric)
		fmt.Printf("encodeRIntoMsgpack sees REALSXP\n")
		// asReal(x)
	case C.STRSXP:
		// a vector of string (pointers to charsxp that are interned)
		fmt.Printf("encodeRIntoMsgpack sees STRSXP\n")
		// CHAR(asChar(x))
	case C.CHARSXP:
		// a single string, interned in a global pool for reuse by STRSXP.
		fmt.Printf("encodeRIntoMsgpack sees CHARSXP\n")
	case C.NILSXP:
		// c(); an empty vector
		fmt.Printf("encodeRIntoMsgpack sees NILSXP\n")

	case C.SYMSXP:
		fmt.Printf("encodeRIntoMsgpack sees SYMSXP\n")
	case C.LISTSXP:
		fmt.Printf("encodeRIntoMsgpack sees LISTSXP\n")
	case C.CLOSXP:
		fmt.Printf("encodeRIntoMsgpack sees CLOSXP\n")
	case C.ENVSXP:
		fmt.Printf("encodeRIntoMsgpack sees ENVSXP\n")
	case C.PROMSXP:
		fmt.Printf("encodeRIntoMsgpack sees PROMSXP\n")
	case C.LANGSXP:
		fmt.Printf("encodeRIntoMsgpack sees LANGSXP\n")
	case C.SPECIALSXP:
		fmt.Printf("encodeRIntoMsgpack sees SPECIALSXP\n")
	case C.BUILTINSXP:
		fmt.Printf("encodeRIntoMsgpack sees BUILTINSXP\n")
	case C.LGLSXP:
		fmt.Printf("encodeRIntoMsgpack sees LGLSXP\n")
	case C.CPLXSXP:
		fmt.Printf("encodeRIntoMsgpack sees CPLXSXP\n")
	case C.DOTSXP:
		fmt.Printf("encodeRIntoMsgpack sees DOTSXP\n")
	case C.ANYSXP:
		fmt.Printf("encodeRIntoMsgpack sees ANYSXP\n")
	case C.EXPRSXP:
		fmt.Printf("encodeRIntoMsgpack sees EXPRSXP\n")
	case C.BCODESXP:
		fmt.Printf("encodeRIntoMsgpack sees BCODESXP\n")
	case C.EXTPTRSXP:
		fmt.Printf("encodeRIntoMsgpack sees EXTPTRSXP\n")
	case C.WEAKREFSXP:
		fmt.Printf("encodeRIntoMsgpack sees WEAKREFSXP\n")
	case C.S4SXP:
		fmt.Printf("encodeRIntoMsgpack sees S4SXP\n")
	case C.RAWSXP:
		fmt.Printf("encodeRIntoMsgpack sees RAWSXP\n")
	default:
		fmt.Printf("encodeRIntoMsgpack sees <unknown>\n")
	}
	fmt.Printf("... warning: encodeRIntoMsgpack() ignoring this input.\n")

	return nil
}
