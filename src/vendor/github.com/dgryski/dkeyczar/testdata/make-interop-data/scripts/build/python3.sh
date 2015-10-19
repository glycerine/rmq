#!/bin/sh -x
#python
cd clones/keyczar-main/python
python3 setup.py build
python3 setup.py test
cd ../../..
