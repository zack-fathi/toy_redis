package main

type kvPair struct {

	Key string
	Value string

}

type kvStore struct {

	Store map[string]string
	
}