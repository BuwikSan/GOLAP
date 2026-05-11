package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	. "golap-benchmark/backend"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Obtain Conn
func ObtainDuckDBConnection() SQLOverhead {
	/* Creates connection to DuckDB database and returns it. */
	// db := openDBConnection("duckdb")
	cmd := exec.Command("backend/db/db_files/duckdb.exe", "backend/db/db_files/DuckDB.db", "-c", "SELECT 1;")

	// Zachytíme výstup pro případ chyby
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("DuckDB inicializace selhala: %v\nVýstup: %s", err, string(output))
	}

	db := &DuckDBOverhead{
		Path:    "backend/db/db_files/DuckDB.db",
		BinPath: "backend/db/db_files/duckdb.exe",
	}

	fmt.Println("DuckDB connected!")
	return db
}

func ObtainPostgresConnection() SQLOverhead {
	/* Creates connection to PostgreSQL database and returns it. */
	// db := openDBConnection("pgx")
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost" // defaultní hodnota
	}
	pgxConnStr := fmt.Sprintf("host=%s port=5432 user=user password=password dbname=benchmark_db sslmode=disable", dbHost)

	pg, err := sql.Open("pgx", pgxConnStr)
	if err != nil {
		log.Fatal(err)
	}

	// Ping - ověř, že funguje
	err = pg.Ping()
	if err != nil {
		log.Fatal(err)
	}

	db := &PostgresOverhead{pg}
	fmt.Println("PostgreSQL connected!")
	return db
}

// Database Drops -------------------------------------------------
func dropDuckDB(db SQLOverhead) {
	/*drop all content from duckDB database*/
	err := db.RunRaw("DROP TABLE IF EXISTS raw_sales")
	if err != nil {
		log.Fatal("Nelze smazat tabulku raw_sales v DuckDB: ", err)
	}
	err = db.RunRaw("DROP SEQUENCE IF EXISTS id_sequence")
	if err != nil {
		log.Fatal("Nelze smazat tabulku raw_sales v DuckDB: ", err)
	}
	DropResultsTables(db)
}

func dropPostgres(db SQLOverhead) {
	/*drop all content from PostgreSQL database*/
	dropTables := func(name string) {
		err := db.RunRaw(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", name))
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
func initSchema(db SQLOverhead, schemaName string) {
	/*initialize schema for selected database*/
	// 1. Načti schema.sql soubor
	path := fmt.Sprintf("./backend/db/db_schemas/%s_schema.sql", schemaName)
	schemaFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Nelze načíst schema:", err)
	}

	// 2. Převeď na string
	schema := string(schemaFile)
	// 3. Spusť SQL příkazy
	err = db.RunRaw(schema)
	if err != nil {
		log.Fatal("Nelze spustit schema.sql: ", err)
	}
	fmt.Println(schemaName + " schema initialized!")
}

// Database initializations -------------------------------------------------

func InitDuckDBConnection() SQLOverhead {
	db := ObtainDuckDBConnection()
	dropDuckDB(db)
	fmt.Println("DuckDB data dropped!")
	initSchema(db, "duckDB")
	return db
}

func InitPostgresConnection() SQLOverhead {
	db := ObtainPostgresConnection()
	dropPostgres(db)
	fmt.Println("PostgreSQL data dropped!")
	initSchema(db, "postgres")
	return db
}
