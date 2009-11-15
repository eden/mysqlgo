// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implements rudimentary MySQL support.
package mysql

// #include <stdlib.h>
// #include "mw.h"
import "C"

import "os"
import "fmt"
import "unsafe"

func init() {
	C.mw_library_init();
}

// ConnInfo represents MySQL connection information.
//	- host: Host to connect to, passed directly into mw_real_connect
//	  which will resolve DNS names.
//	- port: Port on which to connect.
//	- uname: Username to usemw for the connection.
//	- pass: Password to usemw for the connection.
//	- dbname: Initial database to usemw after successfully connecting.
type ConnInfo struct {
	host	string;
	port	int;
	uname	string;
	pass	string;
	dbname	string;
}

// Conn maintains the connection state of a single MySQL connection.
type Conn struct {
	h	C.mw;
}

func NewConn() *Conn {
	my := new(Conn);
	my.h = nil;
	return my;
}

func usemw(h C.mw) unsafe.Pointer { return unsafe.Pointer(h); }
func usemwres(h C.mwres) unsafe.Pointer { return unsafe.Pointer(h); }
func usemwrow(h C.mwrow) unsafe.Pointer { return unsafe.Pointer(h); }
func usemwfield(h C.mwfield) unsafe.Pointer { return unsafe.Pointer(h); }

// Returns the last error that occurred as an os.Error 
func (my *Conn) LastError() os.Error {
	if err := C.mw_error(usemw(my.h)); *err != 0 {
		return os.NewError(C.GoString(err));
	}
	return nil;
}

// Connects to the server specified in the given connection info.
func (my *Conn) Connect(conn *ConnInfo) (err os.Error) {
	args := []*C.char{
		C.CString(conn.host), C.CString(conn.uname),
		C.CString(conn.pass), C.CString(conn.dbname)};

	if (my.h != nil) {
		my.Close();
	}

	my.h = C.mw_init(nil);
	if my.h == nil { return os.ENOMEM }

	C.mw_real_connect(
		usemw(my.h),
		args[0],
		args[1],
		args[2],
		args[3],
		C.int(conn.port));

	for i, _ := range args {
		C.free(unsafe.Pointer(args[i]))
	}

	err = my.LastError();
	if err != nil { my.h = nil };

	return;
}

// Returns a new cursor for the current connection.  If the connection
// hasn't yet been established, nil is returned.
func (my *Conn) Cursor() *Cursor {
	if my.h != nil {
		return &Cursor{my, nil, 0};
	}
	return nil;
}

// Closes and cleans up the connection.
func (my *Conn) Close() {
	C.mw_close(usemw(my.h));
}

// Wraps a MYSQL_RES structure and allows one to execute queries and
// read results.
type Cursor struct {
	my	*Conn;
	res	C.mwres;
	nfields	int;
}

func (c *Cursor) cleanup() {
	if c.res != nil {
		C.mw_free_result(usemwres(c.res));
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
	// and usemw mw_real_query instead (saves malloc/copy)
	q := C.CString(query);
	C.mw_query(usemw(c.my.h), q);
	C.free(unsafe.Pointer(q));

	c.res = C.mw_store_result(usemw(c.my.h));
	c.nfields = int(C.mw_num_fields(usemwres(c.res)));

	return c.my.LastError();
}

func (c *Cursor) Description() []Column {
	columns := make([]Column, c.nfields);
	fields := C.mw_fetch_fields(usemwres(c.res));

	for i := 0; i < c.nfields; i += 1 {
		columns[i] = Column{
			C.GoString(C.mw_field_name_at(usemwfield(fields), C.int(i))),
			int(C.mw_field_type_at(usemwfield(fields), C.int(i)))}
	}

	return columns;
}

func (c *Cursor) FetchOne() (data []interface {}, err os.Error) {
	row := C.mw_fetch_row(usemwres(c.res));
	err = c.my.LastError();

	if row != nil && err == nil {
		data = make([]interface {}, c.nfields);
		for i := 0; i < c.nfields; i += 1 {
			data[i] = C.GoString(C.mw_row(usemwrow(row), C.int(i)));
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
	return int64(C.mw_num_rows(usemwres(c.res)));
}

func (c *Cursor) Close() {
	c.cleanup();
}

type Column struct {
	Name	string;
	Type	int;
}
