// Package redis provides shared Redis client initialization and utilities.
//
// This package contains common Redis connection setup that can be reused
// across different services for caching purposes.
//
// Example usage:
//
//	cfg := redis.Config{
//	    Host: "localhost",
//	    Port: "6379",
//	}
//	client, err := redis.Connect(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
package redis
