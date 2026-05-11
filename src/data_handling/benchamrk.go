package data_handling

import (
	"fmt"
	"log"
	"os"

	. "golap-benchmark/data"
	. "golap-benchmark/src"
)

type BenchmarkRunner struct {
	Postgres SQLOverhead
	DuckDB   SQLOverhead
}

func (b *BenchmarkRunner) runQuery(sql string, db SQLOverhead) (result string, duration string, err error) {
	output, duration, err := db.QueryRow(sql)
	if err != nil {
		log.Fatal(err)
	}

	return output, duration, err
}

func (b *BenchmarkRunner) runQueriesFromList(queryList []string, db SQLOverhead, queryNames []string) {
	saveToFile := false
	var count int = 0

	// Create output directory if it doesn't exist
	outputDir := "./data/saved_queries"
	os.MkdirAll(outputDir, 0755)

	for i, query := range queryList {
		output, duration, _ := b.runQuery(query, db)

		// Print nicely formatted benchmark result
		fmt.Printf("  [%d] %s: %s\n", i+1, queryNames[count], duration)

		filename := fmt.Sprintf("%s/%s_%s.csv", outputDir, db.String(), queryNames[count])
		if saveToFile {
			_ = TableAsStringToCSV(output, db, filename)
		}
		count++
	}
}

func (b *BenchmarkRunner) RunBenchmark() {
	queryNames := []string{
		"Q1_Hierarchy_Rollup_Decade-Year-Model",
		"Q2_Cube_Model-Body-Transmission",
		"Q3_Grouping_Sets_Model-Color-Year",
		"Q4_Rollup_Make-Body",
		"Q5_Cube_Segment",
		"Q6_Multi_Analysis",
	}

	postgresQueryList := []string{
		GetQuerySalesHierarchyRollup(true),
		GetQuerySalesCubeAllDimensions(true),
		GetQuerySalesGroupingSets(true),
		GetQuerySalesRollupMakeBody(true),
		GetQuerySalesCubeSegment(true),
		GetQuerySalesMultiAnalysis(true),
	}

	duckdbQueryList := []string{
		GetQuerySalesHierarchyRollup(false),
		GetQuerySalesCubeAllDimensions(false),
		GetQuerySalesGroupingSets(false),
		GetQuerySalesRollupMakeBody(false),
		GetQuerySalesCubeSegment(false),
		GetQuerySalesMultiAnalysis(false),
	}

	fmt.Println("\n📊 PostgreSQL Benchmark Results:")
	b.runQueriesFromList(postgresQueryList, b.Postgres, queryNames)

	fmt.Println("\n📊 DuckDB Benchmark Results:")
	b.runQueriesFromList(duckdbQueryList, b.DuckDB, queryNames)
}
