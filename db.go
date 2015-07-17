package main

import (
	"database/sql"
	"fmt"
	"gopkg.in/gorp.v1"
	log "gopkg.in/inconshreveable/log15.v2"
)

var dbmap *gorp.DbMap
var db *sql.DB

// Initializes database configuration.
// Sets up the psuedo-orm's table mapping and creates tables if not already existing.
func initDb() {

	if dbmap != nil {
		log.Info("Reusing existing database connection")
		return
	}

	var err error

	log.Info("Initializing database connection")

	db, err = sql.Open("mysql", dsn)
	db.SetMaxOpenConns(dbMaxConn)

	fatalErr(err, "Could not connect to database")

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

	dbmap.AddTableWithName(Broker{}, "mqtt_logger_config").SetKeys(true, "Id")
	dbmap.AddTableWithName(Topic{}, "mqtt_logger_topics").SetKeys(true, "Id")
	dbmap.AddTableWithName(Message{}, "mqtt_logger_messages").SetKeys(true, "Id")
	err = dbmap.CreateTablesIfNotExists()
	fatalErr(err, "Could not create database tables")

	log.Info(fmt.Sprintf("Connected successfully to %s", dsn))

}
