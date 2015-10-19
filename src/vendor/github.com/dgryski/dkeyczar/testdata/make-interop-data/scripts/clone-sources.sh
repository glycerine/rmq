#!/bin/sh -x
mkdir clones
cd clones
git clone --recursive https://code.google.com/r/jtuley-keyczar-dev/  keyczar-main
cd keyczar-main
git merge origin/cpp/all
git merge origin/java/all
git merge origin/python/all
cd ..
git clone --recursive https://github.com/jbtule/keyczar-dotnet.git keyczar-dotnet
mkdir -p keyczar-go/src/github.com/dgryski
cd keyczar-go/src/github.com/dgryski
git clone --recursive https://github.com/dgryski/dkeyczar.git 
cd ../../../../..
