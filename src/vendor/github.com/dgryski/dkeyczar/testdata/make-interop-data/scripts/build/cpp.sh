#!/bin/sh -x
#++
cd clones/keyczar-main/cpp/src/keyczar
make kctests
make keyczart
cd ../../../../..