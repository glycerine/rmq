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

	i := int64(1)
	fr := tf.NewFrameReader(os.Stdin, 1024*1024)

	var frame tf.Frame
	//var raw []byte
	var err error

toploop:
	for ; err == nil; i++ {
		//_, _, err, raw = fr.NextFrame(&frame)
		_, _, err, _ = fr.NextFrame(&frame)
		if err != nil {
			if err == io.EOF {
				break toploop
			}
			C.ReportErrorToR_NoReturn(C.CString(fmt.Sprintf("ReadTmFrame() error reading '%s', fr.NextFrame() at i=%v gave error: '%v'",
				path, i, err)))
		}
		str := frame.Stringify(-1, false, false, false)
		fmt.Printf("I see '%s'\n", str)
	} // end for toploop

	// return decodeMsgpackToR(bytes)
	return C.R_NilValue
}
