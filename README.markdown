MySQL Bindings for Go (golang)
==============================

Implements rudimentary MySQL support for Go via libmysql.  The interface
follows Peter Froehlich's [database
interface](http://github.com/phf/go-sqlite/blob/master/db.go).  This is
automatically included via a git submodule.

Currently, it is possible to share a single connection with multiple
goroutines.  Note, however, that locks are used for certain libmysql calls due
to the thread-sensitvity of those calls.

The `Makefile` assumes `mysql_config` is in your path.

Install
-------

    cd mysqlgo
    make install

Example
-------

    cd mysqlgo
    make example
    ./example -host=127.0.0.1 -user=root -dbname=test

Synopsis
--------

    conn, err := mysql.Open(map[string]interface{} {
      "host": "127.0.0.1",
      "port": 3306,
      "username": "root"
    });
    if err != nil { panic("Connect error:", err) }

	fconn := conn.(db.FancyConnection);

    cur, serr := fconn.ExecuteDirectly("SELECT * FROM table");
    if serr != nil { panic("ExecuteDirectly error:", serr) }

    for t, _ := cur.FetchOne(); t != nil; t, _ = cur.FetchOne() {
      fmt.Println(t)
    }

    cur.Close();
    conn.Close();

TODO
====

 * Better parameter type support (right now only int and strings can be bound
   as parameters in `Prepare` and `ExecuteDirectly`).
 * `DATE`, `TIME` and `DATETIME` support.
 * Implement `TransactionalConnection` methods
 * Implement `InformativeCursor` methods
 * Implement `PythonicCursor` methods
 * Implement `FetchMany` and `FetchAll`
 * More exhaustive testing.  Most of the main methods are tested, but the test
   code needs some refactoring for clarity.
