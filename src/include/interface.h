// -*- mode: C++; c-indent-level: 2; c-basic-offset: 2; tab-width: 8 -*-
///////////////////////////////////////////////////////////////////////////
// Copyright (C) 2014 Jason E. Aten                                      //
// License: Apache 2.0.                                                  //
// http://www.apache.org/licenses/                                       //
///////////////////////////////////////////////////////////////////////////


#ifndef INTERFACE_HPP
#define INTERFACE_HPP

#include <Rinternals.h>
#include <signal.h>

#ifdef __cplusplus
extern "C" {
#endif

unsigned long int get_starting_signint_handler();
unsigned long int get_signint_handler();

  
  extern struct sigaction starting_act;

  SEXP rmq(SEXP name_);
  
  //  int JasonsLinkeMe();
  
  void ReportErrorToR_NoReturn(const char* msg);

  void PrintToR(const char* msg);

  void SetTypeToLANGSXP(SEXP* sexp);
  
  void WarnAndContinue(const char* msg);

  const char* get_string_elt(SEXP x, int i);

  double get_real_elt(SEXP x, int i);

  int get_int_elt(SEXP x, int i);

#ifdef __cplusplus
}
#endif


#endif // INTERFACE_HPP
