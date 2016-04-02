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

	/*
		if evtnum == EvJson || (evtnum >= 2000 && evtnum <= 9999) {
			pp := prettyPrintJson(prettyPrint, frame.Data)
			fmt.Fprintf(w, "  %s", string(pp))
		}
		if evtnum == EvMsgpKafka || evtnum == EvMsgpack {
			// decode msgpack to json with ugorji/go/codec

			var iface interface{}
			dec := codec.NewDecoderBytes(frame.Data, &msgpHelper.mh)
			err := dec.Decode(&iface)
			panicOn(err)

			//Q("iface = '%#v'", iface)

			var wbuf bytes.Buffer
			enc := codec.NewEncoder(&wbuf, &msgpHelper.jh)
			err = enc.Encode(&iface)
			panicOn(err)
			pp := prettyPrintJson(prettyPrint, wbuf.Bytes())
			fmt.Fprintf(w, " %s", string(pp))
		}

		// return decodeMsgpackToR(bytes)

		}
	*/

	returnList := C.allocVector(C.VECSXP, C.R_xlen_t(n))
	C.Rf_protect(returnList)

	const msec = 1e6
	for i, f := range slc {
		tmu := f.Tm()
		fmt.Printf("tmu[%v]=%v\n", i, tmu)
		//tm := time.Unix(0, tmu).UTC()
		//evtnum := f.GetEvtnum()

		//		C.SET_VECTOR_ELT(returnList, C.R_xlen_t(i), C.Rf_ScalarReal(C.double(float64(tmu/msec))))
	}
	C.Rf_unprotect_ptr(returnList)
	return returnList
}
