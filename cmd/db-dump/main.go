package main

import (
	"fmt"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/vmihailenco/msgpack/v5"
)

func main() {
	storePath := os.Getenv("LEVELDB_STORE_PATH")
	if storePath == "" {
		storePath = "./data/store"
	}
	if len(os.Args) > 1 {
		storePath = os.Args[1]
	}

	db, err := leveldb.OpenFile(storePath, nil)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		value := iter.Value()

		fmt.Printf("KEY: %s\n", key)

		var decoded interface{}
		if err := msgpack.Unmarshal(value, &decoded); err == nil {
			fmt.Printf("VALUE (decoded): %+v\n", decoded)
		} else {
			fmt.Printf("VALUE (raw hex): %x\n", value)
		}
		fmt.Println("---")
	}

	if err := iter.Error(); err != nil {
		fmt.Printf("Iterator error: %v\n", err)
	}
}
