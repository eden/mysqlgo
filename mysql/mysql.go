// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implements rudimentary MySQL support.
package mysql

/*
#include <stdlib.h>
#include <string.h>
#include <mysql.h>

MYSQL_BIND *create_bind(int i) {
	return malloc(sizeof(MYSQL_BIND) * i);
}

void assign_bind(
	MYSQL_BIND *binds, unsigned int i,
	enum enum_field_types type,
	void *buf, unsigned long buflen,
	void *len, void *nul, void *error)
{
	binds[i].buffer_type = type;
	binds[i].buffer = buf;
	binds[i].buffer_length = buflen;
	binds[i].length = (unsigned long *)len;
	binds[i].is_null = (my_bool *)nul;
	binds[i].error = (my_bool *)error;
}
*/
import "C"

// phf's experimental DB interface
import "db"
import "os"
import "fmt"
import "sync"
import "unsafe"
import "reflect"
import "strings"

type MysqlError os.ErrorString;

func (e MysqlError) String() string { return string(e) }

var Open db.OpenSignature;
var Version db.VersionSignature;

// Maintains the connection state of a single MySQL connection.
type Connection struct {
	handle	*C.MYSQL;
	lock	*sync.Mutex;
}

// The following arguments can be used to setup the connection:
//
//	- host: Host to connect to, passed directly into mysql_real_connect
//	  which will resolve DNS names.
//	- port: Port on which to connect (int)
//  - socket: Path to unix socket
//	- username: Username to use for the connection.
//	- password: Password to use for the connection.
//	- database: Initial database to use after successfully connecting.
//
// If none are provided, the default will be taken from the user/system-wide
// mysql configuration.
func open(args map[string]interface{}) (conn db.Connection, err os.Error) {
	var host, uname, passwd, dbname, socket *C.char;
	var port C.uint;

	// Local helper function to unpack a CString from the passed in dictionary
	// arguments
	cstringFromMap := func(d map[string]interface{}, key string) (s *C.char) {
		if v, f := d[key]; f { if v, f := v.(string); f { s = C.CString(v) } }
		return;
	};

	host = cstringFromMap(args, "host");
	socket = cstringFromMap(args, "socket");

	// Unpack a uint from the dictionary
	if v, f := args["port"]; f { if s, f := v.(int); f { port = C.uint(s) } }

	if socket == nil && host == nil {
		err = MysqlError("Either 'host' or 'socket' must be defined in args");
		goto cleanup;
	}

	uname = cstringFromMap(args, "username");
	passwd = cstringFromMap(args, "password");
	dbname = cstringFromMap(args, "database");

	c := Connection{};
	c.lock = new(sync.Mutex);
	c.handle = C.mysql_init(nil);
	if c.handle == nil {
		err = MysqlError("Couldn't init handle (likely out of memory?)");
		goto cleanup;
	}

	rc := C.mysql_real_connect(
		c.handle,
		host,
		uname,
		passwd,
		dbname,
		port,
		socket,
		0);	// client flags

	// If an error was set, or if the handle returned is not the same as the
	// one we allocated, there was a problem.
	err = c.LastError();
	if err != nil || rc != c.handle {
		C.mysql_close(c.handle);
	}
	else {
		// Everything's ok, set the returned the connection
		conn = c;
	}

cleanup:
	if host != nil {
		C.free(unsafe.Pointer(host))
	}
	if uname != nil {
		C.free(unsafe.Pointer(uname))
	}
	if passwd != nil {
		C.free(unsafe.Pointer(passwd))
	}
	if dbname != nil {
		C.free(unsafe.Pointer(dbname))
	}
	if socket != nil {
		C.free(unsafe.Pointer(socket))
	}

	return;
}

func version() (ver map[string]string, err os.Error) {
	cver := C.mysql_get_client_version();
	ver["client"] = fmt.Sprintf("%d", uint64(cver));
	return;
}

func init() {
	C.mysql_library_init(0, nil, nil);
	Open = open;
	Version = version;
}

// Returns the last error that occurred as an os.Error
func (conn Connection) LastError() os.Error {
	if err := C.mysql_error(conn.handle); *err != 0 {
		return os.NewError(C.GoString(err))
	}
	return nil;
}

func (conn Connection) Prepare(query string) (dbs db.Statement, e os.Error) {
	s := Statement{};
	s.conn = &conn;
	s.stmt = C.mysql_stmt_init(conn.handle);
	if s.stmt == nil {
		e = MysqlError("Prepare: Couldn't init statement (out of memory?)");
		return;
	}

	cquery := C.CString(query);
	if r := C.mysql_stmt_prepare(s.stmt, cquery, C.ulong(len(query))); r != 0 {
		e = conn.LastError()
	} else {
		dbs = s
	}
	C.free(unsafe.Pointer(cquery));

	return;
}

func (conn Connection) Lock()		{ conn.lock.Lock() }
func (conn Connection) Unlock()	{ conn.lock.Unlock() }


func createParamBinds(args ...) (binds *C.MYSQL_BIND, data []BoundData, err os.Error) {
	a := reflect.NewValue(args).(*reflect.StructValue);
	fcount := a.NumField();
	if fcount > 0 {
		binds = C.create_bind(C.int(fcount));
		data = make([]BoundData, fcount);
		for i := 0; i < fcount; i++ {
			switch arg := a.Field(i).(type) {
			default:
				err = MysqlError(
					fmt.Sprintf("Unsupported param type %T", arg));
				break

			case *reflect.StringValue:
				s := arg.Get();

				data[i] = *NewBoundData(
					MysqlTypeString,
					strings.Bytes(s),
					len(s));

				C.assign_bind(
					binds,
					C.uint(i),
					MysqlTypeString,
					unsafe.Pointer(&data[i].buffer),
					C.ulong(len(s)),
					unsafe.Pointer(&data[i].blen),
					unsafe.Pointer(&data[i].is_null),
					unsafe.Pointer(&data[i].error));
			}
		}
	}
	if err != nil {
		C.free(unsafe.Pointer(binds));
		binds = nil;
		data = nil
	}
	return;
}

func createResultBinds(stmt *C.MYSQL_STMT) (*C.MYSQL_BIND, *[]BoundData) {
	meta := C.mysql_stmt_result_metadata(stmt);
	if meta != nil {
		fcount := C.mysql_num_fields(meta);
		binds := C.create_bind(C.int(fcount));
		data := make([]BoundData, fcount);
		for i := C.uint(0); i < fcount; i+=1 {
			field := C.mysql_fetch_field_direct(meta, i);

			data[i] = *NewBoundData(
				MysqlType(field._type),
				nil,
				int(field.length)
			);

			C.assign_bind(
				binds, i,
				field._type,
				unsafe.Pointer(&data[i].buffer),
				field.length,
				unsafe.Pointer(&data[i].blen),
				unsafe.Pointer(&data[i].is_null),
				unsafe.Pointer(&data[i].error));
		}
		C.mysql_free_result(meta);
		return binds, &data;
	}
	return nil, nil
}

func (conn Connection) Execute(stmt db.Statement, parameters ...)
		(dbcur db.Cursor, err os.Error)
{
	if s, ok := stmt.(Statement); ok {
		var (binds *C.MYSQL_BIND; data []BoundData; e os.Error);
		pcount := uint64(C.mysql_stmt_param_count(s.stmt));

		if pcount > 0 {
			if binds, data, e = createParamBinds(parameters); err != nil {
				if rc := C.mysql_stmt_bind_param(s.stmt, binds); rc != 0 {
					err = conn.LastError();
					C.free(unsafe.Pointer(binds));
					return
				}
			}
			else {
				err = e
			}
			// prevent data no-use errors.  We just need to keep it around so
			// that GC doesn't clean it up.
			data = data
		}

		if rc := C.mysql_stmt_execute(s.stmt); rc != 0 {
			err = conn.LastError()
		}
		else {
			dbcur = NewCursorValue(s)
		}

		if binds != nil {
			C.free(unsafe.Pointer(binds));
		}
	}
	else {
		err = MysqlError("Execute: 'stmt' is not a mysql.Statement");
	}
	return
}

// Closes and cleans up the connection.
func (conn Connection) Close() os.Error {
	C.mysql_close(conn.handle);
	conn.handle = nil;
	return nil;
}

type Statement struct {
	stmt	*C.MYSQL_STMT;
	conn	*Connection;
}

func (s Statement) Close() (err os.Error) {
	if s.stmt != nil {
		if r := C.mysql_stmt_close(s.stmt); r != 0 {
			err = s.conn.LastError()
		}
		s.stmt = nil
	}
	return nil
}

type Cursor struct {
	stmt	*Statement;
	rbinds	*C.MYSQL_BIND;
	rdata	*[]BoundData;
	bound	bool;
}

func NewCursorValue(s Statement) (Cursor) {
	cur := Cursor{};
	cur.stmt = &s;
	cur.bound = false;
	(&cur).setupResultBinds();
	return cur
}

func (c Cursor) MoreResults() bool {
	return false
}

func (c *Cursor) setupResultBinds() (err os.Error) {
	if c.bound { return }
	c.rbinds, c.rdata = createResultBinds(c.stmt.stmt);
	if c.rbinds != nil {
		if rc := C.mysql_stmt_bind_result(c.stmt.stmt, c.rbinds); rc != 0 {
			err = c.stmt.conn.LastError()
		}
	}
	c.bound = true;
	return;
}

func (c Cursor) FetchOne() (res []interface{}, err os.Error) {
	if rc := C.mysql_stmt_fetch(c.stmt.stmt); rc == 0 {
		res = make([]interface{}, len(*c.rdata));
		for i := range(*c.rdata) {
			res[i], _ = (*c.rdata)[i].Value()
		}
	}
	else if rc == 100 {
		// no data
	}
	else {
		err = c.stmt.conn.LastError()
	}
	return
}

func (c Cursor) FetchMany(count int) (res [][]interface{}, err os.Error) {
	return
}

func (c Cursor) FetchAll() (res [][]interface{}, err os.Error) {
	return
}

func (c Cursor) Close() (err os.Error) {
	if c.rbinds != nil {
		C.free(unsafe.Pointer(c.rbinds));
		c.rbinds = nil
	}
	c.rdata = nil;
	c.bound = false;
	return
}
