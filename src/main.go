package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"golap-benchmark/src/db"

)

// Database connections -------------------------------------------------
func main() {
	db.InitDuckDBConnection()
	db.InitPostgresConnection()
}