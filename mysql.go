// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implements rudimentary MySQL support.
package mysql

/*
#include <stdlib.h>
#include <mysql.h>

char *itemAt(MYSQL_ROW row, unsigned int i) { return row[i]; }
MYSQL_FIELD fieldAt(MYSQL_FIELD *field, unsigned int i) { return field[i]; }
*/
import "C"

import "os"
import "fmt"
import "unsafe"

func init() {
	C.mysql_library_init(0, nil, nil);
}

// ConnInfo represents MySQL connection information.
//	- host: Host to connect to, passed directly into mysql_real_connect
//	  which will resolve DNS names.
//	- port: Port on which to connect.
//	- uname: Username to use for the connection.
//	- pass: Password to use for the connection.
//	- dbname: Initial database to use after successfully connecting.
type ConnInfo struct {
	host	string;
	port	int;
	uname	string;
	pass	string;
	dbname	string;
}

// Conn maintains the connection state of a single MySQL connection.
type Conn struct {
	h	C.MYSQL;
	connected	bool;
}

func NewConn() *Conn {
	my := new(Conn);
	C.mysql_init(&my.h);
	my.connected = false;
	return my;
}

// Returns the last error that occurred as an os.Error 
func (my *Conn) LastError() os.Error {
	if err := C.mysql_error(&my.h); *err != 0 {
		return os.NewError(C.GoString(err));
	}
	return nil;
}

// Connects to the server specified in the given connection info.
func (my *Conn) Connect(conn *ConnInfo) err os.Error {
	args := []*C.char{
		C.CString(conn.host), C.CString(conn.uname),
		C.CString(conn.pass), C.CString(conn.dbname)};

	C.mysql_real_connect(
		&my.h,
		args[0],
		args[1],
		args[2],
		args[3],
		C.uint(conn.port),
		nil,
		0);

	for i, _ := range args {
		C.free(unsafe.Pointer(args[i]))
	}

	err = my.LastError();
	my.connected = err != nil;

	return;
}

// Returns a new cursor for the current connection.  If the connection
// hasn't yet been established, nil is returned.
func (my *Conn) Cursor() *Cursor {
	if my.connected {
		return &Cursor{my, nil, 0};
	}
	return nil;
}

// Closes and cleans up the connection.
func (my *Conn) Close() {
	C.mysql_close(&my.h);
	my.connected = false;
}

// Wraps a MYSQL_RES structure and allows one to execute queries and
// read results.
type Cursor struct {
	my	*Conn;
	res	*C.MYSQL_RES;
	nfields	int;
}

func (c *Cursor) cleanup() {
	if c.res != nil {
		C.mysql_free_result(c.res);
		c.res = nil;
		c.nfields = 0;
	}
}

// Executes the given query.  Formats the query using fmt.Sprintf
// Use %q for strings that should be escaped.
func (c *Cursor) Execute(query string, parameters ...) os.Error {
	c.cleanup();

	query = fmt.Sprintf(query, parameters);

	// TODO figure out how to convert a string to a *C.char
	// and use mysql_real_query instead (saves malloc/copy)
	q := C.CString(query);
	C.mysql_query(&c.my.h, q);
	C.free(unsafe.Pointer(q));

	c.res = C.mysql_store_result(&c.my.h);
	c.nfields = int(C.mysql_num_fields(c.res));

	return c.my.LastError();
}

func (c *Cursor) Description() []Column {
	columns := make([]Column, c.nfields);
	fields := C.mysql_fetch_fields(c.res);

	for i := 0; i < c.nfields; i += 1 {
		field := C.fieldAt(fields, C.uint(i));
		columns[i] = Column{
			C.GoString(field.name),
			int(field._type)}
	}

	return columns;
}

func (c *Cursor) FetchOne() (data []interface {}, err os.Error) {
	row := C.mysql_fetch_row(c.res);
	err = c.my.LastError();

	if row != nil && err == nil {
		data = make([]interface {}, c.nfields);
		for i := 0; i < c.nfields; i += 1 {
			data[i] = C.GoString(C.itemAt(row, C.uint(i)));
		}
	}

	return;
}

func (c *Cursor) FetchMany(count int64) (rows []interface {}, err os.Error) {
	rows = make([]interface {}, count);

	i := 0;
	row, err := c.FetchOne();
	for row != nil && err == nil {
		rows[i] = row;
		i += 1;
		row, err = c.FetchOne();
	}

	if err != nil { rows = nil }

	return;
}

func (c *Cursor) FetchAll() ([]interface {}, os.Error) {
	return c.FetchMany(c.RowCount());
}

func (c *Cursor) RowCount() int64 {
	return int64(C.mysql_num_rows(c.res));
}

func (c *Cursor) Close() {
	c.cleanup();
}

type Column struct {
	Name	string;
	Type	int;
}
