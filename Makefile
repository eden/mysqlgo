include $(GOROOT)/src/Make.$(GOARCH)

MYSQL_CONFIG=$(shell which mysql_config)

prereq:
	@test -x "$(MYSQL_CONFIG)" || \
		(echo "Can't find mysql_config in your path."; false)

db_install: db/Makefile
	cd db; make install

mysql_install: mysql/Makefile
	cd mysql; make install

install: prereq db_install mysql_install

example: install example.go
	$(GC) example.go
	$(LD) -o $@ example.$O

clean:
	cd db; make clean
	cd mysql; make clean
