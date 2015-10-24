all:
	cd src && make

check:
	cd src && make check

clean:
	rm -f *~ && cd src && make clean

install:
	cd src && make install

build:
	cd src && make build

superclean:
	find . -name '*~' | xargs rm -f
	cd src && make superclean
