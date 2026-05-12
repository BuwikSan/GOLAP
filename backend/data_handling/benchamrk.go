package data_handling

import (
	"fmt"
	"log"
	"os"

	. "golap-benchmark/backend"
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

func (b *BenchmarkRunner) runQueriesFromList(queryList []string, db SQLOverhead, queryNames []string, headers [][]string, saveToFile bool) []string {
	var count int = 0

	// Create output directory if it doesn't exist
	outputDir := "./data/saved_queries"
	os.MkdirAll(outputDir, 0755)

	outputCollection := []string{}

	for i, query := range queryList {
		output, duration, _ := b.runQuery(query, db)

		// Print nicely formatted benchmark result

		outputCollection = append(outputCollection, fmt.Sprintf("  [%d] %s: %s", i+1, queryNames[count], duration))

		if saveToFile {
			filename := fmt.Sprintf("%s/%s_%s.csv", outputDir, db.String(), queryNames[count])
			_ = TableAsStringToCSV(output, db, filename, headers[count])
		}
		count++
	}
	return outputCollection
}

func (b *BenchmarkRunner) RunBenchmark(saveToFile bool) {
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

	// Get headers for each query
	var queryHeaders [][]string
	for i := 0; i < len(queryNames); i++ {
		queryHeaders = append(queryHeaders, GetQueryHeaders(i))
	}

	postgresTimes := b.runQueriesFromList(postgresQueryList, b.Postgres, queryNames, queryHeaders, saveToFile)
	duckTimes := b.runQueriesFromList(duckdbQueryList, b.DuckDB, queryNames, queryHeaders, saveToFile)

	for i := 0; i < len(queryNames); i++ {
		fmt.Printf("%d. Query comparison:\n", i)
		fmt.Printf("  Postgres: %s\n", postgresTimes[i])
		fmt.Printf("  DuckDB: %s\n", duckTimes[i])
	}
}
