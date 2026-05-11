package main

import (
	"fmt"
	. "golap-benchmark/backend"
	. "golap-benchmark/backend/data_handling"
	. "golap-benchmark/backend/db"
	. "golap-benchmark/backend/testing"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Command-line flags
	// dropResultsFlag := flag.Bool("dropResults", false, "Drop all result tables from DuckDB before running")
	// saveToDuckDBFlag := flag.Bool("saveToDuckDB", false, "Save query results to DuckDB tables (for Streamlit)")
	// flag.Parse()

	// config variables -----------------------------------------------------
	var runInicializationFlag bool = false // je na kurevsky dlouho cca 8 min kvuli postgres fillu
	// db testing (done)
	var compareDBContentsFlag bool = false

	//testingFlag queries // querries už mají totožné výstupy není nutno
	var testingFlag bool = false
	var customQueriesFlag bool = false

	// main line
	var includeDatabasesAndBenchmarksFalg bool = true // Pokud je false, přeskočíme databáze a jdeme jen na CSV porovnání
	var saveToDuckAndCSVFlag bool = true

	// database connections
	var duckDB SQLOverhead
	var postgreSQL SQLOverhead

	if runInicializationFlag {
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

		// Drop result tables if requested

	}

	if compareDBContentsFlag {
		fmt.Println("============== db content comparison ==============")
		CompareDataBetweenDatabases(postgreSQL, duckDB)
	} else {
		fmt.Println("skipping content comparison.....")
		fmt.Println("============== row check ==============")
		DuckGetRawSales(duckDB)
		PostgresGetFactSales(postgreSQL)
	}

	if includeDatabasesAndBenchmarksFalg {
		fmt.Println("============== benchmark ==============")

		Benchmarker := BenchmarkRunner{
			Postgres: postgreSQL,
			DuckDB:   duckDB,
		}
		Benchmarker.RunBenchmark(saveToDuckAndCSVFlag)

		// Save results to DuckDB tables if requested
		if saveToDuckAndCSVFlag {
			fmt.Println("\n📊 User requested: Saving results to DuckDB tables...")
			fmt.Println("🗑️  User requested: Dropping old result tables...")
			DropResultsTables(duckDB)
			err := SaveResultsToDuckDB(duckDB, postgreSQL)
			if err != nil {
				fmt.Printf("❌ Error saving results: %v\n", err)
			} else {
				// Verify that tables were created
				VerifyResultsTables(duckDB)
				fmt.Println("\n✅ Ready for Streamlit dashboard! Start with: streamlit run dashboard.py")
			}
		}
	} else {
		fmt.Println("⏭️  Přeskakuji databáze, jdu přímo na CSV porovnání")
	}

	if testingFlag {
		if customQueriesFlag {
			duckDBQuerryManager := &OLAPCustomQueryManager{
				DB: duckDB,
			}
			postgreSQLQuerryManager := &OLAPCustomQueryManager{
				DB: postgreSQL,
			}
			fmt.Println("============== Querries z builderu ==============")
			duckDBQuerryManager.RunAdvancedOLAP(Rollup, []string{"year_int", "make", "model"})
			postgreSQLQuerryManager.RunAdvancedOLAP(Rollup, []string{"year_int", "make", "model"})
		}

		fmt.Printf("=========== ANALYTIKA SLOUPCŮ - Porovnání hodnot ==============\n")
		// Získej absolutní cestu k data složce
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Chyba při získání pracovního adresáře: %v\n", err)
		} else {
			dataPath := filepath.Join(wd, "data")
			savedQueriesPath := filepath.Join(dataPath, "saved_queries")
			fmt.Printf("🔍 Analytika všech 6 queries v: %s\n\n", savedQueriesPath)

			// Analyzuj všechny 6 queries
			queries := []struct {
				name string
				pg   string
				duck string
			}{
				{"Q1 - Hierarchy Rollup (Decade-Year-Model)", "postgres_Q1_Hierarchy_Rollup_Decade-Year-Model.csv", "duckdb_Q1_Hierarchy_Rollup_Decade-Year-Model.csv"},
				{"Q2 - Cube (Model-Body-Transmission)", "postgres_Q2_Cube_Model-Body-Transmission.csv", "duckdb_Q2_Cube_Model-Body-Transmission.csv"},
				{"Q3 - Grouping Sets (Model-Color-Year)", "postgres_Q3_Grouping_Sets_Model-Color-Year.csv", "duckdb_Q3_Grouping_Sets_Model-Color-Year.csv"},
				{"Q4 - Rollup (Make-Body)", "postgres_Q4_Rollup_Make-Body.csv", "duckdb_Q4_Rollup_Make-Body.csv"},
				{"Q5 - Cube Segment", "postgres_Q5_Cube_Segment.csv", "duckdb_Q5_Cube_Segment.csv"},
				{"Q6 - Multi Analysis", "postgres_Q6_Multi_Analysis.csv", "duckdb_Q6_Multi_Analysis.csv"},
			}

			for _, q := range queries {
				fmt.Printf("\n%s\n", strings.Repeat("=", 120))
				q_pg := filepath.Join(savedQueriesPath, q.pg)
				q_duck := filepath.Join(savedQueriesPath, q.duck)
				PrintDetailedQueryComparison(q_pg, q_duck, q.name)
			}
		}
	}
}

// dodělej comparison csv, a pak udělej vizualizaci a máš to
