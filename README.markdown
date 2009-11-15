MySQL Bindings for Go (golang)
==============================

Implements rudimentary MySQL support for Go.

`mw.{c,h}` is used to wrap mysql since `cgo` currently can't translate mysql
header files due to the following issues.

- http://code.google.com/p/go/issues/detail?id=126
- http://code.google.com/p/go/issues/detail?id=36

The interface vaguely follows [Python's PEP
249](http://www.python.org/dev/peps/pep-0249/)

Synopsis
--------

    conn := mysql.NewConn();
    err := conn.Connect(&mysql.ConnInfo{host, port, user, pass, dbname});
    if err != nil { panic("Connect error") }

    cur := conn.Cursor();
    cur.Execute("SELECT * FROM table");
    tuple, err := cur.FetchOne();
    for ; err == nil && tuple != nil; tuple, err = cur.FetchOne() {
      fmt.Println(tuple)
    }

    cur.Close();
    conn.Close();

Installing and Running Example
------------------------------

    make install
    make example
    ./example -host=127.0.0.1 -user=root -dbname=test

TODO
----
- Proper type conversion.  Right now all values are returned as strings.
- More tests.
- Better documentation.
- Unwrap MySQL bits once `cgo` is fixed.
- Connection pools and thread testing.
