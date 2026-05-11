package main

import (
	"fmt"
	. "golap-benchmark/data"
	. "golap-benchmark/src"
	. "golap-benchmark/src/data_handling"
	. "golap-benchmark/src/db"
	. "golap-benchmark/src/testing"
)

// Database connections -------------------------------------------------
func main() {
	// config variables
	var run_inicialization bool = false
	var testing bool = true

	// database connections
	var duckDB SQLOverhead
	var postgreSQL SQLOverhead

	if run_inicialization {
		PreprocessCSV()
		fmt.Println("============== Inicializace ==============")
		duckDB = InitDuckDBConnection()
		postgreSQL = InitPostgresConnection()
		fmt.Println("============== plnění ==============")
		FillDuckDBWithCars(duckDB)
		FillPostgresWithCars(postgreSQL)
	} else {
		fmt.Println("============== connection ==============")
		duckDB = ObtainDuckDBConnection()
		postgreSQL = ObtainPostgresConnection()
	}

	if testing {
		CompareDataBetweenDatabases(postgreSQL, duckDB)
		duckDBQuerryManager := &OLAPCustomQueryManager{
			DB: duckDB,
		}
		postgreSQLQuerryManager := &OLAPCustomQueryManager{
			DB: postgreSQL,
		}

		fmt.Println("============== Querries z builderu ==============")
		duckDBQuerryManager.RunAdvancedOLAP(Rollup, []string{"year_int", "make", "model"})
		postgreSQLQuerryManager.RunAdvancedOLAP(Rollup, []string{"year_int", "make", "model"})

	} else {
		fmt.Println("============== row check ==============")
		DuckGetRawSales(duckDB)
		PostgresGetFactSales(postgreSQL)
	}

	fmt.Println("============== benchmark ==============")

	Benchmarker := BenchmarkRunner{
		Postgres: postgreSQL,
		DuckDB:   duckDB,
	}
	Benchmarker.RunBenchmark()

	if testing {
		fmt.Printf("=========== porovnavam csv ==============\n")
		comparisons, _ := CompareSavedQueries("./data")
		PrintAllQueriesComparison(comparisons)
	}
}

// dodělej comparison csv, a pak udělej vizualizaci a máš to
