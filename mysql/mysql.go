// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implements rudimentary MySQL support.
package mysql

/*
#include <stdlib.h>
#include <string.h>
#include <mysql.h>

MYSQL_BIND *mysql_bind_create_list(int count) {
	return malloc(sizeof(MYSQL_BIND) * count);
}

void mysql_bind_free(MYSQL_BIND *binds) {
	free(binds);
}

void mysql_bind_assign(
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

// Based on
// http://dev.mysql.com/doc/refman/5.0/en/c-api-prepared-statement-datatypes.html
signed char _fromTiny(void *p) { return *((signed char *)p); }
short int _fromShort(void *p) { return *((short int *)p); }
int _fromLong(void *p) { return *((int *)p); }
long long int _fromLonglong(void *p) { return *((long long int *)p); }
float _fromFloat(void *p) { return *((float *)p); }
double _fromDouble(void *p) { return *((double *)p); }
char _charAt(void *p, int i) { return *((char *) (p + i)); }
*/
import "C"

import (
	"db";
	"os";
	"fmt";
	"sync";
	"http";
	"unsafe";
	"reflect";
	"strings";
	"strconv";
)

func platformConvertTiny(ptr unsafe.Pointer) int8 {
	return int8(C._fromTiny(ptr))
}
func platformConvertShort(ptr unsafe.Pointer) int16 {
	return int16(C._fromShort(ptr))
}
func platformConvertLong(ptr unsafe.Pointer) int {
	return int(C._fromLong(ptr))
}
func platformConvertLonglong(ptr unsafe.Pointer) int64 {
	return int64(C._fromLonglong(ptr))
}
func platformConvertFloat(ptr unsafe.Pointer) float32 {
	return float32(C._fromFloat(ptr))
}
func platformConvertDouble(ptr unsafe.Pointer) float64 {
	return float64(C._fromDouble(ptr))
}
func platformConvertString(ptr unsafe.Pointer, slen int) string {
	bytes := make([]byte, slen);
	for i := range (bytes) {
		bytes[i] = byte(C._charAt(ptr, C.int(i)))
	}
	return string(bytes);
}

type MysqlError os.ErrorString

func (e MysqlError) String() string	{ return string(e) }

var Open db.OpenSignature
var Version db.VersionSignature

// Maintains the connection state of a single MySQL connection.
type Connection struct {
	handle	*C.MYSQL;
	lock	*sync.Mutex;
}

// The URL passed into this function should be of the form:
//
//   mysql://user:pass@host:port/database_name
//   mysql://user:pass@socket./database_name?socket=/tmp/mysql.sock
//
// Every option except for the database is optional, and missing items will be
// taken from the system-wide configuration.
//
// The scheme may be omitted:
//
//   //user:pass@host:port/database_name
//
func open(uri string) (conn db.Connection, err os.Error) {
	var host, uname, passwd, dbname, socket *C.char;
	var port C.uint;

	url, urlError := http.ParseURL(uri);
	if urlError != nil {
		err = MysqlError(fmt.Sprintf("Couldn't parse URL: %s", urlError));
		return;
	}

	if len(url.Scheme) > 0 && url.Scheme != "mysql" {
		err = MysqlError(fmt.Sprintf("Invalid scheme: %s", url.Scheme));
		return;
	}

	if len(url.Host) > 0 {
		if url.Host == "socket." {
			sock := strings.Split(url.RawQuery, "=", 2);
			if len(sock) != 2 || sock[0] != "socket" {
				err = MysqlError(
					fmt.Sprintf("Invalid socket specified: %s", url.RawQuery));
				return;
			}
			sock = strings.Split(sock[1], "&", 2);
			socket = C.CString(sock[0]);
		} else {
			hostport := strings.Split(url.Host, ":", 2);
			if len(hostport) == 2 && len(hostport[1]) > 0 {
				p, _ := strconv.Atoi(hostport[1]);
				port = C.uint(p);
			}
			if len(hostport[0]) > 0 {
				host = C.CString(hostport[0])
			}
		}
	}

	if len(url.Userinfo) > 0 {
		userpass := strings.Split(url.Userinfo, ":", 2);
		if len(userpass) == 2 && len(userpass[1]) > 0 {
			passwd = C.CString(userpass[1])
		}
		if len(userpass[0]) > 0 {
			uname = C.CString(userpass[0])
		}
	}

	if len(url.Path) > 0 {
		path := strings.Split(url.Path, "/", 2);
		if len(path) == 2 && len(path[0]) == 0 && len(path[1]) > 0 {
			dbname = C.CString(path[1])
		}
	}

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
	err = c.lastError();
	if err != nil || rc != c.handle {
		C.mysql_close(c.handle)
	} else {
		// Everything's ok, set the returned the connection
		conn = c
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
func (conn Connection) lastError() os.Error {
	if err := C.mysql_error(conn.handle); *err != 0 {
		return os.NewError(C.GoString(err))
	}
	return nil;
}

func (conn Connection) Prepare(query string) (dbs db.Statement, e os.Error) {
	s := Statement{};
	s.conn = &conn;

	conn.Lock();
	s.stmt = C.mysql_stmt_init(conn.handle);
	conn.Unlock();

	if s.stmt == nil {
		e = MysqlError("Prepare: Couldn't init statement (out of memory?)");
		return;
	}

	conn.Lock();
	cquery := strings.Bytes(query);
	if r := C.mysql_stmt_prepare(
		s.stmt, (*C.char)(unsafe.Pointer(&cquery[0])), C.ulong(len(query))); r != 0 {
		e = conn.lastError()
	} else {
		dbs = s
	}
	conn.Unlock();

	return;
}

func (conn Connection) Lock()	{ conn.lock.Lock() }
func (conn Connection) Unlock()	{ conn.lock.Unlock() }

func createParamBinds(args ...) (binds *C.MYSQL_BIND, data []BoundData, err os.Error) {
	a := reflect.NewValue(args).(*reflect.StructValue);
	fcount := a.NumField();
	if fcount > 0 {
		binds = C.mysql_bind_create_list(C.int(fcount));
		data = make([]BoundData, fcount);
		for i := 0; i < fcount; i++ {
			switch arg := a.Field(i).(type) {
			default:
				err = MysqlError(
					fmt.Sprintf("Unsupported param type %T", arg));
				break;

			case *reflect.IntValue:
				/* TODO use the native platform to do this conversion */
				data[i] = *NewBoundData(
					MysqlTypeLong,
					nil,
					0);
				if len(data[i].buffer) != 4 {
					err = MysqlError(
						fmt.Sprintf("int was not 4 bytes long, it was %d",
							len(data[i].buffer)));
					break;
				}
				v := arg.Get();
				for j := uint(0); j < 4; j += 1 {
					data[i].buffer[j] = uint8((v >> (j * 8)) & 0xff)
				}

				C.mysql_bind_assign(
					binds,
					C.uint(i),
					MysqlTypeLong,
					unsafe.Pointer(&data[i].buffer[0]),
					C.ulong(4),
					unsafe.Pointer(&data[i].blen),
					unsafe.Pointer(&data[i].is_null),
					unsafe.Pointer(&data[i].error));

			case *reflect.StringValue:
				b := strings.Bytes(arg.Get());

				data[i] = *NewBoundData(
					MysqlTypeString,
					b,
					len(b));

				C.mysql_bind_assign(
					binds,
					C.uint(i),
					MysqlTypeString,
					unsafe.Pointer(&data[i].buffer[0]),
					C.ulong(len(b)),
					unsafe.Pointer(&data[i].blen),
					unsafe.Pointer(&data[i].is_null),
					unsafe.Pointer(&data[i].error));
			}
		}
	}
	if err != nil {
		C.free(unsafe.Pointer(binds));
		binds = nil;
		data = nil;
	}
	return;
}

func createResultBinds(stmt *C.MYSQL_STMT) (*C.MYSQL_BIND, *[]BoundData) {
	meta := C.mysql_stmt_result_metadata(stmt);
	if meta != nil {
		fcount := C.mysql_num_fields(meta);
		binds := C.mysql_bind_create_list(C.int(fcount));
		data := make([]BoundData, fcount);
		for i := C.uint(0); i < fcount; i += 1 {
			field := C.mysql_fetch_field_direct(meta, i);

			data[i] = *NewBoundData(
				MysqlType(field._type),
				nil,
				int(field.length));

			C.mysql_bind_assign(
				binds, i,
				field._type,
				unsafe.Pointer(&data[i].buffer[0]),
				field.length,
				unsafe.Pointer(&data[i].blen),
				unsafe.Pointer(&data[i].is_null),
				unsafe.Pointer(&data[i].error));
		}
		C.mysql_free_result(meta);
		return binds, &data;
	}
	return nil, nil;
}

func (conn Connection) execute(stmt db.Statement, parameters ...) (dbcur *cursor, err os.Error) {

	dbcur = nil;
	if s, ok := stmt.(Statement); ok {
		var (
			binds	*C.MYSQL_BIND	= nil;
			data	[]BoundData;
			e	os.Error;
		)
		pcount := uint64(C.mysql_stmt_param_count(s.stmt));

		if pcount > 0 {
			if binds, data, e = createParamBinds(parameters); e == nil {
				if rc := C.mysql_stmt_bind_param(s.stmt, binds); rc != 0 {
					err = conn.lastError();
					goto cleanup;
				}
			} else {
				err = e;
				goto cleanup;
			}
			// prevent data no-use errors.  We just need to keep it around so
			// that GC doesn't clean it up.
			data = data;
		}

		conn.Lock();
		if rc := C.mysql_stmt_execute(s.stmt); rc != 0 {
			err = conn.lastError()
		} else {
			// Must call store result before unlocking...
			if rc := C.mysql_stmt_store_result(s.stmt); rc != 0 {
				err = conn.lastError()
			} else {
				dbcur = NewCursorValue(s)
			}
		}

	cleanup:
		if binds != nil {
			C.free(unsafe.Pointer(binds))
		}
		conn.Unlock();
	} else {
		err = MysqlError("Execute: 'stmt' is not a mysql.Statement")
	}

	return;
}

func returnResults(dc *cursor, ch chan db.Result) {
	r, e := dc.Fetch();
	for ; !closed(ch) && r != nil && e == nil; r, e = dc.Fetch() {
		ch <- Result{r, nil}
	}
	if e != nil {
		ch <- Result{nil, e}
	}
	close(ch);
	dc.Close();
}

func (conn Connection) Execute(stmt db.Statement, parameters ...) (ch <-chan db.Result, err os.Error) {
	var dc *cursor;
	dc, err = conn.execute(stmt, parameters);
	if err != nil {
		ch = nil;
		return;
	}
	sendch := make(chan db.Result);
	go returnResults(dc, sendch);
	ch = sendch;
	return;
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
		s.conn.Lock();
		if r := C.mysql_stmt_close(s.stmt); r != 0 {
			err = s.conn.lastError()
		}
		s.stmt = nil;
		s.conn.Unlock();
	}
	return nil;
}

type cursor struct {
	stmt	*Statement;
	rbinds	*C.MYSQL_BIND;
	rdata	*[]BoundData;
	bound	bool;
}

func NewCursorValue(s Statement) *cursor {
	cur := &cursor{};
	cur.stmt = &s;
	cur.bound = false;
	cur.setupResultBinds();
	return cur;
}

func (c *cursor) setupResultBinds() (err os.Error) {
	if c.bound {
		return
	}
	c.rbinds, c.rdata = createResultBinds(c.stmt.stmt);
	if c.rbinds != nil {
		if rc := C.mysql_stmt_bind_result(c.stmt.stmt, c.rbinds); rc != 0 {
			err = c.stmt.conn.lastError()
		}
	}
	c.bound = true;
	return;
}

func (c *cursor) Fetch() (res []interface{}, err os.Error) {
	if rc := C.mysql_stmt_fetch(c.stmt.stmt); rc == 0 {
		res = make([]interface{}, len(*c.rdata));
		rdata := *c.rdata;
		for i := range (rdata) {
			res[i], _ = rdata[i].Value()
		}
	} else if rc == 100 {
		// no data
	} else {
		err = c.stmt.conn.lastError()
	}
	return;
}

func (c *cursor) Close() (err os.Error) {
	if c.rbinds != nil {
		c.stmt.conn.Lock();
		C.mysql_bind_free(c.rbinds);
		c.stmt.conn.Unlock();
		c.rbinds = nil;
	}
	c.rdata = nil;
	c.bound = false;
	return;
}

type Result struct {
	data	[]interface{};
	error	os.Error;
}

func (r Result) Data() []interface{}	{ return r.data }
func (r Result) Error() os.Error	{ return r.error }
