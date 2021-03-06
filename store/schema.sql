CREATE EXTENSION hstore;

CREATE SCHEMA task_functions;

CREATE TABLE tasks (
	id serial primary key,
	parameters hstore
);
CREATE TABLE task_functions (
	id serial primary key,
	priority integer,
	max_duration_seconds integer,
	max_measurements integer,
	max_rate_per_second integer,
	task_function information_schema.sql_identifier,
	enabled boolean
);
CREATE TABLE scheduled_functions (
	task_function integer references task_functions(id),
	expiration_time timestamp,
	measurements_remaining integer,
	priority integer,
	scheduled_time timestamp
);
CREATE TABLE scheduler_configuration (
	concurrent_functions integer
);
CREATE TABLE queries (
	id serial primary key,
	"timestamp" timestamp,
	client_ip text,
	raw_request bytea,
	task integer references tasks(id),
	substrate text,
	parameters_json text,
	response_body bytea
);
CREATE TABLE parsed_queries (
	query integer references queries(id),
	"timestamp" timestamp,
	measurement_id text,
	client_ip text,
	client_location text,
	substrate text,
	parameters hstore
);
CREATE TABLE results (
	id serial primary key,
	"timestamp" timestamp,
	client_ip text,
	raw_request bytea
);
CREATE TABLE parsed_results (
	result integer references results(id),
	"timestamp" timestamp,
	measurement_id text,
	outcome text,
	message text,
	origin text,
	referer text,
	client_ip text,
	client_location text,
	user_agent text
);
