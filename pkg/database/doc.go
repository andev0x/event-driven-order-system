// Package database provides shared database initialization and utilities.
//
// This package contains common database connection setup and configuration
// that can be reused across different services. It supports MySQL connections
// with configurable connection pooling.
//
// Example usage:
//
//	cfg := database.Config{
//	    Host:     "localhost",
//	    Port:     "3306",
//	    User:     "root",
//	    Password: "secret",
//	    Name:     "mydb",
//	}
//	db, err := database.Connect(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
package database
