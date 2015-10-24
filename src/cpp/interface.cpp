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
#include <string.h>
#include <signal.h>

#include "interface.h"




#ifdef __cplusplus
extern "C" {
#endif

struct sigaction starting_act;

void __attribute__ ((constructor)) my_init(void) {
    sigaction(SIGINT, NULL, &starting_act);
    printf("   ++ a starts, starting_act.sa_handler = %p\n", starting_act.sa_handler);
    printf("   constructor my_init for interface.cpp called!\n");
}

unsigned long int get_starting_signint_handler() {
    return (unsigned long int)(starting_act.sa_handler);
}

unsigned long int get_signint_handler() {
    struct sigaction act;
    sigaction(SIGINT, NULL, &act);    
    return (unsigned long int)(act.sa_handler);
}
  
  

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

  //  int JasonsLinkeMe() {
  //    printf("\n\n 88888 JasonsLinkeMe called!\n\n");
  //    return 7777;
  //  }

  void ReportErrorToR_NoReturn(const char* msg) {
    Rf_error(msg);
  }
  
  void PrintToR(const char* msg) {
    REprintf(msg);
  }
  
  void WarnAndContinue(const char* msg) {
    warning(msg);
  }
  
  void SetTypeToLANGSXP(SEXP* sexp) {
    SET_TYPEOF(*sexp, LANGSXP);
  }

  const char* get_string_elt(SEXP x, int i) {
    return CHAR(STRING_ELT(x, i));
  }

  double get_real_elt(SEXP x, int i) {
    return REAL(x)[i];
  }


  int get_int_elt(SEXP x, int i) {
    return INTEGER(x)[i];
  }



  SEXP CallbackToHandler(SEXP handler_, SEXP arg_, SEXP rho_) {
    SEXP evalres;
    
    SEXP R_fcall, msg;
    if(!isFunction(handler_)) error("‘handler’ must be a function");
    if(!isEnvironment(rho_)) error("‘rho’ should be an environment. e.g. new.env()");
    
    PROTECT(R_fcall = lang2(handler_, arg_));
    PROTECT(evalres = eval(R_fcall, rho_));
    UNPROTECT(2);
    return evalres;
  }
  

#ifdef __cplusplus
}
#endif

 
