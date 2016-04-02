package main

/*
#cgo LDFLAGS: -L/usr/local/lib64/R/lib -lm -lR ${SRCDIR}/libinterface.a
#cgo CFLAGS: -I${SRCDIR}/../include
#include <string.h>
#include <signal.h>
#include "interface.h"
*/
import "C"

import (
	"fmt"
	tf "github.com/glycerine/tmframe"
	"io"
	"os"
	"unsafe"
	//"time"
)

//export ReadTmFrame
//
// ReadTmFrame reads a TMFRAME file and returns the frames as
// an R dataframe.
//
func ReadTmFrame(path_ C.SEXP) C.SEXP {

	// s must be a RAWSXP
	if C.TYPEOF(path_) != C.STRSXP {
		C.ReportErrorToR_NoReturn(C.CString("ReadTmFrame() error: path is not a string path to TMFRAME file."))
	}

	cpath := C.R_CHAR(C.STRING_ELT(path_, 0))
	path := C.GoString(cpath)

	if !FileExists(path) {
		C.ReportErrorToR_NoReturn(C.CString(fmt.Sprintf("ReadTmFrame() error: bad path '%s'; does not exist", path)))
	}

	///  begin TMFRAME read
	f, err := os.Open(path)
	if err != nil {
		C.ReportErrorToR_NoReturn(C.CString(fmt.Sprintf("ReadTmFrame() error, could not open path '%s': '%s'", path, err)))
	}
	i := int64(1)
	fr := tf.NewFrameReader(f, 1024*1024)

	var frame *tf.Frame
	//var raw []byte
	res := []*tf.Frame{}

toploop:
	for ; err == nil; i++ {
		frame, _, err, _ = fr.NextFrame(nil)
		if err != nil {
			if err == io.EOF {
				break toploop
			}
			C.ReportErrorToR_NoReturn(C.CString(fmt.Sprintf("ReadTmFrame() error reading '%s', fr.NextFrame() at i=%v gave error: '%v'",
				path, i, err)))
		}
		res = append(res, frame)
	} // end for toploop

	if len(res) > 0 {
		return tmFramesToR(res)
	}
	return C.R_NilValue
}

func tmFramesToR(slc []*tf.Frame) C.SEXP {
	n := len(slc)
	if n == 0 {
		return C.R_NilValue
	}

	cols := 2

	pti := slc[0].GetPTI()
	firstPti := pti
	var payloadList, payload2List C.SEXP

	switch pti {
	case tf.PtiOneInt64:
		payloadList = C.allocVector(C.STRSXP, C.R_xlen_t(n))
		C.Rf_protect(payloadList)
	case tf.PtiOneFloat64:
		payloadList = C.allocVector(C.REALSXP, C.R_xlen_t(n))
		C.Rf_protect(payloadList)
	case tf.PtiTwo64:
		payloadList = C.allocVector(C.REALSXP, C.R_xlen_t(n))
		C.Rf_protect(payloadList)
		payload2List = C.allocVector(C.STRSXP, C.R_xlen_t(n))
		C.Rf_protect(payload2List)
		cols++

	case tf.PtiUDE:
		payloadList = C.allocVector(C.VECSXP, C.R_xlen_t(n))
		C.Rf_protect(payloadList)

	case tf.PtiZero:
	case tf.PtiNull:
	case tf.PtiNA:
	case tf.PtiNaN:
	}

	returnList := C.allocVector(C.VECSXP, C.R_xlen_t(cols))
	C.Rf_protect(returnList)

	timestampSlice := C.allocVector(C.REALSXP, C.R_xlen_t(n))
	C.Rf_protect(timestampSlice)
	size := unsafe.Sizeof(C.double(0))

	var rhs C.double
	ptrNumSlice := unsafe.Pointer(C.REAL(timestampSlice))
	const msec = 1e6
	for i, f := range slc {
		// timestamp
		tmu := f.Tm()
		ftm := float64(tmu / msec)
		//fmt.Printf("tmu[%v]=%v / ftm=%v\n", i, tmu, ftm)
		rhs = C.double(ftm)
		*((*C.double)(unsafe.Pointer(uintptr(ptrNumSlice) + size*uintptr(i)))) = rhs

		// payload
		pti = f.GetPTI()
		if pti != firstPti {
			panic(fmt.Sprintf("inconsistent pti, firstPti was '%v', now we have '%v'",
				firstPti, pti))
		}
		switch pti {
		case tf.PtiOneInt64:
			C.SET_STRING_ELT(payloadList, C.R_xlen_t(i), C.mkChar(C.CString(fmt.Sprintf("%d", f.Ude))))
		case tf.PtiOneFloat64:
			rhs = C.double(f.V0)
			ptrPayList := unsafe.Pointer(C.REAL(payloadList))
			*((*C.double)(unsafe.Pointer(uintptr(ptrPayList) + size*uintptr(i)))) = rhs

		case tf.PtiTwo64:
			rhs = C.double(f.V0)
			ptrPayList := unsafe.Pointer(C.REAL(payloadList))
			*((*C.double)(unsafe.Pointer(uintptr(ptrPayList) + size*uintptr(i)))) = rhs
			C.SET_STRING_ELT(payload2List, C.R_xlen_t(i), C.mkChar(C.CString(fmt.Sprintf("%d", f.Ude))))

		case tf.PtiUDE:

			// probably json or msgpack, try to decode it.
			evtnum := f.GetEvtnum()
			if evtnum == tf.EvJson || (evtnum >= 2000 && evtnum <= 9999) {
				tmp := decodeJsonToR(f.Data)
				C.Rf_protect(tmp)
				C.SET_VECTOR_ELT(payloadList, C.R_xlen_t(i), tmp)
				C.Rf_unprotect_ptr(tmp)
			} else if evtnum == tf.EvMsgpKafka || evtnum == tf.EvMsgpack {
				tmp := decodeMsgpackToR(f.Data)
				C.Rf_protect(tmp)
				C.SET_VECTOR_ELT(payloadList, C.R_xlen_t(i), tmp)
				C.Rf_unprotect_ptr(tmp)
			}

		case tf.PtiZero:
		case tf.PtiNull:
		case tf.PtiNA:
		case tf.PtiNaN:
		}

	} // end for range slc

	C.SET_VECTOR_ELT(returnList, C.R_xlen_t(0), timestampSlice)
	C.SET_VECTOR_ELT(returnList, C.R_xlen_t(1), payloadList)
	if cols == 3 {
		C.SET_VECTOR_ELT(returnList, C.R_xlen_t(2), payload2List)
		C.Rf_unprotect_ptr(payload2List)
	}

	C.Rf_unprotect_ptr(timestampSlice)
	C.Rf_unprotect_ptr(payloadList)

	C.Rf_unprotect_ptr(returnList)
	return returnList
}
