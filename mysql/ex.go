// ex - simple example

package main

import ("mysql"; "fmt")

func main() {
	conn, e := mysql.Open(map[string]interface{} {
		"host": "localhost",
		"port": 3306,
		"username": "root",
		"database": "test"
	});

	if e == nil {
		if stmt, e := conn.Prepare("SELECT * FROM mysql.user"); e == nil {
			if cur, err := conn.Execute(stmt); err == nil {
				for res, _ := cur.FetchOne(); res != nil; res, _ = cur.FetchOne() {
					fmt.Printf("%#v\n", res);
				}
				cur.Close();
			}
			stmt.Close();
		}
		else {
			fmt.Printf("%s\n", e);
		}
		conn.Close();
	}
	else {
		fmt.Printf("%s\n", e);
	}
}
