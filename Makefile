include $(GOROOT)/src/Make.$(GOARCH)

TARG=mysql
CGOFILES=mysql.go
MYSQL_CONFIG=$(shell which mysql_config)
MW_CFLAGS=$(shell $(MYSQL_CONFIG) --cflags)
CGO_LDFLAGS=mw.o $(shell $(MYSQL_CONFIG) --libs)
CLEANFILES+=mw.o example

include $(GOROOT)/src/Make.pkg

prereq:
	@test -x "$(MYSQL_CONFIG)" || \
		(echo "Can't find mysql_config in your path."; false)

mysql_mysql.so: mw.o mysql.cgo4.o
	gcc $(_CGO_CFLAGS_$(GOARCH)) $(_CGO_LDFLAGS_$(GOOS)) -o $@ mysql.cgo4.o $(CGO_LDFLAGS)

example: install example.go
	$(GC) example.go
	$(LD) -o $@ example.$O

mw.o: prereq mw.c mw.h
	gcc $(MW_CFLAGS) -o mw.o -c mw.c
