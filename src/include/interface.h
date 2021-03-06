// -*- mode: C++; c-indent-level: 2; c-basic-offset: 2; tab-width: 8 -*-
///////////////////////////////////////////////////////////////////////////
// Copyright (C) 2014 Jason E. Aten                                      //
// License: Apache 2.0.                                                  //
// http://www.apache.org/licenses/                                       //
///////////////////////////////////////////////////////////////////////////


#ifndef INTERFACE_HPP
#define INTERFACE_HPP

#include <R.h>
#include <Rinternals.h>
#include <R_ext/Rdynload.h>
#include <R_ext/Utils.h>
#include <R_ext/Parse.h>
#include <Rembedded.h>
#include <signal.h>

#ifdef __cplusplus
extern "C" {
#endif

int Rf_isProtected(SEXP s); // debugging utility from R-3.x.x/src/main/memory.c
  
unsigned long int get_starting_signint_handler();
unsigned long int get_signint_handler();

  void restore_all_starting_signal_handlers();  
  extern struct sigaction starting_act[NSIG];

  SEXP rmq(SEXP name_);
  
  //  int JasonsLinkeMe();
  
  void ReportErrorToR_NoReturn(const char* msg);

  void PrintToR(const char* msg);

  void SetTypeToLANGSXP(SEXP* sexp);
  
  void WarnAndContinue(const char* msg);

  const char* get_string_elt(SEXP x, unsigned long long i);

  double get_real_elt(SEXP x, unsigned long long i);

  int get_int_elt(SEXP x, unsigned long long i);

  void set_lglsxp_true(SEXP lgl, unsigned long long i);
  void set_lglsxp_false(SEXP lgl, unsigned long long i);
  int get_lglsxp(SEXP lgl, unsigned long long i);

  // locate the next newline character in the raw array,
  // starting at beg, and up to but not including endx.
  // If not found, will return endx.
  unsigned long long next_newline_pos(SEXP raw, unsigned long long beg, unsigned long long endx);
  
  unsigned char* get_raw_elt_ptr(SEXP raw, unsigned long long i);  
  
  void callInitEmbeddedR();

  // PRE: callInitEmbeddedR() has been called exactly once before entering.
  // IMPORTANT: caller must PROTECT the returned SEXP and unprotect when done. Unless it is R_NilValue.
  SEXP callParseEval(const char* evalme, int* evalErrorOccurred);

  void callEndEmbeddedR();


#ifdef __cplusplus
}
#endif


#endif // INTERFACE_HPP
