package data_handling

// Q1: Temporal Hierarchy (Decade -> Year -> Model)
func GetQuerySalesHierarchyRollup(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				(y.year_int / 10) * 10 as decade,
				y.year_int as year,
				m.model_name as model,
				COUNT(*) as sales_count,
				ROUND(AVG(s.selling_price)::NUMERIC, 2) as avg_price,
				SUM(s.selling_price) as total_revenue
			FROM fact_sales s
			JOIN dim_year y ON s.year_id = y.year_id
			JOIN dim_model m ON s.model_id = m.model_id
			GROUP BY ROLLUP ((y.year_int / 10) * 10, y.year_int, m.model_name)
			ORDER BY decade DESC, year DESC, sales_count DESC`
	}
	return `
		SELECT
			(year_int / 10) * 10 as decade,
			year_int as year,
			model,
			COUNT(*) as sales_count,
			ROUND(AVG(selling_price), 2) as avg_price,
			SUM(selling_price) as total_revenue
		FROM raw_sales
		GROUP BY ROLLUP ((year_int / 10) * 10, year_int, model)
		ORDER BY decade DESC, year DESC, sales_count DESC`
}

// Q2: 3D CUBE (Model, Body, Transmission)
func GetQuerySalesCubeAllDimensions(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				COALESCE(m.model_name, 'ALL MODELS') as model,
				COALESCE(b.body_name, 'ALL BODIES') as body,
				COALESCE(t.transmission_name, 'ALL TRANS') as transmission,
				COUNT(*) as sales_count,
				ROUND(AVG(s.selling_price)::NUMERIC, 2) as avg_price
			FROM fact_sales s
			JOIN dim_model m ON s.model_id = m.model_id
			JOIN dim_body b ON s.body_id = b.body_id
			JOIN dim_transmission t ON s.transmission_id = t.transmission_id
			GROUP BY CUBE (m.model_name, b.body_name, t.transmission_name)
			ORDER BY sales_count DESC`
	}
	return `
		SELECT
			COALESCE(model, 'ALL MODELS') as model,
			COALESCE(body, 'ALL BODIES') as body,
			COALESCE(transmission, 'ALL TRANS') as transmission,
			COUNT(*) as sales_count,
			ROUND(AVG(selling_price), 2) as avg_price
		FROM raw_sales
		GROUP BY CUBE (model, body, transmission)
		ORDER BY sales_count DESC`
}

// Q3: GROUPING SETS (Model, Color, Year)
func GetQuerySalesGroupingSets(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				COALESCE(m.model_name, 'TOTAL') as model,
				COALESCE(c.color_name, 'ALL') as color,
				COALESCE(CAST(y.year_int AS VARCHAR), 'ALL') as year,
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
			ORDER BY sales_count DESC`
	}
	return `
		SELECT
			COALESCE(model, 'TOTAL') as model,
			COALESCE(color, 'ALL') as color,
			COALESCE(CAST(year_int AS VARCHAR), 'ALL') as year,
			COUNT(*) as sales_count,
			SUM(selling_price) as total_revenue
		FROM raw_sales
		GROUP BY GROUPING SETS (
			(model, color, year_int),
			(model, color),
			(year_int),
			()
		)
		ORDER BY sales_count DESC`
}

// Q4: ROLLUP (Make -> Body)
func GetQuerySalesRollupMakeBody(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				COALESCE(mk.make_name, 'TOTAL') as make,
				COALESCE(b.body_name, 'ALL') as body,
				COUNT(*) as sales_count,
				ROUND(AVG(s.selling_price)::NUMERIC, 2) as avg_price
			FROM fact_sales s
			JOIN dim_make mk ON s.make_id = mk.make_id
			JOIN dim_body b ON s.body_id = b.body_id
			GROUP BY ROLLUP (mk.make_name, b.body_name)
			ORDER BY make, sales_count DESC`
	}
	return `
		SELECT
			COALESCE(make, 'TOTAL') as make,
			COALESCE(body, 'ALL') as body,
			COUNT(*) as sales_count,
			ROUND(AVG(selling_price), 2) as avg_price
		FROM raw_sales
		GROUP BY ROLLUP (make, body)
		ORDER BY make, sales_count DESC`
}

// Q5: CUBE with Price Segment (using CTE for compatibility)
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
				COALESCE(model_name, 'ALL') as model,
				COALESCE(segment, 'ALL') as segment,
				COALESCE(transmission_name, 'ALL') as transmission,
				COUNT(*) as sales_count,
				ROUND(AVG(selling_price)::NUMERIC, 2) as avg_price
			FROM segmented_sales
			GROUP BY CUBE (model_name, segment, transmission_name)
			ORDER BY sales_count DESC`
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
			COALESCE(model, 'ALL') as model,
			COALESCE(segment, 'ALL') as segment,
			COALESCE(transmission, 'ALL') as transmission,
			COUNT(*) as sales_count,
			ROUND(AVG(selling_price), 2) as avg_price
		FROM segmented_sales
		GROUP BY CUBE (model, segment, transmission)
		ORDER BY sales_count DESC`
}

// Q6: Multi-dimensional Analysis (Model+Make, Year+Body)
func GetQuerySalesMultiAnalysis(isPostgres bool) string {
	if isPostgres {
		return `
			SELECT
				COALESCE(mk.make_name, 'TOTAL') as make,
				COALESCE(m.model_name, 'ALL') as model,
				COALESCE(CAST(y.year_int AS VARCHAR), 'ALL') as year,
				COALESCE(b.body_name, 'ALL') as body,
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
			ORDER BY sales_count DESC`
	}
	return `
		SELECT
			COALESCE(make, 'TOTAL') as make,
			COALESCE(model, 'ALL') as model,
			COALESCE(CAST(year_int AS VARCHAR), 'ALL') as year,
			COALESCE(body, 'ALL') as body,
			COUNT(*) as sales_count,
			SUM(selling_price) as total_revenue
		FROM raw_sales
		GROUP BY GROUPING SETS (
			(make, model),
			(year_int, body),
			(year_int),
			()
		)
		ORDER BY sales_count DESC`
}
