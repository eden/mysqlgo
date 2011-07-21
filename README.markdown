MySQL Bindings for Go (golang)
==============================
![Abandoned](http://stillmaintained.com/eden/mysqlgo.png)

Implements MySQL support for Go via libmysql.  The interface implements Peter
Froehlich's [database interface](http://github.com/phf/go-db).  This is
automatically included via a git submodule.

Currently, it is possible to share a single connection with multiple
goroutines.  Note, however, that locks are used for certain libmysql calls due
to the thread-sensitvity of those calls.

The `Makefile` assumes `mysql_config` is in your path.

Synopsis
--------

    conn, err := mysql.Open("mysql://root@127.0.0.1:3306/test");
    if err != nil { panic("Connect error:", err) }

    stmt, serr := conn.Prepare("SELECT * FROM table WHERE name LIKE ?");
    if serr != nil { panic("Prepare error:", serr) }

    rs, rerr := conn.Execute(stmt, "George%");
    if rerr != nil { panic("Execute error:", rerr) }

    for result := range rs.Iter() {
      data := result.Data();
      fmt.Println(data)
    }

    conn.Close();

Install/Run Example
-------------------

    $ git clone git://github.com/eden/mysqlgo.git
    $ cd mysqlgo
    $ make install
    $ make example
    $ ./example -host=127.0.0.1 -user=root -dbname=test

TODO
====

Most basic operations (execute queries, get results) are implemented and
somewhat tested.  The major things left are:

 * Better parameter type support (right now only int and strings can be bound
   as parameters in `Iterate` and `Execute`).
 * `DATE`, `TIME` and `DATETIME` support.
 * Implement `TransactionalConnection` methods
 * More exhaustive testing.  Most of the main methods are tested, but the test
   code needs some refactoring for clarity.
 * Cleanup internals.
