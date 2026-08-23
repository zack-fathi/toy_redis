package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
)

func TestConcurrentStoreAccess(t *testing.T) {
	db := &database{
		Store:  make(map[string]string),
		Hashes: make(map[string]map[string]string),
	}
	commands := []string{"SET", "HSET"}

	var wg sync.WaitGroup
	for workerID := 0; workerID < 10; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			random := rand.New(rand.NewPCG(uint64(id+1), uint64(id+100)))
			for operation := 0; operation < 100; operation++ {
				command := commands[random.IntN(len(commands))]
				key := fmt.Sprintf("key-%d-%d", id, random.IntN(20))
				value := fmt.Sprintf("value-%d-%d", id, random.IntN(1000))

				if command == "SET" {
					resp := []byte{}
					db.setString([]string{"SET", key, value}, &resp)
					t.Logf("worker %d executed SET key=%s value=%s", id, key, value)
					continue
				}

				hash := fmt.Sprintf("hash-%d-%d", id, random.IntN(10))
				field := fmt.Sprintf("field-%d", random.IntN(20))
				resp := []byte{}
				fields := []string{"HSET", hash, field, value}
				db.setHashFields(fields, &resp)
				t.Logf("worker %d executed HSET hash=%s field=%s value=%s", id, hash, field, value)
			}

			for operation := 0; operation < 100; operation++ {
				key := fmt.Sprintf("key-%d-%d", id, random.IntN(20))
				resp := []byte{}
				db.getString([]string{"GET", key}, &resp)

				hash := fmt.Sprintf("hash-%d-%d", id, random.IntN(10))
				field := fmt.Sprintf("field-%d", random.IntN(20))
				resp = []byte{}
				db.getHashField([]string{"HGET", hash, field}, &resp)
				resp = []byte{}
				db.getAllHashFields([]string{"HGETALL", hash}, &resp)
			}

		}(workerID)
	}

	wg.Wait()
}
