MySQL Bindings for Go (golang)
==============================

Implements rudimentary MySQL support for Go via libmysql.  The interface
vaguely follows [Python's PEP 249](http://www.python.org/dev/peps/pep-0249/).
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

    stmt, serr := conn.Prepare("SELECT * FROM table");
    if serr != nil { panic("Prepare error:", serr) }

    cur, cerr := conn.Execute(stmt);
    if cerr != nil { panic("Execute error:", cerr) }

    for t, _ := cur.FetchOne(); t != nil; t, _ = cur.FetchOne() {
      fmt.Println(t)
    }

    cur.Close();
    stmt.Close();
    conn.Close();
