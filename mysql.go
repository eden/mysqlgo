// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implements rudimentary MySQL support.
package mysql

// #include <stdlib.h>
// #include <mysql.h>
// char *row_at(MYSQL_ROW row, int i) { return row[i]; }
// MYSQL_FIELD field_at(MYSQL_FIELD *fields, int i) { return fields[i]; }
import "C"

import "os"
import "fmt"
import "sync"
import "unsafe"

func init() {
	C.mysql_library_init(0, nil, nil);
}

var MaxFetchCount = 65535

// ConnInfo represents MySQL connection information.
//	- host: Host to connect to, passed directly into mysql_real_connect
//	  which will resolve DNS names.
//	- port: Port on which to connect.
//	- uname: Username to use for the connection.
//	- pass: Password to use for the connection.
//	- dbname: Initial database to use after successfully connecting.
type ConnInfo struct {
	Host	string;
	Port	int;
	Uname	string;
	Pass	string;
	Dbname	string;
}

// Conn maintains the connection state of a single MySQL connection.
type Conn struct {
	h	*C.MYSQL;
	queryLock	sync.Mutex;
}

// Return a new database connection.
func NewConn() *Conn	{ return new(Conn) }

// Returns the last error that occurred as an os.Error 
func (my *Conn) LastError() os.Error {
	if err := C.mysql_error(my.h); *err != 0 {
		return os.NewError(C.GoString(err));
	}
	return nil;
}

// Connects to the server specified in the given connection info.
func (my *Conn) Connect(conn *ConnInfo) (err os.Error) {
	args := []*C.char{
		C.CString(conn.Host), C.CString(conn.Uname),
		C.CString(conn.Pass), C.CString(conn.Dbname)};

	if (my.h != nil) { my.Close() }

	my.h = C.mysql_init(nil);
	if my.h == nil { return os.ENOMEM }

	C.mysql_real_connect(
		my.h,
		args[0],
		args[1],
		args[2],
		args[3],
		C.uint(conn.Port),
		nil,
		0);

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

func (my *Conn) lock()	{ my.queryLock.Lock() }
func (my *Conn) unlock()	{ my.queryLock.Unlock() }

// Closes and cleans up the connection.
func (my *Conn) Close() {
	C.mysql_close(my.h);
	my.h = nil;
}

// Wraps a MYSQL_RES structure and allows one to execute queries and
// read results.
type Cursor struct {
	my	*Conn;
	res	*C.MYSQL_RES;
	nfields	int;
}

// Closes and frees the currently stored result.
func (c *Cursor) cleanup() {
	if c.res != nil {
		C.mysql_free_result(c.res);
		c.res = nil;
		c.nfields = 0;
	}
}

// Executes the given query and stores the result into the cursor.  Use the
// Fetch* methods to access results.  Before sending the query to mysql,
// Execute will send the query and the varargs to fmt.Sprintf. Use %q for
// strings that should be escaped.
//
// At the moment, this does not perform any SQL injection cleansing.  It's up
// to the caller to make sure its queries are not subject to attack.
func (c *Cursor) Execute(query string, parameters ...) (err os.Error) {
	c.cleanup();

	query = fmt.Sprintf(query, parameters);

	// mysql_query and mysql_store_result can't be interleaved between threads
	// on the same connection so we need to lock the two operations together
	c.my.lock();

	// TODO figure out how to convert a string to a *C.char
	// and use mysql_real_query instead (saves malloc/copy)
	q := C.CString(query);
	rcode := C.mysql_query(c.my.h, q);
	C.free(unsafe.Pointer(q));

	if err = c.my.LastError(); err != nil || rcode != 0 {
		if err == nil { err = os.NewError("Query failed.") }
		goto UnlockAndReturn
	}

	c.nfields = int(C.mysql_field_count(c.my.h));
	c.res = C.mysql_store_result(c.my.h);
	err = c.my.LastError();
	if err != nil || (c.res == nil && c.nfields > 0) {
		if err == nil {
			err = os.NewError("No results returned.");
			c.cleanup();
		}
		goto UnlockAndReturn
	}

UnlockAndReturn:
	c.my.unlock();
	return;
}

// Returns a tuple of column names and types of the current result.
func (c *Cursor) Description() []Column {
	columns := make([]Column, c.nfields);

	if c.res != nil && c.nfields > 0 {
		fields := C.mysql_fetch_fields(c.res);

		for i := 0; i < c.nfields; i += 1 {
			columns[i] = Column{
				C.GoString(C.field_at(fields, C.int(i)).name),
				int(C.field_at(fields, C.int(i))._type)}
		}
	}

	return columns;
}

// Returns a tuple of the next row in the result set.  If there are no results
// or an error occurred, nil is returned.  The error, if any, is given as the
// second return value.
func (c *Cursor) FetchOne() (data []interface {}, err os.Error) {
	if c.res == nil { return nil, os.NewError("Fetch called before Execute") }

	row := C.mysql_fetch_row(c.res);
	err = c.my.LastError();

	if row != nil && err == nil {
		data = make([]interface {}, c.nfields);
		for i := 0; i < c.nfields; i += 1 {
			data[i] = C.GoString(C.row_at(row, C.int(i)));
		}
	}

	return;
}

// Returns a list of tuples for `count` rows in the result set.  May return
// less than `count` rows if there are not enough rows left in the result set
// left.  nil is returned only if an error occurred.  The error, if any, is
// given as the second return value.  Returns an error if count is 0.
func (c *Cursor) FetchMany(count uint16) (rows [][]interface {}, err os.Error) {
	if count == 0 { return nil, os.NewError("Invalid count") }

	rows = make([][]interface {}, count);
	i := uint16(0);
	row, err := c.FetchOne();
	for i < count && row != nil && err == nil {
		rows[i] = row;
		i += 1;
		row, err = c.FetchOne();
	}

	if err != nil { rows = nil }

	return;
}

// Returns a list of tuples of all remaining tuples in the result set.  nil is
// returned only if an error occurred.  The error, if any, is given as the
// second return value.  If the result set contains more than MaxFetchCount
// rows, an error is returned.
func (c *Cursor) FetchAll() ([][]interface {}, os.Error) {
	count := c.RowCount();
	if count >= uint64(MaxFetchCount) {
		return nil, os.NewError(
			"Too many rows in result set.  Use FetchOne or FetchMany instead")
	}
	return c.FetchMany(uint16(count));
}

// Returns the number of rows returned from the current result set.
func (c *Cursor) RowCount() uint64 {
	if c.res == nil { return 0 }
	return uint64(C.mysql_num_rows(c.res));
}

// Closes the current result set and prepares the cursor for re-use.
func (c *Cursor) Close() {
	c.cleanup();
}

type Column struct {
	Name	string;
	Type	int;
}
