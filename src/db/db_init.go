package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/marcboeker/go-duckdb"
)

// Database connections -------------------------------------------------
func openDBConnection(DBType string) *sql.DB {
	/*initialize selected database and return connection*/
	// Otevři DuckDB spojení
	pgxConnStr := "host=172.25.254.161 port=5432 user=user dbname=benchmark_db sslmode=disable"
	duckDBPath := "src/data_handling/db_files/DuckDB.db"

	var DBInterface string
	switch DBType {
	case "duckdb":
		DBInterface = fmt.Sprintf("file:%s", duckDBPath)

		_, err := os.Stat(duckDBPath)
		if os.IsNotExist(err) {
			_, err := os.Create(duckDBPath)
			if err != nil {
				log.Fatal("Nelze vytvořit DuckDB soubor:", err)
			}
		}
	case "pgx":
		DBInterface = pgxConnStr
	}

	db, err := sql.Open(DBType, DBInterface)
	if err != nil {
		log.Fatal(err)
	}

	// Ping - ověř, že funguje
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DuckDB connected!")
	return db
}

// Database Drops -------------------------------------------------
func dropDuckDB(db *sql.DB) {
	/*drop all content from duckDB database*/
	_, err := db.Exec("DROP TABLE IF EXISTS raw_sales")
	if err != nil {
		log.Fatal("Nelze smazat tabulku raw_sales v DuckDB:", err)
	}
}

func dropPostgres(db *sql.DB) {
	/*drop all content from PostgreSQL database*/
	dropTables := func(name string) {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", name))
		if err != nil {
			log.Fatalf("Nelze smazat tabulku %s: %v", name, err)
		}
	}
	dropTables("fact_sales")
	dropTables("dim_year")
	dropTables("dim_make")
	dropTables("dim_model")
	dropTables("dim_trim")
	dropTables("dim_body")
	dropTables("dim_transmission")
	dropTables("dim_color")
	dropTables("dim_interior")
}

// Schema initialization -------------------------------------------------
func initSchema(db *sql.DB, schemaName string) {
	/*initialize schema for selected database*/
	// 1. Načti schema.sql soubor
	path := fmt.Sprintf("./src/db/db_schemas/%s_schema.sql", schemaName)
	schemaFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Nelze načíst schema:", err)
	}

	// 2. Převeď na string
	schema := string(schemaFile)
	// 3. Spusť SQL příkazy
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal("Nelze spustit schema.sql:", err)
	}

	fmt.Println(schemaName + " schema initialized!")
}

// Database initializations -------------------------------------------------

func InitDuckDBConnection() *sql.DB {
	/* Creates connection to DuckDB database and returns it. */
	db := openDBConnection("duckdb")
	fmt.Println("DuckDB connected!")
	dropDuckDB(db)
	fmt.Println("DuckDB data dropped!")
	initSchema(db, "duckDB")
	return db
}

func InitPostgresConnection() *sql.DB {
	/* Creates connection to PostgreSQL database and returns it. */
	db := openDBConnection("pgx")
	fmt.Println("PostgreSQL connected!")
	dropPostgres(db)
	fmt.Println("PostgreSQL data dropped!")
	initSchema(db, "postgres")
	return db
}