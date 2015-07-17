package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"time"
)

var version float64
var serverName string
var intervalDur string
var dsn string
var dbMaxConn int
var listenStr string
var loggerID int64
var cfg Config

func init() {

	version = 0.2
	serverName = "Golang MQTT Client Logger"

	flag.IntVar(&dbMaxConn, "connections", 10, "The maxiumum number of open database connections.")
	flag.StringVar(&intervalDur, "interval", "10s", "Interval to reload configuration, in seconds")
	flag.StringVar(&listenStr, "listen", "127.0.0.1:3000", "The IP address/port to listen on.")
	flag.Int64Var(&loggerID, "logger-id", 1, "The ID of the logger")

	dbProto := flag.String("db-protocol", "tcp", "The database protocol. Either 'unix' or 'tcp'.")
	dbName := flag.String("db-name", "mqttlogger", "The database name.")
	dbHost := flag.String("db-host", "127.0.0.1", "The database server.")
	dbUser := flag.String("db-user", "root", "The database user")
	dbPass := flag.String("db-pass", "", "The database password.")
	dbPort := flag.String("db-port", "3306", "The database port. Only if using 'tcp'.")

	flag.Parse()

	dsn = fmt.Sprintf("%s:%s@%s(%s:%s)/%s", *dbUser, *dbPass, *dbProto, *dbHost, *dbPort, *dbName)
	configTimer(&intervalDur)

}

// The main function. All Go programs start here.
func main() {

	// This is just to keep things open. Could be used later on to output
	// stored data, reload configuration, etc.
	http.HandleFunc("/", addDefaultHeaders(indexHandler))
	http.HandleFunc("/status", addDefaultHeaders(statusHandler))

	log.Info(fmt.Sprintf("About to listen on %s", listenStr))
	log.Info(fmt.Sprintf("LoggerID: %v", loggerID))

	err := http.ListenAndServe(listenStr, nil)
	fatalErr(err, "Server failed to start")

}

// Adds a set of default headers to each HTTP response.
func addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Server", fmt.Sprintf("%s/%v", serverName, version))
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Pragma", "no-cache")
		fn(w, r)
	}
}

// This is a simple handler which prints the current time when
// the uri is requested.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", time.Now())
}

// This prints the current status to the browser as json.
func statusHandler(w http.ResponseWriter, r *http.Request) {

	b, err := json.Marshal(status)
	if err != nil {
		fmt.Fprintf(w, `{"error": %s}`, err.Error())
		return
	}

	fmt.Fprintf(w, string(b))

}
