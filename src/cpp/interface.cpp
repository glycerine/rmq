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

  // called after the ((constructor)) routines.
void R_init_rmq(DllInfo *info)
{
  /* Register routines,
     allocate resources. */
  //printf("   R_init_rmq() called\n");
}

  // called when R wants to unload.
void R_unload_rmq(DllInfo *info)
{
  /* Release resources. */
  //printf("   R_unload_rmq() called\n");
}

struct sigaction starting_act[NSIG];

  // the ((constructor)) annotation makes this get
  // called before the cshared go routine runtime initializes,
  // which is why we manipulate the signal handling table here.
  // 
  // By setting SIGINT to be SIG_IGN, the go runtime won't take
  // it over. Then we can safely restore the R handler for
  // for SIGINT once the go runtime as completed initialization.
  // See the init() function in rmq.go for that logic.
  //
void __attribute__ ((constructor)) my_init(void) {
  for (int i = 0; i < NSIG; i++) {
    sigaction(i, NULL, &starting_act[i]);
  }
  //printf("   ++ a starts, starting_act.sa_handler = %p\n", starting_act.sa_handler);
  //printf("   constructor my_init for interface.cpp called!\n");

    // to avoid go taking over the SIGINT handler, we 
    // temporarily set SIGINT to SIG_IGN (no handler), which
    // means that the go runtime initialization in runtime/signal1_unix.go
    // will leave it alone. Hence we can later re-install the R
    // SIGINT handler and skip having the go runtime crash on 
    // them (OSX only; see https://github.com/golang/go/issues/13034)
    // This means our web server will need to poll R and ask it if
    // it wants us to return, even when "blocked" waiting on
    // a socket.
    struct sigaction act_with_ignore_sigint;
    act_with_ignore_sigint.sa_handler = SIG_IGN;
    sigaction(SIGINT, &act_with_ignore_sigint, NULL);
}

  void restore_all_starting_signal_handlers() {
    for (int i = 0; i < NSIG; i++) {
      sigaction(i, &starting_act[i], NULL);
    }
  }

unsigned long int get_starting_signint_handler() {
    return (unsigned long int)(starting_act[SIGINT].sa_handler);
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

 
