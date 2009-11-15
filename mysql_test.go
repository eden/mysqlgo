
package mysql_test

import "testing"
import "mysql"

func TestConnectFails(t *testing.T) {
	conn := mysql.NewConn();
	err := conn.Connect(&mysql.ConnInfo{
		"127.0.0.1", 2323,
		"user", "pass",
		"_no_exist"
	});
	if err == nil { t.Error("err == nil") }
}

func TestSelect(t *testing.T) {
	conn := mysql.NewConn();
	err := conn.Connect(&mysql.ConnInfo{
		"127.0.0.1", 3306,
		"root", "",
		"test"
	});
	if err != nil {
		t.Log("No local database found, skipping.");
		return;
	}
	cur := conn.Cursor();
	cur.Execute("SELECT 1, 'two', NOW()");

	if row, err := cur.FetchOne(); err == nil {
		if f, ok := row[0].(string); !ok || f != "1" { t.Fail() }
		if f, ok := row[1].(string); !ok || f != "two" { t.Fail() }
	}
	else {
		t.Error("Error fetching from db: " + err.String());
	}
}
