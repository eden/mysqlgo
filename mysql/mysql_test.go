// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test package for mysql.  Requires a mysql server to be running locally with
// a user `root` with a blank password and a database called `test`.
package mysql_test

import (
	"container/vector";
	"testing";
	"mysql";
	"rand";
	"db";
	"os";
)

func defaultConn(t *testing.T) *db.Connection {
	conn, e := mysql.Open("//root@localhost:3306/test");
	if conn == nil || e != nil {
		t.Error("Couldn't connect to root@127.0.0.1:3306:test", e);
		return nil;
	}
	return &conn;
}

var tableT = []string{
	"道可道，非常道。", "名可名，非常名。",
	"無名天地之始；", "有名萬物之母。",
	"故常無欲以觀其妙；", "常有欲以觀其徼。",
	"此兩者同出而異名，", "同謂之玄。",
	"玄之又玄，眾妙之門。",
	"test",
	"test2",
	"test3",
	"test4",
	"test5",
}

func prepareTestTable(t *testing.T, conn *db.Connection) {
	stmt, sErr := conn.Prepare(
		"CREATE TEMPORARY TABLE t (i INT, s VARCHAR(100));");
	if sErr != nil {
		error(t, sErr, "Couldn't prepare statement");
		return;
	}

	ch, cErr := conn.Execute(stmt);
	if cErr != nil {
		error(t, cErr, "Couldn't create temporary table test.t");
		return;
	}
	close(ch);

	stmt, sErr = conn.Prepare("INSERT INTO t (i, s) VALUES (?, ?)");
	if sErr != nil {
		error(t, sErr, "Couldn't prepare statement");
		return;
	}

	for i, s := range tableT {
		ch, cErr = conn.Execute(stmt, i, s);
		if cErr != nil {
			error(t, cErr, "Couldn't insert into temporary table test.t");
			return;
		}
		close(ch);
	}
	stmt.Close();
}

func startTestWithLoadedFixture(t *testing.T) (conn *db.Connection) {
	conn = defaultConn(t);
	if conn == nil {
		return
	}

	prepareTestTable(t, conn);
	return;
}

func error(t *testing.T, err os.Error, msg string) {
	if err == nil {
		t.Error(msg)
	} else {
		t.Errorf("%s: %s\n", msg, err.String())
	}
}

func TestOne(t *testing.T) {
	conn := startTestWithLoadedFixture(t);
	if conn == nil { t.Error("conn was nil"); return }

	stmt, sErr := conn.Prepare(
		"SELECT i AS pos, s AS phrase FROM t ORDER BY pos ASC");
	if sErr != nil {
		error(t, sErr,
			"Couldn't prepare for select from temporary table test.t")
	}
	ch, cErr := conn.Execute(stmt);
	if cErr != nil {
		error(t, cErr, "Couldn't execute statement")
	}

	i := 0;
	for res := range(ch) {
		if res.Error() != nil {
			error(t, res.Error(), "Couldn't fetch from channel")
		}

		row := res.Data();
		if v, ok := row[0].(int); !ok || i != v {
			if ok {
				t.Errorf("Mismatch %d != %d", i, v)
			} else {
				t.Errorf("Couldn't convert %T to int.", row[0])
			}
		}
		if v, ok := row[1].(string); !ok || tableT[i] != v {
			if ok {
				t.Errorf("Mismatch %q != %q", tableT[i], v)
			} else {
				t.Errorf("Couldn't convert %T to string.", row[1])
			}
		}
		i += 1;
	}

	stmt.Close();
	conn.Close();
}

func prepareEmpty(t *testing.T, conn *db.Connection, ch chan int) {
	stmt, sErr := conn.Prepare(
		"SELECT * FROM t ORDER BY RAND()");
	if sErr != nil {
		error(t, sErr, "Couldn't prepare")
	}
	stmt.Close();
	ch <- 1;
}

func TestReentrantPrepare(t *testing.T) {
	conn := startTestWithLoadedFixture(t);
	if conn == nil { t.Error("conn was nil"); return }

	ch := make([]chan int, 100);

	for i, _ := range ch {
		ch[i] = make(chan int);
		go prepareEmpty(t, conn, ch[i]);
	}
	for _, c := range ch {
		<-c;
	}

	conn.Close();
}

func execute(t *testing.T, conn *db.Connection, stmt *db.Statement, ch chan int) {
	c, cErr := conn.Execute(*stmt, rand.Int());
	if cErr != nil {
		error(t, cErr, "Couldn't select")
	}
	for res := range(c) {
		if res.Error() != nil {
			error(t, res.Error(), "Couldn't fetch")
		}
	}
	ch <- 1;
}

func TestReentrantExecute(t *testing.T) {
	conn := startTestWithLoadedFixture(t);
	if conn == nil { t.Error("conn was nil"); return }

	stmt, sErr := conn.Prepare(
		"SELECT * FROM t ORDER BY RAND()");
	if sErr != nil {
		error(t, sErr, "Couldn't prepare")
	}

	ch := make([]chan int, 10);

	for i, _ := range ch {
		ch[i] = make(chan int);
		go execute(t, conn, &stmt, ch[i]);
	}
	for _, c := range ch {
		<-c
	}

	stmt.Close();
	conn.Close();
}

func findRand(t *testing.T, conn *db.Connection, ch chan *vector.Vector) {
	stmt, sErr := conn.Prepare(
		"SELECT * FROM t WHERE i != ? ORDER BY RAND()");
	if sErr != nil {
		error(t, sErr, "Couldn't prepare")
	}

	c, cErr := conn.Execute(stmt, rand.Int());
	if cErr != nil {
		error(t, cErr, "Couldn't select")
	}

	vout := new(vector.Vector);
	for res := range(c) {
		if res.Error() != nil {
			error(t, res.Error(), "Couldn't fetch")
		}
		vout.Push(res.Data());
	}

	if vout.Len() != len(tableT) {
		t.Error("Invalid length")
	}

	stmt.Close();
	ch <- vout;
}

func TestPrepareExecuteReentrant(t *testing.T) {
	for j := 0; j < 10; j++ {
		conn := startTestWithLoadedFixture(t);
		if conn == nil { t.Error("conn was nil"); return }

		ch := make([]chan *vector.Vector, 100);

		for i, _ := range ch {
			ch[i] = make(chan *vector.Vector);
			go findRand(t, conn, ch[i]);
		}
		for _, c := range ch {
			res := <-c;
			if res.Len() != len(tableT) {
				t.Error("Invalid results")
			}
		}

		conn.Close();
	}
}

func TestChannelInterface(t *testing.T) {
	con := startTestWithLoadedFixture(t);
	if con == nil { t.Error("conn was nil"); return }
	conn := *con;

	stmt, sErr := conn.Prepare(
		"SELECT ?, i AS pos, s AS phrase FROM t ORDER BY pos ASC");
	if sErr != nil { error(t, sErr, "Couldn't Prepare") }

	ch, err := conn.Execute(stmt, 123);
	if err != nil { error(t, err, "Couldn't Execute") }

	i := 0;
	for r := range ch {
		var pos int;
		row := r.Data();

		if i := row[0].(int); i != 123 {
			t.Error("Invalid parameter bind in result");
		}
		if pos = row[1].(int); pos < 0 || pos >= len(tableT) {
			t.Error("Invalid result bind pos (1)");
		}
		else {
			if str := row[2].(string); tableT[pos] != str {
				t.Error("Invalid result bind phrase (2)",
					str, "!=", tableT[pos]);
			}
		}
		i += 1
	}
	conn.Close();
}

func TestChannelInterfacePrematureClose(t *testing.T) {
	con := startTestWithLoadedFixture(t);
	if con == nil { t.Error("conn was nil"); return }
	conn := *con;

	execOne := func() {
		stmt, sErr := conn.Prepare(
			"SELECT ?, i AS pos, s AS phrase FROM t ORDER BY pos ASC");
		if sErr != nil { error(t, sErr, "Couldn't Prepare") }

		ch, err := conn.Execute(stmt, 123);
		if err != nil { error(t, err, "Couldn't Execute") }

		r := <-ch;
		row := r.Data();

		if i := row[0].(int); i != 123 {
			t.Error("Invalid parameter bind in result");
		}
		if pos := row[1].(int); pos >= 0 && pos < len(tableT) {
			if str := row[2].(string); tableT[pos] != str {
				t.Error("Invalid result bind phrase (2)",
					str, "!=", tableT[pos]);
			}
		}
		else {
			t.Error("Invalid result bind pos (1)");
		}
		close(ch);
	};

	for i := 0; i < 1000; i += 1 { execOne() }

	conn.Close();
}
