#!/bin/sh -x
#dotnet
cd clones/keyczar-dotnet/Keyczar
.ci/PreXbuild.sh
./mono-build.sh Keyczar.sln
.ci/PostXbuild.sh
cd ../../..