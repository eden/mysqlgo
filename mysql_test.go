// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test package for mysql.  Requires a mysql server to be running locally with
// a user `root` with a blank password and a database called `test`.
package mysql_test

import "testing"
import "mysql"
import "rand"
import "os"

func defaultConn(t *testing.T) *mysql.Conn {
	conn := mysql.NewConn();
	err := conn.Connect(&mysql.ConnInfo{
		"127.0.0.1", 3306,
		"root", "",
		"test"
	});
	if err != nil {
		t.Log("Couldn't connect to root:@127.0.0.1:3306:test\n%s",
			err);
		return nil;
	}
	return conn
}

var tableT = []string{
	// Uncomment once we have fixed sql quoting logic.
	// Currently, it just uses fmt.Sprintf and relies on %q which
	// mysql does not interpret properly.

	/*"道可道，非常道。", "名可名，非常名。",
	"無名天地之始；", "有名萬物之母。",
	"故常無欲以觀其妙；", "常有欲以觀其徼。",
	"此兩者同出而異名，", "同謂之玄。",
	"玄之又玄，眾妙之門。", */

	"lorem",
	"ipsum",
	"dolor",
	"sit",
	"amet"
	"consectetur",
	"adipisicing",
	"elit",
	"sed"
}

func prepareTestTable(t *testing.T, cur *mysql.Cursor) {
	err := cur.Execute("CREATE TEMPORARY TABLE t (i INT, s VARCHAR(20));");
	if err != nil {
		error(t, err, "Couldn't create temporary table test.t")
	}
	for i, s := range tableT {
		err = cur.Execute("INSERT INTO t (i, s) VALUES (%d, %q)",
			i, s);
		if err != nil {
			error(t, err, "Couldn't insert into temporary table test.t")
		}
		if count := cur.RowCount(); count != 0 {
			t.Error("Returned rows for INSERT statement.")
		}
	}
}

func startTestWithLoadedFixture(t *testing.T)
	(conn *mysql.Conn, cur *mysql.Cursor)
{
	conn = defaultConn(t);
	if conn == nil { return }

	cur = conn.Cursor();
	if cur != nil { prepareTestTable(t, cur) }

	return;
}

func error(t *testing.T, err os.Error, msg string) {
	if err == nil {
		t.Error(msg)
	}
	else {
		t.Errorf("%s: %s\n", msg, err.String());
	}
}

func TestCursorErrors(t *testing.T) {
	// Cursor on an unconnected object returns nil
	conn := mysql.NewConn();
	if conn.Cursor() != nil {
		error(t, nil, "Unconnected Cursor should be nil")
	}

	conn = defaultConn(t);
	if conn == nil { return }

	cur := conn.Cursor();

	// Fetch called before Execute should return nil with an error
	res, err := cur.FetchOne();
	if res != nil || err == nil {
		error(t, nil, "FetchOne before Execute should error")
	}

	resm, err := cur.FetchMany(10);
	if resm != nil || err == nil {
		error(t, nil, "FetchMany(10) before Execute should error")
	}

	resm, err = cur.FetchAll();
	if resm != nil || err == nil {
		error(t, nil, "FetchAll before Execute should error")
	}

	// Invalid statements return errors
	err = cur.Execute("1");
	if res != nil || err == nil {
		error(t, nil, "Invalid statement should return errors")
	}

	// No result statements should not error, but should not have results.
	err = cur.Execute("# No results");
	if err != nil {
		error(t, err, "No-result SQL statement should not error")
	}

	res, err = cur.FetchOne();
	if res != nil || err == nil {
		error(t, nil, "FetchOne on no-result statement should error")
	}

	resm, err = cur.FetchMany(10);
	if res != nil || err == nil {
		error(t, nil, "FetchMany on no-result statement should error")
	}

	resm, err = cur.FetchAll();
	if res != nil || err == nil {
		error(t, nil, "FetchAll on no-result statement should error")
	}

	cols := cur.Description();
	if len(cols) > 0 {
		t.Error("Description should return no 0-length columns")
	}

	conn.Close();
	cur.Close();
}

func TestCursor(t *testing.T) {
	conn, cur := startTestWithLoadedFixture(t);

	err := cur.Execute("SELECT i AS pos, s AS phrase FROM t ORDER BY pos ASC");
	if err != nil {
		error(t, err, "Couldn't select from temporary table test.t")
	}
	if count := cur.RowCount(); int(count) != len(tableT) {
		t.Error("Result count doesn't match inserted count.")
	}

	cols := cur.Description();
	if len(cols) != 2 { t.Error("Description should return 2 columns") }
	if cols[0].Name != "pos" { t.Error("Description()[0] != 'pos'") }
	if cols[1].Name != "phrase" { t.Error("Description()[0] != 'phrase'") }

	i := 0;
	var row []interface {};
	row, err = cur.FetchOne();
	for row != nil {
		if err != nil { error(t, err, "Couldn't FetchOne()") }
		if v, ok := row[1].(string); !ok || tableT[i] != v {
			if ok { t.Errorf("Mismatch %q != %q", tableT[i], v) }
			else { t.Errorf("Couldn't convert %v to string.", row[1]) }
		}
		i += 1;
		row, err = cur.FetchOne();
	}

	// Test FetchMany
	err = cur.Execute("SELECT i AS pos, s AS phrase FROM t ORDER BY pos ASC");
	if err != nil {
		error(t, err, "Couldn't select from temporary table test.t")
	}

	var results [][]interface {};
	results, err = cur.FetchMany(3);
	if err != nil { error(t, err, "Error") }
	if len(results) != 3 {
		t.Errorf("Result count mismatch 3 != %d", len(results))
	}
	for i, v := range results {
		if v, ok := v[1].(string); !ok || tableT[i] != v {
			if ok { t.Errorf("Mismatch %q != %q", tableT[i], v) }
			else { t.Errorf("Couldn't convert %v to string.", row[1]) }
		}
	}

	// Test FetchAll
	err = cur.Execute("SELECT i AS pos, s AS phrase FROM t ORDER BY pos ASC");
	if err != nil {
		error(t, err, "Couldn't select from temporary table test.t")
	}

	results, err = cur.FetchAll();
	if err != nil { error(t, err, "Error") }

	if len(results) != len(tableT) {
		t.Errorf("Result count mismatch %d != %d", len(tableT), len(results))
	}
	for i, v := range results {
		if v, ok := v[1].(string); !ok || tableT[i] != v {
			if ok { t.Errorf("Mismatch %q != %q", tableT[i], v) }
			else { t.Errorf("Couldn't convert %v to string.", row[1]) }
		}
	}

	cur.Close();
	conn.Close();
}

func findRand(t *testing.T, conn *mysql.Conn, ch chan [][]interface {}) {
	cur := conn.Cursor();
	err := cur.Execute(
		"SELECT * FROM t WHERE i != %d ORDER BY RAND()",
		rand.Int());
	if err != nil { error(t, err, "Couldn't select") }

	var res [][]interface{};
	res, err = cur.FetchAll();
	if err != nil { error(t, err, "Couldn't fetch") }
	if len(res) != len(tableT) { t.Error("Invalid length") }

	cur.Close();
	ch <- res
}

func TestReentrant(t *testing.T) {
	conn, cur := startTestWithLoadedFixture(t);
	cur.Close();

	ch := make([]chan [][]interface {}, 500);

	for i, _ := range ch {
		ch[i] = make(chan [][]interface{});
		go findRand(t, conn, ch[i])
	}
	for _, c := range ch {
		res := <-c;
		if len(res) != len(tableT) {
			t.Error("Invalid results")
		}
	}

	conn.Close();
}