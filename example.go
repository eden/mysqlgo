package main

import "mysql"
import "fmt"
import "os"

func main() {
	conn := mysql.NewConn();
	err := conn.Connect(&mysql.ConnInfo{
		"host", 3306, "user", "pass", "db",
	});

	if err != nil {
		fmt.Printf("Error: %s\n", err);
		os.Exit(1)
	}

	cur := conn.Cursor();
	cur.Execute("SELECT * FROM objects WHERE field < %i",
		100);

	tuple, err := cur.FetchOne();
	for ; tuple != nil; tuple, err = cur.FetchOne() {
		fmt.Println(tuple)
	}
	if err != nil {
		fmt.Printf("Encountered error: %s\n", err.String())
	}
	cur.Close();
	conn.Close();
}
