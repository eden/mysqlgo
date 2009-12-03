// Copyright 2009 Eden Li. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Example usage of mysqlgo.

package main

import (
	"mysql";
	"flag";
	"rand";
	"fmt";
	"os";
	"db";
)

var (
	host	string;
	port	int;
	user	string;
	pass	string;
	dbname	string;
	help	bool;
)

func init() {
	flag.StringVar(&host, "host", "127.0.0.1",
		"Connect to this MySQL host");
	flag.IntVar(&port, "port", 3306,
		"Connect on this port");
	flag.StringVar(&user, "user", "root",
		"Connect with this username");
	flag.StringVar(&pass, "pass", "",
		"Connect with this password");
	flag.StringVar(&dbname, "database", "test", "Default database");
	flag.BoolVar(&help, "help", false, "Print this help message and quit");
}

func main() {
	flag.Parse();
	if help { flag.Usage(); os.Exit(1) }

	conn, err := mysql.Open(map[string]interface{} {
		"host": host,
		"port": port,
		"user": user,
		"pass": pass,
		"database": dbname
	});
	fconn := conn.(db.FancyConnection);

	if err != nil {
		fmt.Printf("Error connecting to %s:%d: %s\n",
			host, port, err);
		flag.Usage();
		os.Exit(1)
	}
	var stmt db.Statement;

	fmt.Println("Creating temporary table __hello");
	_, e := fconn.ExecuteDirectly("CREATE TEMPORARY TABLE __hello (i INT)");
	if e != nil { fmt.Printf("Error: %s", err); os.Exit(1); }

	fmt.Println("Inserting 100 random numbers");
	stmt, e = conn.Prepare("INSERT INTO __hello (i) VALUE (?)");
	if e != nil { fmt.Printf("Error: %s\n", err); os.Exit(1); }

	for i := 0; i < 100; i+=1 {
		_, err = conn.Execute(stmt, rand.Int());
		if err != nil { fmt.Printf("Error: %s\n", err); os.Exit(1); }
	}
	stmt.Close();

	fmt.Println("Reading numbers in lexical order");
	stmt, e = conn.Prepare("SELECT i FROM __hello ORDER BY i ASC");
	if e != nil { fmt.Printf("Error: %s\n", err); os.Exit(1); }

	cur, _ := conn.Execute(stmt);
	for t, _ := cur.FetchOne(); t != nil; t, _ = cur.FetchOne() {
		if v, ok := t[0].(int); ok {
			fmt.Printf("%d\n", v)
		} else {
			fmt.Printf("Error converting %T to int\n", v)
		}
	}
	cur.Close();
	stmt.Close();

	conn.Close();
}
