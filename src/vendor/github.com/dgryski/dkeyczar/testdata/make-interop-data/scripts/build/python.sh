#!/bin/sh -x
#python
cd clones/keyczar-main/python
python setup.py build
python setup.py test
cd ../../..