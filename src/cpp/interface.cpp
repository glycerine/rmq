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

  // the ((constructor)) annotation makes my_init() get
  // called before the cshared go routine runtime initializes,
  // which is why we manipulate the signal handling table here.
  // 
  // By setting SIGINT to be SIG_IGN, the go runtime won't take
  // it over. Then we can safely restore the R handlers for
  // for SIGINT and all other signals once the go runtime as
  // completed initialization. The go routine won't ever
  // see any signals, which is good since when embedded as
  // a cshared library, it still doesn't play nicely with
  // the host process (https://github.com/golang/go/issues/13034)
  // as of October 2015/ go1.5.1.
  //
  // See the init() function in rmq.go for the logic that
  // restores the signal handlers.
  //
  void __attribute__ ((constructor)) my_init(void) {
    for (int i = 0; i < NSIG; i++) {
      sigaction(i, NULL, &starting_act[i]);
    }
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

  void callInitEmbeddedR() {
	char *my_argv[]= {(char*)"r.embedded.in.golang", (char*)"--silent", (char*)"--vanilla", (char*)"--slave"};
    Rf_initEmbeddedR(sizeof(my_argv)/sizeof(my_argv[0]), my_argv);
  }

  // PRE: callInitEmbeddedR() has been called exactly once before entering.
  // IMPORTANT: caller must PROTECT the returned SEXP and unprotect when done. Unless it is R_NilValue.
  SEXP callParseEval(const char* evalme, int* evalErrorOccured) {
    SEXP ans,expression, myCmd;
    evalErrorOccured = 0;
    ParseStatus parseStatusCode;

    PROTECT(myCmd = mkString(evalme));

    /* PARSE_NULL will not be returned by R_ParseVector 
       typedef enum {
       PARSE_NULL,
       PARSE_OK,
       PARSE_INCOMPLETE,
       PARSE_ERROR,
       PARSE_EOF
       } ParseStatus;
    */
    PROTECT(expression = R_ParseVector(myCmd, 1, &parseStatusCode, R_NilValue));
    if (parseStatusCode != PARSE_OK) {
      UNPROTECT(2);
      return R_NilValue;
    }

    ans = R_tryEval(VECTOR_ELT(expression,0), R_GlobalEnv, evalErrorOccured);
    UNPROTECT(2);
    // evalErrorOccured will be 1 if an error happened, and ans will be R_NilValue
    return ans; // caller must protect and unprotect when done
  }

  void callEndEmbeddedR() {
    Rf_endEmbeddedR(0);
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

 
