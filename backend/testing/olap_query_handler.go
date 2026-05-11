package testing

import (
	"fmt"
	"log"
	"strings"

	. "golap-benchmark/backend"
)

// Mapa: Název sloupce v aplikaci -> [Tabulka v Postgresu, Sloupec s ID, Sloupec s názvem]
var dimensionsMap = map[string][3]string{
	"year_int":     {"dim_year", "year_id", "year_int"},
	"make":         {"dim_make", "make_id", "make_name"},
	"model":        {"dim_model", "model_id", "model_name"},
	"trim":         {"dim_trim", "trim_id", "trim_name"},
	"body":         {"dim_body", "body_id", "body_name"},
	"transmission": {"dim_transmission", "transmission_id", "transmission_name"},
	"color":        {"dim_color", "color_id", "color_name"},
	"interior":     {"dim_interior", "interior_id", "interior_name"},
}

type OLAPMode string

const (
	Rollup       OLAPMode = "ROLLUP"
	Cube         OLAPMode = "CUBE"
	GroupingSets OLAPMode = "GROUPING SETS"
)

// olap query manager
type OLAPCustomQueryManager struct {
	DB SQLOverhead
}

func (m *OLAPCustomQueryManager) GetDB() SQLOverhead {
	if m.DB == nil {
		log.Fatal("No database set")
	}
	return m.DB
}

func (m OLAPCustomQueryManager) SetDB(db SQLOverhead) {
	m.DB = db
}

func (m *OLAPCustomQueryManager) RunAdvancedOLAP(mode OLAPMode, dims []string) (string, error) {
	var sql string

	var selectCols, joins, groupCols []string

	switch m.DB.(type) {
	case *PostgresOverhead:
		for _, dName := range dims {
			d := dimensionsMap[dName] // Tvoje mapa dimenzí
			// COALESCE zajistí hezké popisky u totálů (místo NULL napíše 'ALL')
			// CAST to TEXT to handle both numeric and string types
			selectCols = append(selectCols, fmt.Sprintf("COALESCE(CAST(d_%s.%s AS TEXT), 'ALL %s') as %s", dName, d[2], dName, dName))
			joins = append(joins, fmt.Sprintf("JOIN %s d_%s ON s.%s = d_%s.%s", d[0], dName, d[1], dName, d[1]))
			groupCols = append(groupCols, fmt.Sprintf("d_%s.%s", dName, d[2]))
		}

		sql = fmt.Sprintf(`
			SELECT %s, COUNT(*) as count, AVG(s.selling_price) as avg_price
			FROM fact_sales s
			%s
			GROUP BY %s(%s)
			ORDER BY count DESC`,
			strings.Join(selectCols, ", "),
			strings.Join(joins, "\n"),
			mode, // ROLLUP nebo CUBE
			strings.Join(groupCols, ", "),
		)

	default:
		var groupCols []string
		for _, d := range dims {
			groupCols = append(groupCols, d)
		}

		cols := strings.Join(dims, ", ")
		sql = fmt.Sprintf(`
			SELECT %s, COUNT(*) as count, AVG(selling_price) as avg_price
			FROM raw_sales
			GROUP BY %s(%s)
			ORDER BY count DESC`,
			cols, mode, cols)
	}

	output, duration, err := m.DB.QueryRow(sql)
	if err != nil {
		log.Fatal(err)
	}

	ResultDB := fmt.Sprintf("%T", m.DB)
	ResultQueryType := mode
	ResultTime := duration

	fmt.Println("DB:", ResultDB)
	fmt.Println("QueryType:", ResultQueryType)
	fmt.Println("Time:", ResultTime)

	return output, err
}
