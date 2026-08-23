package main

import (
	"strings"
	"testing"
)

func newTestDatabase() *database {
	return &database{
		Store:  make(map[string]string),
		Hashes: make(map[string]map[string]string),
	}
}

func dispatchForTest(t *testing.T, db *database, fields []string) []byte {
	t.Helper()

	resp, err := dispatchCommand(fields, db, nil, false)
	if err != nil {
		t.Fatalf("dispatchCommand returned unexpected error: %v", err)
	}
	return resp
}

func TestSetAndGet(t *testing.T) {
	db := newTestDatabase()

	if got := string(dispatchForTest(t, db, []string{"SET", "name", "Zackery"})); got != "+OK\r\n" {
		t.Fatalf("SET response = %q, want %q", got, "+OK\r\n")
	}

	if got := string(dispatchForTest(t, db, []string{"GET", "name"})); got != "$7\r\nZackery\r\n" {
		t.Fatalf("GET response = %q, want %q", got, "$7\r\nZackery\r\n")
	}

	if got := string(dispatchForTest(t, db, []string{"SET", "name", "Updated"})); got != "+OK\r\n" {
		t.Fatalf("SET overwrite response = %q, want %q", got, "+OK\r\n")
	}

	if got := string(dispatchForTest(t, db, []string{"GET", "name"})); got != "$7\r\nUpdated\r\n" {
		t.Fatalf("GET after overwrite = %q, want %q", got, "$7\r\nUpdated\r\n")
	}
}

func TestMissingValues(t *testing.T) {
	db := newTestDatabase()

	if got := string(dispatchForTest(t, db, []string{"GET", "missing"})); got != "$-1\r\n" {
		t.Fatalf("missing GET response = %q, want %q", got, "$-1\r\n")
	}

	if got := string(dispatchForTest(t, db, []string{"HGET", "users", "missing"})); got != "$-1\r\n" {
		t.Fatalf("missing HGET response = %q, want %q", got, "$-1\r\n")
	}

	if got := string(dispatchForTest(t, db, []string{"HGETALL", "missing"})); got != "*0\r\n" {
		t.Fatalf("missing HGETALL response = %q, want %q", got, "*0\r\n")
	}
}

func TestHashCommands(t *testing.T) {
	db := newTestDatabase()

	if got := string(dispatchForTest(t, db, []string{"HSET", "users", "zackery", "engineer"})); got != ":1\r\n" {
		t.Fatalf("new HSET response = %q, want %q", got, ":1\r\n")
	}

	if got := string(dispatchForTest(t, db, []string{"HSET", "users", "zackery", "developer"})); got != ":0\r\n" {
		t.Fatalf("HSET overwrite response = %q, want %q", got, ":0\r\n")
	}

	if got := string(dispatchForTest(t, db, []string{"HGET", "users", "zackery"})); got != "$9\r\ndeveloper\r\n" {
		t.Fatalf("HGET response = %q, want %q", got, "$9\r\ndeveloper\r\n")
	}

	if got := string(dispatchForTest(t, db, []string{"HSET", "users", "alice", "designer"})); got != ":1\r\n" {
		t.Fatalf("second HSET response = %q, want %q", got, ":1\r\n")
	}

	got := string(dispatchForTest(t, db, []string{"HGETALL", "users"}))
	for _, expected := range []string{"*4\r\n", "$7\r\nzackery\r\n", "$9\r\ndeveloper\r\n", "$5\r\nalice\r\n", "$8\r\ndesigner\r\n"} {
		if !strings.Contains(got, expected) {
			t.Errorf("HGETALL response %q does not contain %q", got, expected)
		}
	}
}

func TestWrongArgumentCount(t *testing.T) {
	db := newTestDatabase()

	tests := []struct {
		name   string
		fields []string
		want   string
	}{
		{"SET missing value", []string{"SET", "name"}, "-ERR wrong number of arguments for 'set' command\r\n"},
		{"HSET missing value", []string{"HSET", "users", "name"}, "-ERR wrong number of arguments for 'hset' command\r\n"},
		{"HGET missing field", []string{"HGET", "users"}, "-ERR wrong number of arguments for command\r\n"},
		{"HGETALL extra field", []string{"HGETALL", "users", "extra"}, "-ERR wrong number of arguments for command\r\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := string(dispatchForTest(t, db, test.fields)); got != test.want {
				t.Fatalf("response = %q, want %q", got, test.want)
			}
		})
	}
}
