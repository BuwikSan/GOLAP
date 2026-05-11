package db

import (
	"fmt"
	"log"
	"strings"

	. "golap-benchmark/backend"
	data_handling "golap-benchmark/backend/data_handling"
)

// SaveResultsToDuckDB saves all 6 OLAP query results from both PostgreSQL and DuckDB
// into DuckDB tables for later analysis and visualization
//
// TABLES CREATED:
// - results_q1_postgres, results_q1_duckdb
// - results_q2_postgres, results_q2_duckdb
// - results_q3_postgres, results_q3_duckdb
// - results_q4_postgres, results_q4_duckdb
// - results_q5_postgres, results_q5_duckdb
// - results_q6_postgres, results_q6_duckdb
//
// Plus empty tables for data mining results:
// - anomalies
// - clusters
// - forecasts
// - correlations
func SaveResultsToDuckDB(duckdb SQLOverhead, postgresDB SQLOverhead) error {
	fmt.Println("\n💾 Saving OLAP query results to DuckDB tables...")

	// Query definitions and names
	queryNames := []string{
		"Q1_Hierarchy_Rollup_Decade-Year-Model",
		"Q2_Cube_Model-Body-Transmission",
		"Q3_Grouping_Sets_Model-Color-Year",
		"Q4_Rollup_Make-Body",
		"Q5_Cube_Segment",
		"Q6_Multi_Analysis",
	}

	postgresQueryList := []string{
		data_handling.GetQuerySalesHierarchyRollup(true),
		data_handling.GetQuerySalesCubeAllDimensions(true),
		data_handling.GetQuerySalesGroupingSets(true),
		data_handling.GetQuerySalesRollupMakeBody(true),
		data_handling.GetQuerySalesCubeSegment(true),
		data_handling.GetQuerySalesMultiAnalysis(true),
	}

	duckdbQueryList := []string{
		data_handling.GetQuerySalesHierarchyRollup(false),
		data_handling.GetQuerySalesCubeAllDimensions(false),
		data_handling.GetQuerySalesGroupingSets(false),
		data_handling.GetQuerySalesRollupMakeBody(false),
		data_handling.GetQuerySalesCubeSegment(false),
		data_handling.GetQuerySalesMultiAnalysis(false),
	}

	// Save PostgreSQL results
	for i, query := range postgresQueryList {
		tableName := fmt.Sprintf("results_%s_postgres", strings.ToLower(strings.Split(queryNames[i], "_")[0]))

		// Create table by wrapping the query in CREATE TABLE AS
		createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s AS %s", tableName, query)

		_, err := postgresDB.Exec(createTableSQL)
		if err != nil {
			log.Printf("❌ Failed to save %s from PostgreSQL: %v", tableName, err)
			return err
		}
		fmt.Printf("✅ Saved %s (PostgreSQL)\n", tableName)
	}

	// Save DuckDB results
	for i, query := range duckdbQueryList {
		tableName := fmt.Sprintf("results_%s_duckdb", strings.ToLower(strings.Split(queryNames[i], "_")[0]))

		// Create table by wrapping the query in CREATE TABLE AS
		createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s AS %s", tableName, query)

		err := duckdb.RunRaw(createTableSQL)
		if err != nil {
			log.Printf("❌ Failed to save %s to DuckDB: %v", tableName, err)
			return err
		}
		fmt.Printf("✅ Saved %s (DuckDB)\n", tableName)
	}

	// Create data mining result tables (empty, to be populated by algorithms)
	dataMiningTables := []string{
		`CREATE TABLE IF NOT EXISTS anomalies (
			model VARCHAR,
			year INTEGER,
			metric VARCHAR,
			value NUMERIC,
			anomaly_score NUMERIC,
			reason VARCHAR,
			created_at TIMESTAMP DEFAULT current_timestamp
		)`,
		`CREATE TABLE IF NOT EXISTS clusters (
			cluster_id INTEGER,
			model VARCHAR,
			avg_price NUMERIC,
			sales_count INTEGER,
			cluster_center_price NUMERIC,
			created_at TIMESTAMP DEFAULT current_timestamp
		)`,
		`CREATE TABLE IF NOT EXISTS forecasts (
			model VARCHAR,
			year INTEGER,
			predicted_sales INTEGER,
			historical_sales INTEGER,
			confidence NUMERIC,
			method VARCHAR,
			created_at TIMESTAMP DEFAULT current_timestamp
		)`,
		`CREATE TABLE IF NOT EXISTS correlations (
			dimension_1 VARCHAR,
			dimension_2 VARCHAR,
			correlation NUMERIC,
			p_value NUMERIC,
			created_at TIMESTAMP DEFAULT current_timestamp
		)`,
	}

	for _, tableSQL := range dataMiningTables {
		err := duckdb.RunRaw(tableSQL)
		if err != nil {
			log.Printf("❌ Failed to create data mining table: %v", err)
			return err
		}
	}

	fmt.Println("✅ All query results and data mining tables created successfully!")
	fmt.Println("📊 Ready for Streamlit dashboard consumption")
	return nil
}

// DropResultsTables deletes all result tables from DuckDB for a clean restart
//
// TABLES DROPPED:
// - All results_* tables (Q1-Q6, both PostgreSQL and DuckDB)
// - All data mining tables (anomalies, clusters, forecasts, correlations)
//
// This is useful for:
// - Clearing old results before rerunning analysis
// - Testing/debugging
// - Resetting state between iterations
func DropResultsTables(duckdb SQLOverhead) error {
	fmt.Println("\n🗑️ Dropping all result and data mining tables from DuckDB...")

	tablesToDrop := []string{
		"results_q1_postgres",
		"results_q1_duckdb",
		"results_q2_postgres",
		"results_q2_duckdb",
		"results_q3_postgres",
		"results_q3_duckdb",
		"results_q4_postgres",
		"results_q4_duckdb",
		"results_q5_postgres",
		"results_q5_duckdb",
		"results_q6_postgres",
		"results_q6_duckdb",
		"anomalies",
		"clusters",
		"forecasts",
		"correlations",
	}

	for _, tableName := range tablesToDrop {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)

		err := duckdb.RunRaw(dropSQL)
		if err != nil {
			log.Printf("⚠️  Warning: Could not drop %s: %v", tableName, err)
			// Don't return error on drop - just warn
		} else {
			fmt.Printf("✅ Dropped %s\n", tableName)
		}
	}

	fmt.Println("✅ All result tables cleared!")
	return nil
}

// VerifyResultsTables lists all created result tables from DuckDB
// Useful for debugging and verification
func VerifyResultsTables(duckdb SQLOverhead) ([]string, error) {
	// Query to list all tables
	listTablesSQL := `
		SELECT name FROM duckdb_tables()
		WHERE name LIKE 'results_%' OR name IN ('anomalies', 'clusters', 'forecasts', 'correlations')
		ORDER BY name
	`

	output, _, err := duckdb.QueryRow(listTablesSQL)
	if err != nil {
		return nil, err
	}

	// Parse output into slice
	tables := strings.Split(strings.TrimSpace(output), "\n")

	fmt.Printf("\n📋 Verified result tables in DuckDB:\n")
	for _, table := range tables {
		if table != "" {
			fmt.Printf("  ✅ %s\n", table)
		}
	}

	return tables, nil
}
