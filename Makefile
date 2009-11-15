include $(GOROOT)/src/Make.$(GOARCH)

TARG=mysql
CGOFILES=mysql.go
MW_CFLAGS=$(shell mysql_config --cflags)
CGO_LDFLAGS=mw.o $(shell mysql_config --libs)
CLEANFILES+=main

include $(GOROOT)/src/Make.pkg

mysql_mysql.so: mw.o mysql.cgo4.o
	gcc $(_CGO_CFLAGS_$(GOARCH)) $(_CGO_LDFLAGS_$(GOOS)) -o $@ mysql.cgo4.o $(CGO_LDFLAGS)

mw.o: mw.c mw.h
	gcc $(MW_CFLAGS) -o mw.o -c mw.c
