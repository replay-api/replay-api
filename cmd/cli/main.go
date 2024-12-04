package main

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"

// 	// migrate "github.com/golang-migrate/migrate/v4"
// 	// _ "github.com/golang-migrate/migrate/v4/database/postgres"
// 	// _ "github.com/golang-migrate/migrate/v4/source/file"
// )

// func main() {
// 	cmd := os.Args[1]

// 	log.Printf("Running migration command: %s", cmd)

// 	if cmd == "newdb" {
// 		newDB()

// 		return
// 	}

// 	m, err := migrate.New(
// 		"file://migrations",
// 		"postgres://postgres:postgres@localhost:9432/replay_api?sslmode=disable") // &sslmode=enable // in pq: $ psql -h localhost -p 9432 -U postgres -d replay_api
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if err := m.Up(); err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Printf("Migrator initialized!")

// 	if cmd == "up" {
// 		err = up(m)
// 	} else if cmd == "down" {
// 		err = down(m)
// 	} else {
// 		log.Fatal("Invalid command")
// 	}

// 	if err != nil {
// 		panic(err)
// 	}
// }

// func up(m *migrate.Migrate) error {
// 	return m.Up()
// }

// func down(m *migrate.Migrate) error {
// 	return m.Down()
// }

// func newDB() {
// 	connStr := "user=postgres password=postgres sslmode=disable host=localhost port=9432"
// 	db, err := sql.Open("postgres", connStr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	_, err = db.Exec("CREATE DATABASE replay_api;")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS replay_api;")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("Database created successfully!")
// }
