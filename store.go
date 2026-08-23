package main

import (
	"log/slog"
	"strconv"
	"sync"
)

type database struct {
	Store  map[string]string
	Hashes map[string]map[string]string

	storeMu sync.RWMutex
	hashMu  sync.RWMutex
}

func (db *database) setString(fields []string, resp *[]byte) {

	db.storeMu.Lock()
	defer db.storeMu.Unlock()

	key := fields[1]
	value := fields[2]
	db.Store[key] = value
	*resp = []byte("+OK\r\n")

}

func (db *database) getString(fields []string, resp *[]byte) {

	db.storeMu.RLock()
	defer db.storeMu.RUnlock()

	key := fields[1]
	value, ok := db.Store[key]
	if !ok {
		slog.Debug("string key not found", "key", key)
		*resp = []byte("$-1\r\n")
		return
	}
	*resp = []byte(
		"$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n",
	)

}
func (db *database) setHashFields(fields []string, resp *[]byte) {

	db.hashMu.Lock()
	defer db.hashMu.Unlock()

	hash := fields[1]
	if _, exists := db.Hashes[hash]; !exists {
		db.Hashes[hash] = make(map[string]string)
	}

	newFields := 0
	for i := 2; i < len(fields); i += 2 {
		field := fields[i]
		newValue := fields[i+1]

		if _, exists := db.Hashes[hash][field]; !exists {
			newFields++
		}

		db.Hashes[hash][field] = newValue
	}
	*resp = []byte(":" + strconv.Itoa(newFields) + "\r\n")
}
func (db *database) getHashField(fields []string, resp *[]byte) {

	db.hashMu.RLock()
	defer db.hashMu.RUnlock()

	hash := fields[1]
	field := fields[2]
	if _, exists := db.Hashes[hash][field]; !exists {
		slog.Debug("hash field not found", "hash", hash, "field", field)
		*resp = []byte("$-1\r\n")
		return
	}

	value := db.Hashes[hash][field]
	*resp = []byte(
		"$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n",
	)

}
func (db *database) getAllHashFields(fields []string, resp *[]byte) {

	db.hashMu.RLock()
	defer db.hashMu.RUnlock()

	hash := fields[1]
	if _, exists := db.Hashes[hash]; !exists {
		slog.Debug("hash key not found", "hash", hash)
		*resp = []byte("*0\r\n")
		return
	}

	var respArray []string
	for field, value := range db.Hashes[hash] {
		respArray = append(respArray, field, value)
	}

	*resp = []byte("*" + strconv.Itoa(len(respArray)) + "\r\n")
	for _, value := range respArray {
		encodedValue := []byte(
			"$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n",
		)
		*resp = append(*resp, encodedValue...)
	}
}
