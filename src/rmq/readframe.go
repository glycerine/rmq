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

	payloadList := C.allocVector(C.VECSXP, C.R_xlen_t(n))
	C.Rf_protect(payloadList)

	returnList := C.allocVector(C.VECSXP, C.R_xlen_t(2))
	C.Rf_protect(returnList)

	var sxpTy C.SEXPTYPE = C.REALSXP
	numSlice := C.allocVector(sxpTy, C.R_xlen_t(n))
	C.Rf_protect(numSlice)
	size := unsafe.Sizeof(C.double(0))

	var rhs C.double
	ptrNumSlice := unsafe.Pointer(C.REAL(numSlice))
	const msec = 1e6
	for i, f := range slc {
		// timestamp
		tmu := f.Tm()
		ftm := float64(tmu / msec)
		fmt.Printf("tmu[%v]=%v / ftm=%v\n", i, tmu, ftm)
		rhs = C.double(ftm)
		*((*C.double)(unsafe.Pointer(uintptr(ptrNumSlice) + size*uintptr(i)))) = rhs

		// payload
		evtnum := f.GetEvtnum()
		if evtnum == tf.EvJson || (evtnum >= 2000 && evtnum <= 9999) {
			C.SET_VECTOR_ELT(payloadList, C.R_xlen_t(i), decodeJsonToR(f.Data))

		} else if evtnum == tf.EvMsgpKafka || evtnum == tf.EvMsgpack {
			C.SET_VECTOR_ELT(payloadList, C.R_xlen_t(i), decodeMsgpackToR(f.Data))
		}
	}

	C.SET_VECTOR_ELT(returnList, C.R_xlen_t(0), numSlice)
	C.SET_VECTOR_ELT(returnList, C.R_xlen_t(1), payloadList)

	C.Rf_unprotect_ptr(numSlice)
	C.Rf_unprotect_ptr(returnList)
	C.Rf_unprotect_ptr(payloadList)
	return returnList
}
