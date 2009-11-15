include $(GOROOT)/src/Make.$(GOARCH)

TARG=mysql
CGOFILES=mysql.go
CGO_CFLAGS=$(shell mysql_config --cflags)
CGO_LDFLAGS=$(shell mysql_config --libs)
CLEANFILES+=main

include $(GOROOT)/src/Make.pkg

main: install main.go
	$(GC) main.go
	$(LD) -o $@ main.$O
