CREATE EXTENSION hstore;

CREATE TABLE tasks (
	id serial primary key,
	parameters hstore
);
CREATE TABLE schedules (
	id serial primary key,
	priority integer,
	max_duration_seconds integer,
	max_measurements integer,
	max_rate_per_second integer,
	parameters hstore
);
CREATE TABLE currently_scheduled (
	schedule integer references schedules(id);
	expiration_time timestamp;
	measurements_remaining integer;
	priority integer;
);
CREATE TABLE scheduler_configuration (
	concurrent_schedules integer;
	maximum_priority_scheduled integer;
);
CREATE TABLE already_scheduled (
	schedule integer references schedules(id)
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
	origin text,
	referer text,
	client_ip text,
	client_location text,
	user_agent text
);
