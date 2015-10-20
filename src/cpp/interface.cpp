///////////////////////////////////////////////////////////////////////////
// Copyright (C) 2014 Jason E. Aten                                      //
// License: Apache 2.0.                                                  //
// http://www.apache.org/licenses/                                       //
///////////////////////////////////////////////////////////////////////////

#include <stdint.h>
#include <string>
#include <sstream>
#include <stdexcept>
#include <string.h>
#include <stdlib.h>

#include "interface.h"



#ifdef __cplusplus
extern "C" {
#endif


int symbol_string_to_int(const char* s) {
  if (0 == strncmp(s, "hello", 6)) {
    return 2;
  }
  if (0==strncmp(s,"world", 6)) {
    return 43;
  }
  return -1000;
}

SEXP rmq(SEXP str_) {
  SEXP ans; 

  if(TYPEOF(str_) != STRSXP) {
    REprintf("argument to rmq() must be a string to be decoded to its integer constant value in the rmq pkg.\n");
    return R_NilValue;
  }

  PROTECT(ans = allocVector(INTSXP,1));
  const char* s = CHAR(STRING_ELT(str_,0));
  int rc = symbol_string_to_int(s);
  INTEGER(ans)[0] = rc;
  UNPROTECT(1);
  if (rc == -1000) {
    REprintf("error: could not translate string '%s' to rmq constant.\n", s);
    return R_NilValue;
  }
  return ans;
}

  int JasonsLinkeMe() {
    printf("\n\n 88888 JasonsLinkeMe called!\n\n");
    return 7777;
  }

  void ReportErrorToR_NoReturn(const char* msg) {
    Rf_error(msg);
  }
  
  void PrintToR(const char* msg) {
    REprintf(msg);
  }
  
  void SetTypeToLANGSXP(SEXP* sexp) {
    SET_TYPEOF(*sexp, LANGSXP);
  }


#ifdef __cplusplus
}
#endif

 
