include $(GOROOT)/src/Make.$(GOARCH)

TARG=mysql
CGOFILES=mysql.go
MYSQL_CONFIG=$(shell which mysql_config)
CGO_CFLAGS=$(shell $(MYSQL_CONFIG) --cflags)
CGO_LDFLAGS=$(shell $(MYSQL_CONFIG) --libs)

include $(GOROOT)/src/Make.pkg

prereq:
	@test -x "$(MYSQL_CONFIG)" || \
		(echo "Can't find mysql_config in your path."; false)

ex: install ex.go
	$(GC) ex.go
	$(LD) -o $@ ex.$O
