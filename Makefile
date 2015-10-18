all:
	cd src && make

check:
	cd src && make check

clean:
	rm -f *~ && cd src && make clean
