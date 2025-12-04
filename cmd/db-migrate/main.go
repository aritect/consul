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

	fmt.Println("Migration: AritectBuysThreadId to BuysThreadId.")

	migrated := 0
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		var data map[string]interface{}
		if err := msgpack.Unmarshal(value, &data); err != nil {
			continue
		}

		if aritectBuysThreadId, ok := data["AritectBuysThreadId"]; ok {
			if _, hasBuys := data["BuysThreadId"]; !hasBuys {
				data["BuysThreadId"] = aritectBuysThreadId
			}
			delete(data, "AritectBuysThreadId")

			newValue, err := msgpack.Marshal(data)
			if err != nil {
				fmt.Printf("Failed to marshal %s: %v\n", key, err)
				continue
			}

			keyCopy := make([]byte, len(key))
			copy(keyCopy, key)

			if err := db.Put(keyCopy, newValue, nil); err != nil {
				fmt.Printf("Failed to update %s: %v\n", keyCopy, err)
				continue
			}

			fmt.Printf("Migrated: %s (BuysThreadId=%v)\n", keyCopy, aritectBuysThreadId)
			migrated++
		}
	}

	fmt.Printf("Migration complete: %d records updated.\n", migrated)
}
