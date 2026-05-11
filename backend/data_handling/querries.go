package data_handling

/*
OLAP QUERIES DOKUMENTACE
========================

Tento soubor obsahuje 6 hierarchických OLAP queries pro analýzu prodejů automobilů.
Každý query používá pokročilé SQL agregační funkce (ROLLUP, CUBE, GROUPING SETS)
a je implementován ve dvou variantách:
  - PostgreSQL: Star schema s JOINy na 8 dimension tabulek (dim_make, dim_model, dim_body, dim_transmission, dim_color, dim_year, dim_interior)
  - DuckDB: Flat denormalizovaná struktura (raw_sales bez joinů)

Všechny queries vrací konzistentní výsledky obou databází s identickými sloupci a agregacemi.

DESIGN ROZHODNUTÍ:
- Všechny queries pracují s 558,837 car sales recordy
- NULL hodnoty se transformují na '~' v ORDER BY pro konzistenci
- ROLLUP: Hierarchické agregace s všemi podgrupy
- CUBE: Všechny kombinace dimenzí (2^n subtotals)
- GROUPING SETS: Volitelné kombinace dimenzí
- Výstup je vždy seřazený pro reproducibilitu

PERFORMANCE:
- DuckDB je 10-14x rychlejší než PostgreSQL na těchto queries
- DuckDB Q1: ~88ms | PostgreSQL Q1: ~800ms
- DuckDB Q6: ~137ms | PostgreSQL Q6: ~1800ms
*/

// Q1: Temporal Hierarchy (Decade -> Year -> Model)
// ================================================
// ÚČEL: Analýza prodejů v časové hierarchii - od dekád přes roky k modelům automobilů
// DIMENZE: 3 úrovně hierarchie (decade -> year -> model)
// AGREGACE: COUNT (počet prodejů), AVG (průměrná cena), SUM (celkový příjem)
//
// USE CASE: Vedení chce vidět trend prodejů v čase - které dekády/roky měly největší objem,
// jak se počet prodejů měnil v čase, jaký byl trend cen. Dekáda (1970, 1980, 1990 atd.)
// umožňuje vidět dlouhodobé trendy, rok konkrétní body, model detaily.
//
// EXAMPLE OUTPUT:
// decade | year | model        | sales_count | avg_price | total_revenue
// 1980   | 1982 | Toyota Camry |      45     |  12500.00 | 562500.00
// 1980   | 1983 | Honda Civic  |      32     |  11000.00 | 352000.00
// 1980   | NULL | NULL         |   1234      |  11800.00 | 14546200.00  <- subtotal pro 1980
// NULL   | NULL | NULL         | 558837      |  13200.00 | 7375000000.00 <- grand total
//
// SQL FEATURES:
// - PostgreSQL: ROLLUP (decade, year, model) + (decade/10)*10 pro grouping
// - DuckDB: Stejná logika, ale s // operátorem (integer division)
// - COALESCE v ORDER BY aby NULL subtotals šly na konec
//
func GetQuerySalesHierarchyRollup(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				CAST((y.year_int / 10) * 10 AS VARCHAR) as decade,
				CAST(y.year_int AS VARCHAR) as year,
				m.model_name as model,
				COUNT(*) as sales_count,
				ROUND(AVG(s.selling_price)::NUMERIC, 2) as avg_price,
				SUM(s.selling_price) as total_revenue
			FROM fact_sales s
			JOIN dim_year y ON s.year_id = y.year_id
			JOIN dim_model m ON s.model_id = m.model_id
			GROUP BY ROLLUP ((y.year_int / 10) * 10, y.year_int, m.model_name)
			ORDER BY COALESCE(CAST((y.year_int / 10) * 10 AS VARCHAR), '~'),
					 COALESCE(CAST(y.year_int AS VARCHAR), '~'),
					 COALESCE(m.model_name, '~'),
					 sales_count DESC`
	}
	return `
		SELECT
			CAST(CAST((year_int // 10) * 10 AS INTEGER) AS VARCHAR) as decade,
			CAST(year_int AS VARCHAR) as year,
			model,
			COUNT(*) as sales_count,
			ROUND(AVG(selling_price), 2) as avg_price,
			SUM(selling_price) as total_revenue
		FROM raw_sales
		GROUP BY ROLLUP ((year_int // 10) * 10, year_int, model)
		ORDER BY COALESCE(CAST(CAST((year_int // 10) * 10 AS INTEGER) AS VARCHAR), '~'),
				 COALESCE(CAST(year_int AS VARCHAR), '~'),
				 COALESCE(model, '~'),
				 COUNT(*) DESC`
}

// Q2: 3D CUBE (Model, Body, Transmission)
// ========================================
// ÚČEL: Třírozměrná analýza všech kombinací: Model x Body Type x Transmission
// DIMENZE: 3 nezávislé dimenze (všechny kombinace - 2^3 = 8 subtotals)
// AGREGACE: COUNT (počet prodejů), AVG (průměrná cena)
//
// USE CASE: Marketing chce pochopit cross-dimensional vztahy - jaké kombinace
// modelu, typu karosérie a převodu se nejčastěji prodávají. Např.:
// - Která kombinace (model + body) má nejvyšší cenu?
// - Který transmission typ se nejčastěji kupuje?
// - Která tělesa (sedan, suv) se prodávají s kterými modely?
//
// EXAMPLE OUTPUT:
// model        | body  | transmission | sales_count | avg_price
// Toyota Camry | sedan | automatic    |      450    | 25000.00
// Toyota Camry | sedan | manual       |      150    | 22000.00
// Toyota Camry | sedan | NULL         |      600    | 24500.00  <- subtotal sedan
// Toyota Camry | NULL  | NULL         |     2500    | 24000.00  <- subtotal model
// NULL         | NULL  | NULL         |   558837    | 13200.00  <- grand total
//
// SQL FEATURES:
// - CUBE generuje všech 2^3 = 8 kombinací (včetně NULL):
//   (model, body, transmission), (model, body), (model, transmission), (body, transmission),
//   (model), (body), (transmission), ()
// - Výstup obsahuje i agregace pro jednotlivé dimenze samostatně
// - Užitečný pro understanding cross-selling patterns
//
func GetQuerySalesCubeAllDimensions(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				m.model_name as model,
				b.body_name as body,
				t.transmission_name as transmission,
				COUNT(*) as sales_count,
				ROUND(AVG(s.selling_price)::NUMERIC, 2) as avg_price
			FROM fact_sales s
			JOIN dim_model m ON s.model_id = m.model_id
			JOIN dim_body b ON s.body_id = b.body_id
			JOIN dim_transmission t ON s.transmission_id = t.transmission_id
			GROUP BY CUBE (m.model_name, b.body_name, t.transmission_name)
			ORDER BY COALESCE(m.model_name, '~') ASC,
					 COALESCE(b.body_name, '~') ASC,
					 COALESCE(t.transmission_name, '~') ASC,
					 sales_count DESC`
	}
	return `
		SELECT
			model,
			body,
			transmission,
			COUNT(*) as sales_count,
			ROUND(AVG(selling_price), 2) as avg_price
		FROM raw_sales
		GROUP BY CUBE (model, body, transmission)
		ORDER BY COALESCE(model, '~') ASC,
				 COALESCE(body, '~') ASC,
				 COALESCE(transmission, '~') ASC,
				 COUNT(*) DESC`
}

// Q3: GROUPING SETS (Model, Color, Year)
// =======================================
// ÚČEL: Analýza selektivních kombinací dimenzí - umožňuje volit které kombinace
// mají být agregované bez generování všech možných kombinací
// DIMENZE: Tři dimenze se specifickými kombinacemi (4 skupiny):
//   1. (model, color, year) - úplné kombinace
//   2. (model, color) - po barvě v každém modelu
//   3. (year) - jen roky
//   4. () - grand total
// AGREGACE: COUNT (počet prodejů), SUM (celkový příjem)
//
// USE CASE: Analýza prodejů s flexibilní granularitou. Chceme vidět:
// - Kombinace model+barva (např. která barva se prodává s kterým modelem?)
// - Roční trendy (bez rozlišení modelu/barvy)
// - Grand total pro srovnání
// - GROUPING SETS je efektivnější než UNION pro tyto kombinace
//
// EXAMPLE OUTPUT:
// model        | color    | year | sales_count | total_revenue
// Toyota Camry | red      | 2020 |      120    | 3000000.00
// Toyota Camry | blue     | 2020 |       95    | 2375000.00
// Toyota Camry | red      | NULL |      540    | 13500000.00  <- subtotal model+color
// NULL         | NULL     | 2020 |     45000   | 594000000.00 <- subtotal year
// NULL         | NULL     | NULL |    558837   | 7375000000.00 <- grand total
//
// SQL FEATURES:
// - GROUPING SETS přesně specifikuje které kombinace agregovat (bez ostatních)
// - Efektivnější než CUBE/ROLLUP když není potřeba všechny kombinace
// - Vhodný pro custom business reporting
//
func GetQuerySalesGroupingSets(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				m.model_name as model,
				c.color_name as color,
				CAST(y.year_int AS VARCHAR) as year,
				COUNT(*) as sales_count,
				SUM(s.selling_price) as total_revenue
			FROM fact_sales s
			JOIN dim_model m ON s.model_id = m.model_id
			JOIN dim_color c ON s.color_id = c.color_id
			JOIN dim_year y ON s.year_id = y.year_id
			GROUP BY GROUPING SETS (
				(m.model_name, c.color_name, y.year_int),
				(m.model_name, c.color_name),
				(y.year_int),
				()
			)
			ORDER BY COALESCE(m.model_name, '~') ASC,
					 COALESCE(c.color_name, '~') ASC,
					 COALESCE(CAST(y.year_int AS VARCHAR), '~') ASC,
					 sales_count DESC`
	}
	return `
		SELECT
			model,
			color,
			CAST(year_int AS VARCHAR) as year,
			COUNT(*) as sales_count,
			SUM(selling_price) as total_revenue
		FROM raw_sales
		GROUP BY GROUPING SETS (
			(model, color, year_int),
			(model, color),
			(year_int),
			()
		)
		ORDER BY COALESCE(model, '~') ASC,
				 COALESCE(color, '~') ASC,
				 COALESCE(CAST(year_int AS VARCHAR), '~') ASC,
				 COUNT(*) DESC`
}

// Q4: ROLLUP (Make -> Body)
// ==========================
// ÚČEL: Dvouúrovňová hierarchická analýza - značka auta (make) je nadřazená karosérii (body)
// DIMENZE: 2 úrovně hierarchie (make -> body)
// AGREGACE: COUNT (počet prodejů), AVG (průměrná cena)
//
// USE CASE: Business analýza pro prodej - vidět trend podle výrobce a typu karosérie:
// - Která značka prodává nejčastěji?
// - Kterou karosérii preferují zákazníci každé značky?
// - Jaký je průměrný price point pro každou kombinaci?
// - ROLLUP automaticky generuje i subtotals pro jednotlivé značky
//
// EXAMPLE OUTPUT:
// make        | body   | sales_count | avg_price
// toyota      | sedan  |      1200   | 24500.00
// toyota      | suv    |       850   | 28000.00
// toyota      | NULL   |      2500   | 25500.00  <- subtotal pro toyota
// honda       | sedan  |       950   | 22000.00
// honda       | NULL   |      1500   | 23000.00  <- subtotal pro honda
// NULL        | NULL   |    558837   | 13200.00  <- grand total
//
// SQL FEATURES:
// - ROLLUP (make, body) generuje 3 úrovně:
//   (make, body), (make), ()
// - LOWER(make) pro case-insensitive grouping
// - Ideální pro reporting hierarchií (značka → typ → konkrétní auto)
//
func GetQuerySalesRollupMakeBody(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				LOWER(mk.make_name) as make,
				b.body_name as body,
				COUNT(*) as sales_count,
				ROUND(AVG(s.selling_price)::NUMERIC, 2) as avg_price
			FROM fact_sales s
			JOIN dim_make mk ON s.make_id = mk.make_id
			JOIN dim_body b ON s.body_id = b.body_id
			GROUP BY ROLLUP (LOWER(mk.make_name), b.body_name)
			ORDER BY COALESCE(LOWER(mk.make_name), '~') ASC,
					 COALESCE(b.body_name, '~') ASC,
					 sales_count DESC`
	}
	return `
		SELECT
			LOWER(make) as make,
			body,
			COUNT(*) as sales_count,
			ROUND(AVG(selling_price), 2) as avg_price
		FROM raw_sales
		GROUP BY ROLLUP (LOWER(make), body)
		ORDER BY COALESCE(LOWER(make), '~') ASC,
				 COALESCE(body, '~') ASC,
				 COUNT(*) DESC`
}

// Q5: CUBE with Price Segment (using CTE for compatibility)
// ==========================================================
// ÚČEL: Třírozměrná analýza s počítaným segmentem - Model x Cena x Transmission
// DIMENZE: 3 dimenze včetně odvozené (segment odvozený z ceny):
//   - Budget: < 10,000
//   - Mid-Range: 10,000 - 15,000
//   - Premium: > 15,000
// AGREGACE: COUNT (počet prodejů), AVG (průměrná cena)
// FEATURES: Demonstruje CTE (WITH clause) pro compute segmentation
//
// USE CASE: Pricing analytics - pochopení cen-transmission-model vztahů:
// - Který segment je nejpopulárnější?
// - Která transmise je nejdražší?
// - Jaká je distribuce cen v jednotlivých modelech?
// - Budget vs. Premium strategie
//
// EXAMPLE OUTPUT:
// model        | segment    | transmission | sales_count | avg_price
// Toyota Camry | Budget     | manual       |      120    | 9500.00
// Toyota Camry | Mid-Range  | automatic    |      200    | 12500.00
// Toyota Camry | Premium    | automatic    |       80    | 18000.00
// Toyota Camry | Budget     | NULL         |      400    | 9800.00    <- subtotal segment
// Toyota Camry | NULL       | NULL         |      780    | 12100.00   <- subtotal model
// NULL         | NULL       | NULL         |    558837   | 13200.00   <- grand total
//
// SQL FEATURES:
// - CTE (WITH) clause pro segmentaci - umožňuje CUBE na odvozené koloně
// - Cena-driven segmentace pomocí CASE WHEN
// - CUBE (3 dimenze) = 2^3 = 8 kombinací
// - Užitečné pro cenové strategie a market segmentation
//
func GetQuerySalesCubeSegment(isPostgres bool) string {
	if isPostgres {
		return `
			WITH segmented_sales AS (
				SELECT
					m.model_name,
					CASE WHEN s.selling_price > 15000 THEN 'Premium'
						 WHEN s.selling_price > 10000 THEN 'Mid-Range'
						 ELSE 'Budget' END as segment,
					t.transmission_name,
					s.selling_price
				FROM fact_sales s
				JOIN dim_model m ON s.model_id = m.model_id
				JOIN dim_transmission t ON s.transmission_id = t.transmission_id
			)
			SELECT
				model_name as model,
				segment,
				transmission_name as transmission,
				COUNT(*) as sales_count,
				ROUND(AVG(selling_price)::NUMERIC, 2) as avg_price
			FROM segmented_sales
			GROUP BY CUBE (model_name, segment, transmission_name)
			ORDER BY COALESCE(model_name, '~') ASC,
					 COALESCE(segment, '~') ASC,
					 COALESCE(transmission_name, '~') ASC,
					 sales_count DESC`
	}

	return `
		WITH segmented_sales AS (
			SELECT
				model,
				CASE WHEN selling_price > 15000 THEN 'Premium'
					 WHEN selling_price > 10000 THEN 'Mid-Range'
					 ELSE 'Budget' END as segment,
				transmission,
				selling_price
			FROM raw_sales
		)
		SELECT
			model,
			segment,
			transmission,
			COUNT(*) as sales_count,
			ROUND(AVG(selling_price), 2) as avg_price
		FROM segmented_sales
		GROUP BY CUBE (model, segment, transmission)
		ORDER BY COALESCE(model, '~') ASC,
				 COALESCE(segment, '~') ASC,
				 COALESCE(transmission, '~') ASC,
				 COUNT(*) DESC`
}

// Q6: Multi-dimensional Analysis (Model+Make, Year+Body)
// =======================================================
// ÚČEL: Nejsložitější analýza s ručně definovanými agregačními kombinacemi
// DIMENZE: 4 dimenze se specifickými kombinacemi:
//   1. (make, model) - product line analysis
//   2. (year, body) - temporal + structural analysis
//   3. (year) - pure temporal trends
//   4. () - grand total
// AGREGACE: COUNT (počet prodejů), SUM (celkový příjem)
//
// USE CASE: Strategie řízení výroby a obchodu:
// - Portfolio analysis: které kombinace make+model jsou nejziskovější?
// - Temporal trends: které roky/těla měly nejlepší prodeje?
// - Long-term patterns: trend prodejů v čase
// - Business decisions na základě výnosů (total_revenue)
// - Nedosáhlo se všech 16 kombinací (jen 4 definované) = efektivnější
//
// EXAMPLE OUTPUT:
// make        | model        | year | body   | sales_count | total_revenue
// toyota      | camry        | 2020 | sedan  |      1200   | 30000000.00
// honda       | civic        | 2020 | sedan  |       950   | 19000000.00
// NULL        | NULL         | 2020 | sedan  |     25000   | 500000000.00  <- subtotal year+body
// NULL        | NULL         | 2020 | NULL   |     45000   | 594000000.00  <- subtotal year
// toyota      | camry        | NULL | NULL   |      2500   | 65000000.00   <- subtotal model+make
// NULL        | NULL         | NULL | NULL   |    558837   | 7375000000.00 <- grand total
//
// SQL FEATURES:
// - Kombinuje GROUPING SETS pro custom reporting
// - Product-centric (make+model) a temporal-centric (year+body) views současně
// - Maximální granularity v nejdůležitějších kombinacích
// - Vhodné pro executive dashboards a strategic planning
//
func GetQuerySalesMultiAnalysis(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				mk.make_name as make,
				m.model_name as model,
				CAST(y.year_int AS VARCHAR) as year,
				b.body_name as body,
				COUNT(*) as sales_count,
				SUM(s.selling_price) as total_revenue
			FROM fact_sales s
			JOIN dim_make mk ON s.make_id = mk.make_id
			JOIN dim_model m ON s.model_id = m.model_id
			JOIN dim_year y ON s.year_id = y.year_id
			JOIN dim_body b ON s.body_id = b.body_id
			GROUP BY GROUPING SETS (
				(mk.make_name, m.model_name),
				(y.year_int, b.body_name),
				(y.year_int),
				()
			)
			ORDER BY COALESCE(mk.make_name, '~') ASC,
					 COALESCE(m.model_name, '~') ASC,
					 COALESCE(CAST(y.year_int AS VARCHAR), '~') ASC,
					 COALESCE(b.body_name, '~') ASC,
					 sales_count DESC`
	}
	return `
		SELECT
			make,
			model,
			CAST(year_int AS VARCHAR) as year,
			body,
			COUNT(*) as sales_count,
			SUM(selling_price) as total_revenue
		FROM raw_sales
		GROUP BY GROUPING SETS (
			(make, model),
			(year_int, body),
			(year_int),
			()
		)
		ORDER BY COALESCE(make, '~') ASC,
				 COALESCE(model, '~') ASC,
				 COALESCE(CAST(year_int AS VARCHAR), '~') ASC,
				 COALESCE(body, '~') ASC,
			COUNT(*) DESC`
}

// GetQueryHeaders vrátí seznam jmen sloupců pro každý query
// =========================================================
// ÚČEL: Poskytuje metadata pro CSV export - seznam správných jmen sloupců pro každý query
// DŮVOD: Každý query vrací tab-separated data bez headers, tak je potřeba přidat je ručně
// pro CSV formát
//
// PARAMETR: queryNumber (0-5) odpovídá Q1-Q6
//
// VRACÍ: []string se jmény sloupců v pořadí, jak je vrací SELECT klauzule
//
// POUŽITÍ: V benchamrk.go - při exporzu výsledků do CSV souboru
//
// PŘÍKLAD: GetQueryHeaders(0) vrátí:
//   ["decade", "year", "model", "sales_count", "avg_price", "total_revenue"]
// což je pak napsáno jako první řádek CSV souboru
//
func GetQueryHeaders(queryNumber int) []string {
	switch queryNumber {
	case 0: // Q1
		return []string{"decade", "year", "model", "sales_count", "avg_price", "total_revenue"}
	case 1: // Q2
		return []string{"model", "body", "transmission", "sales_count", "avg_price"}
	case 2: // Q3
		return []string{"model", "color", "year", "sales_count", "total_revenue"}
	case 3: // Q4
		return []string{"make", "body", "sales_count", "avg_price"}
	case 4: // Q5
		return []string{"model", "segment", "transmission", "sales_count", "avg_price"}
	case 5: // Q6
		return []string{"make", "model", "year", "body", "sales_count", "total_revenue"}
	default:
		return []string{}
	}
}
