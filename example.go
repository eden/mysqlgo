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

func errorAndQuit(err os.Error) {
	fmt.Printf("Error: %s\n", err.String());
	os.Exit(1);
}

func main() {
	flag.Parse();
	if help {
		flag.Usage();
		os.Exit(1);
	}

	conn, err := mysql.Open(fmt.Sprintf("mysql://%s:%s@%s:%d/%s",
		user,
		pass,
		host,
		port,
		dbname));

	if err != nil {
		fmt.Printf("Error connecting to %s:%d: %s\n",
			host, port, err);
		flag.Usage();
		os.Exit(1);
	}
	var stmt db.Statement;

	// Create table
	fmt.Printf("Creating temporary table %s.__hello\n", dbname);

	stmt, e := conn.Prepare(
		"CREATE TEMPORARY TABLE __hello (i INT, s VARCHAR(255))");
	if e != nil {
		errorAndQuit(e)
	}

	_, e = conn.Execute(stmt);
	if e != nil {
		errorAndQuit(e)
	}

	// Populate table
	fmt.Println("Inserting 100 random numbers");

	stmt, e = conn.Prepare("INSERT INTO __hello (i, s) VALUE (?, ?)");
	if e != nil {
		errorAndQuit(e)
	}

	for i := 0; i < 100; i += 1 {
		_, e = conn.Execute(stmt, rand.Int(), fmt.Sprintf("id%d", rand.Int()));
		if e != nil {
			errorAndQuit(e)
		}
	}
	stmt.Close();

	// Read from table
	fmt.Println("Reading numbers in numeric order");

	stmt, e = conn.Prepare(
		"SELECT i, s FROM __hello WHERE s LIKE ? ORDER BY i ASC");
	if e != nil {
		errorAndQuit(e)
	}

	rs, cErr := conn.Execute(stmt, "id%");
	if cErr != nil {
		errorAndQuit(cErr)
	}

	for res := range rs.Iter() {
		row := res.Data();

		if v, ok := row[0].(int); ok {
			fmt.Printf("%d %v\n", v, row[1])
		} else {
			fmt.Printf("Error converting %T to int\n", v)
		}
	}
	stmt.Close();
	conn.Close();
}
