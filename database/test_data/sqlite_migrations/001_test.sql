-- +goose Up
CREATE TABLE test (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL
);
