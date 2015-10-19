#!/bin/sh -x
scripts/build/java.sh > logs/build_java.txt  2>&1
scripts/build/python.sh > logs/build_python.txt  2>&1
scripts/build/python3.sh > logs/build_python3.txt 2>&1
scripts/build/cpp.sh > logs/build_cpp.txt  2>&1
scripts/build/dotnet.sh > logs/build_dotnet.txt  2>&1
scripts/build/go.sh > logs/build_go.txt  2>&1
