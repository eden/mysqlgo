package main

import (
	"mysql";
	"flag";
	"rand";
	"fmt";
	"os";
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

	conn := mysql.NewConn();
	err := conn.Connect(&mysql.ConnInfo{host, port, user, pass, dbname});

	if err != nil {
		fmt.Printf("Error connecting to %s:%d: %s\n",
			host, port, err);

		flag.Usage();

		os.Exit(1)
	}

	cur := conn.Cursor();

	fmt.Println("Creating temporary table __hello");
	err = cur.Execute("CREATE TEMPORARY TABLE __hello (i INT)");
	if err != nil { fmt.Printf("Error: %s", err); os.Exit(1); }

	fmt.Println("Inserting 100 random ints");
	for i := 0; i < 100; i+=1 {
		err = cur.Execute("INSERT INTO __hello (i) VALUE (%d)",
			rand.Int());
		if err != nil { fmt.Printf("Error: %s", err); os.Exit(1); }
	}

	fmt.Println("Reading ints in order");
	cur.Execute("SELECT i FROM __hello ORDER BY i ASC");
	tuple, err := cur.FetchOne();
	for ; tuple != nil; tuple, err = cur.FetchOne() {
		fmt.Printf("%s\n", tuple[0])
	}

	cur.Close();
	conn.Close();
}
