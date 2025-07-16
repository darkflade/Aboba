package main

// Configure server ip and TLS
const (
	listenIP = "192.168.1.3:443"
	certPath = "certs/server.crt"
	keyPath  = "certs/server.key"

	unidocKey = "6f5addfcb0e2f0221366807632afb5e224a938a9968b0d8a35baa541eab2fc33"

	host     = "127.127.126.1"
	dbName   = "WorkDB"
	user     = "root"
	password = ""
	port     = "3306"
	dsn      = user + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbName + "?parseTime=true&multiStatements=true"
)

// Global variable----------------------------------------------
var storageEmployeeImages = "./static/images"
