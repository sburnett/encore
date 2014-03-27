package store

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/lib/pq/hstore"
)

type postgresStore struct {
	db *sql.DB
}

func openPostgres(db *sql.DB) Store {
	return &postgresStore{
		db: db,
	}
}

func (store *postgresStore) Close() {
	store.db.Close()
}

func (store *postgresStore) Schedules() <-chan *Schedule {
	schedules := make(chan *Schedule)
	go func() {
		schedulesStmt, err := store.db.Prepare("SELECT id, priority, max_duration_seconds, max_measurements, max_rate_per_second, parameters FROM schedules WHERE NOT EXISTS (SELECT NULL FROM already_scheduled WHERE schedule = id) ORDER BY priority ASC")
		if err != nil {
			log.Fatalf("error preparing schedules select statement: %v", err)
		}
		tasksStmt, err := store.db.Prepare("SELECT id, parameters FROM tasks WHERE parameters @> $1")
		if err != nil {
			log.Fatalf("error preparing scheduling parameters select statement: %v", err)
		}
		updateStmt, err := store.db.Prepare("INSERT INTO already_scheduled (schedule) VALUES ($1)")
		if err != nil {
			log.Fatalf("error preparing scheduling update statement: %v", err)
		}
		deleteStmt, err := store.db.Prepare("DELETE FROM already_scheduled")
		if err != nil {
			log.Fatalf("error preparing scheduling delete statement: %v", err)
		}

		for {
			rows, err := schedulesStmt.Query()
			if err != nil {
				log.Fatalf("error selecting schedules: %v", err)
			}
			for rows.Next() {
				var schedule Schedule
				var maxDurationSeconds int
				var parameters hstore.Hstore
				if err := rows.Scan(&schedule.Id, &schedule.Priority, &maxDurationSeconds, &schedule.MaxMeasurements, &schedule.MaxRatePerSecond, &parameters); err != nil {
					log.Fatalf("error scanning schedule: %v", err)
				}
				schedule.MaxDuration = time.Duration(maxDurationSeconds) * time.Second

				rows, err := tasksStmt.Query(parameters)
				if err != nil {
					log.Fatalf("error selecting tasks that match a schedule", err)
				}
				for rows.Next() {
					var task Task
					var parameters hstore.Hstore
					if err := rows.Scan(&task.Id, &parameters); err != nil {
						log.Printf("error scanning task parameters")
						continue
					}
					task.TemplateParameters = parameters.Map
					schedule.Tasks = append(schedule.Tasks, task)
				}

				if schedule.Tasks == nil {
					log.Printf("skipping schedule with no matching tasks: id %v", schedule.Id)
					continue
				}

				schedules <- &schedule

				if _, err := updateStmt.Exec(schedule.Id); err != nil {
					log.Fatalf("error updating scheduling priority: %v", err)
				}
			}

			log.Printf("done with all schedules; recycling old schedules")
			if result, err := deleteStmt.Exec(); err != nil {
				log.Fatalf("error clearing already_scheduled table: %v", err)
			} else if affected, err := result.RowsAffected(); err != nil {
				log.Fatalf("error fetching number of affected rows: %v", err)
			} else if affected == 0 {
				log.Printf("no schedules available")
				time.Sleep(time.Second)
			}
		}
		close(schedules)
	}()
	return schedules
}

func (store *postgresStore) WriteTasks(tasks <-chan *Task) {
	tasksStmt, err := store.db.Prepare("INSERT INTO tasks (parameters) VALUES ($1)")
	if err != nil {
		log.Fatalf("error preparing tasks insert statement: %v", err)
	}
	defer tasksStmt.Close()

	for task := range tasks {
		parameters := hstore.Hstore{task.TemplateParameters}
		tasksStmt.Exec(parameters)
	}
}

func (store *postgresStore) WriteQueries(queries <-chan *Query) {
	queriesStmt, err := store.db.Prepare("INSERT INTO queries (timestamp, client_ip, task, raw_request, substrate, parameters_json, response_body) VALUES ($1, $2, $3, $4, $5, $6, $7)")
	if err != nil {
		log.Fatalf("error preparing queries insert statement: %v", err)
	}
	defer queriesStmt.Close()

	for query := range queries {
		if _, err := queriesStmt.Exec(query.Timestamp, query.RemoteAddr, query.Task, query.RawRequest, query.Substrate, query.ParametersJson, query.ResponseBody); err != nil {
			log.Printf("error inserting query: %v", err)
			continue
		}
	}
}

func (store *postgresStore) Queries() <-chan *Query {
	queries := make(chan *Query)
	go func() {
		defer close(queries)

		rows, err := store.db.Query("SELECT id, timestamp, client_ip, task, raw_request, substrate, parameters_json FROM queries")
		if err != nil {
			log.Fatalf("error selecting queries: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var query Query
			if err := rows.Scan(&query.Id, &query.Timestamp, &query.RemoteAddr, &query.Task, &query.RawRequest, &query.Substrate, &query.ParametersJson); err != nil {
				log.Printf("error reading query: %v", err)
			}
			queries <- &query
		}
	}()
	return queries
}

func (store *postgresStore) UnparsedQueries() <-chan *Query {
	queries := make(chan *Query)
	go func() {
		defer close(queries)

		rows, err := store.db.Query("SELECT id, timestamp, client_ip, task, raw_request, substrate, parameters_json FROM queries WHERE NOT EXISTS (SELECT NULL FROM parsed_queries WHERE query = id)")
		if err != nil {
			log.Fatalf("error selecting queries: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var query Query
			if err := rows.Scan(&query.Id, &query.Timestamp, &query.RemoteAddr, &query.Task, &query.RawRequest, &query.Substrate, &query.ParametersJson); err != nil {
				log.Printf("error reading query: %v", err)
			}
			queries <- &query
		}
	}()
	return queries
}

func (store *postgresStore) WriteParsedQueries(parsedQueries <-chan *ParsedQuery) {
	insertIntoQueries, err := store.db.Prepare(`INSERT INTO parsed_queries (query, measurement_id, timestamp, client_ip, client_location, substrate, parameters) VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		log.Fatalf("error preparing parsed_queries insert statement: %v", err)
	}
	defer insertIntoQueries.Close()

	for parsedQuery := range parsedQueries {
		if _, err := insertIntoQueries.Exec(parsedQuery.Query, parsedQuery.MeasurementId, parsedQuery.Timestamp, parsedQuery.ClientIp.String(), parsedQuery.ClientLocation, parsedQuery.Substrate, hstore.Hstore{parsedQuery.Parameters}); err != nil {
			log.Printf("error inserting parsed query: %v", err)
		}
	}
}

func (store *postgresStore) WriteResults(results <-chan *Result) {
	resultsStmt, err := store.db.Prepare("INSERT INTO results (timestamp, client_ip, raw_request) VALUES ($1, $2, $3)")
	if err != nil {
		log.Fatalf("error preparing results insert statement: %v", err)
	}
	defer resultsStmt.Close()

	for result := range results {
		if _, err := resultsStmt.Exec(result.Timestamp, result.RemoteAddr, result.RawRequest); err != nil {
			log.Printf("error inserting result: %v", err)
			continue
		}
	}
}

func (store *postgresStore) Results() <-chan *Result {
	results := make(chan *Result)
	go func() {
		defer close(results)

		rows, err := store.db.Query("SELECT id, timestamp, client_ip, raw_request FROM results")
		if err != nil {
			log.Fatalf("error selecting results: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var result Result
			if err := rows.Scan(&result.Id, &result.Timestamp, &result.RemoteAddr, &result.RawRequest); err != nil {
				log.Printf("error scanning result: %v", err)
			}
			results <- &result
		}
	}()
	return results
}

func (store *postgresStore) UnparsedResults() <-chan *Result {
	results := make(chan *Result)
	go func() {
		defer close(results)

		rows, err := store.db.Query("SELECT id, timestamp, client_ip, raw_request FROM results WHERE NOT EXISTS (SELECT NULL FROM parsed_results WHERE result = id)")
		if err != nil {
			log.Fatalf("error selecting results: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var result Result
			if err := rows.Scan(&result.Id, &result.Timestamp, &result.RemoteAddr, &result.RawRequest); err != nil {
				log.Printf("error scanning result: %v", err)
			}
			results <- &result
		}
	}()
	return results
}

func (store *postgresStore) WriteParsedResults(parsedResults <-chan *ParsedResult) {
	insertIntoResults, err := store.db.Prepare("INSERT INTO parsed_results (result, measurement_id, timestamp, outcome, origin, referer, client_ip, client_location, user_agent) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)")
	if err != nil {
		log.Fatalf("error preparing parsed_results insertion statement: %v", err)
	}
	defer insertIntoResults.Close()

	for parsedResult := range parsedResults {
		if _, err := insertIntoResults.Exec(parsedResult.Result, parsedResult.MeasurementId, parsedResult.Timestamp, parsedResult.Outcome, parsedResult.Origin, parsedResult.Referer, parsedResult.ClientIp.String(), parsedResult.ClientLocation, parsedResult.UserAgent); err != nil {
			log.Printf("error inserting parsed result: %v", err)
		}
	}
}

func (store *postgresStore) ComputeResultsTables() error {
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("DROP TABLE IF EXISTS results_per_referer"); err != nil {
		return err
	}
	if _, err := tx.Exec("SELECT referer, count(distinct measurement_id) results INTO results_per_referer FROM parsed_results WHERE outcome = 'init' GROUP BY referer"); err != nil {
		return err
	}
	if _, err := tx.Exec("CREATE INDEX ON results_per_referer (referer)"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	tx, err = store.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("DROP TABLE IF EXISTS results_per_day"); err != nil {
		return err
	}
	if _, err := tx.Exec(`SELECT referer, "timestamp"::date AS day, count(distinct measurement_id) results INTO results_per_day FROM parsed_results WHERE outcome = 'init' GROUP BY referer, timestamp::date`); err != nil {
		return err
	}
	if _, err := tx.Exec("CREATE INDEX ON results_per_day (referer)"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	tx, err = store.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("DROP TABLE IF EXISTS results_per_country"); err != nil {
		return err
	}
	if _, err := tx.Exec("SELECT referer, client_location country, count(distinct measurement_id) results INTO results_per_country FROM parsed_results WHERE outcome = 'init' GROUP BY referer, client_location"); err != nil {
		return err
	}
	if _, err := tx.Exec("CREATE INDEX ON results_per_country (referer)"); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (store *postgresStore) CountResultsForReferrer(requests <-chan CountResultsRequest) {
	query, err := store.db.Prepare("SELECT results FROM results_per_referer WHERE referer = $1")
	if err != nil {
		log.Fatalf("error preparing result count statement: %v", err)
	}

	for request := range requests {
		row := query.QueryRow(request.Referer)
		var count int
		if err := row.Scan(&count); err != nil {
			log.Printf("error scanning result count %s: %v", request.Referer, err)
			request.Response <- CountResultsResponse{
				Err: err,
			}
			continue
		}
		request.Response <- CountResultsResponse{
			Count: count,
			Err:   nil,
		}
	}
}

func (store *postgresStore) ResultsPerDayForReferrer(requests <-chan ResultsPerDayRequest) {
	query, err := store.db.Prepare("SELECT day, results FROM results_per_day WHERE referer = $1 ORDER BY day")
	if err != nil {
		log.Fatalf("error preparing result count statement: %v", err)
	}

	for request := range requests {
		rows, err := query.Query(request.Referer)
		if err != nil {
			request.Response <- ResultsPerDayResponse{
				Err: err,
			}
			continue
		}
		results := make(map[string]int)
		var resultsErr error
		for rows.Next() {
			var day time.Time
			var count int
			if err := rows.Scan(&day, &count); err != nil {
				results = nil
				resultsErr = err
				break
			}
			results[day.Format("2006-01-02")] = count
		}

		if results == nil {
			request.Response <- ResultsPerDayResponse{
				Err: resultsErr,
			}
			continue
		}

		request.Response <- ResultsPerDayResponse{
			Results: results,
			Err:     err,
		}
	}
}

func (store *postgresStore) ResultsPerCountryForReferrer(requests <-chan ResultsPerCountryRequest) {
	query, err := store.db.Prepare("SELECT country, results FROM results_per_country WHERE referer = $1 ORDER BY results DESC")
	if err != nil {
		log.Fatalf("error preparing result count statement: %v", err)
	}

	for request := range requests {
		rows, err := query.Query(request.Referer)
		if err != nil {
			request.Response <- ResultsPerCountryResponse{
				Err: err,
			}
			continue
		}
		results := make(map[string]int)
		for rows.Next() {
			var country string
			var count int
			if err := rows.Scan(&country, &count); err != nil {
				results = nil
				break
			}
			results[country] = count
		}

		if results == nil {
			request.Response <- ResultsPerCountryResponse{
				Err: err,
			}
			continue
		}

		request.Response <- ResultsPerCountryResponse{
			Results: results,
			Err:     err,
		}
	}
}
