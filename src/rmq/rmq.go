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

func decodeMsgpackToR(reply []byte) C.SEXP {

	var mh codec.MsgpackHandle

	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	var r interface{}

	// configure extensions
	// e.g. for msgpack, define functions and enable Time support for tag 1
	mh.AddExt(reflect.TypeOf(time.Time{}), 1, timeEncExt, timeDecExt)
	mh.RawToString = true
	mh.WriteExt = true
	mh.SignedInteger = true

	err := codec.NewDecoderBytes(reply, &mh).Decode(&r)
	panicOn(err)
	fmt.Printf("decoded type : %T\n", r)
	fmt.Printf("decoded value: %v\n", r)

	s := C.allocList(1)
	C.Rf_protect(s)

	switch val := r.(type) {
	default:
		fmt.Printf("unknown type in type switch, val = %#v\n", val)
	}
	/*
		rawmsg := C.allocVector(C.RAWSXP, C.R_xlen_t(len(reply)))
		C.Rf_protect(rawmsg)
		C.memcpy(unsafe.Pointer(C.RAW(rawmsg)), unsafe.Pointer(&reply[0]), C.size_t(len(reply)))
		C.Rf_unprotect(1)
	*/

	return s
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
