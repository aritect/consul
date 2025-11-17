package store

import (
	"consul-telegram-bot/internal/metrics"
	"os"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type Store struct {
	path         string
	db           *leveldb.DB
	options      *opt.Options
	writeOptions *opt.WriteOptions
	readOptions  *opt.ReadOptions
}

var globalInstance *Store

func GetInstance() *Store {
	return globalInstance
}

func New(path string, reset bool, compression bool) (*Store, error) {
	options := &opt.Options{
		Filter:      filter.NewBloomFilter(10),
		Compression: opt.NoCompression,
	}

	writeOptions := &opt.WriteOptions{
		Sync: false,
	}

	if reset {
		err := os.RemoveAll(path)
		if err != nil {
			return nil, err
		}

		writeOptions.Sync = true
	}

	if compression {
		options.Compression = opt.SnappyCompression
	}

	readOptions := &opt.ReadOptions{
		Strict: opt.DefaultStrict,
	}

	db, err := leveldb.OpenFile(path, options)
	if err != nil {
		return nil, err
	}

	s := &Store{
		path:         path,
		db:           db,
		options:      options,
		writeOptions: writeOptions,
		readOptions:  readOptions,
	}

	return s, nil
}

func (s *Store) Put(key, value []byte) error {
	start := time.Now()
	err := s.db.Put(key, value, s.writeOptions)

	status := "success"
	if err != nil {
		status = "error"
		metrics.ErrorsTotal.WithLabelValues("leveldb", "put").Inc()
	}

	metrics.LevelDBOperations.WithLabelValues("put", status).Inc()
	metrics.ProcessingDuration.WithLabelValues("leveldb_put").Observe(float64(time.Since(start).Nanoseconds() / 1000))

	return err
}

func (s *Store) Get(key []byte) ([]byte, error) {
	start := time.Now()
	value, err := s.db.Get(key, s.readOptions)

	status := "success"
	if err != nil {
		status = "error"
		if err != leveldb.ErrNotFound {
			metrics.ErrorsTotal.WithLabelValues("leveldb", "get").Inc()
		}
	}

	metrics.LevelDBOperations.WithLabelValues("get", status).Inc()
	metrics.ProcessingDuration.WithLabelValues("leveldb_get").Observe(float64(time.Since(start).Nanoseconds() / 1000))

	return value, err
}

func (s *Store) Has(key []byte) (bool, error) {
	return s.db.Has(key, s.readOptions)
}

func (s *Store) Delete(key []byte) error {
	start := time.Now()
	err := s.db.Delete(key, s.writeOptions)

	status := "success"
	if err != nil {
		status = "error"
		metrics.ErrorsTotal.WithLabelValues("leveldb", "delete").Inc()
	}

	metrics.LevelDBOperations.WithLabelValues("delete", status).Inc()
	metrics.ProcessingDuration.WithLabelValues("leveldb_delete").Observe(float64(time.Since(start).Nanoseconds() / 1000))

	return err
}

func (s *Store) Iterator() iterator.Iterator {
	return s.db.NewIterator(nil, nil)
}

func (s *Store) MakeGlobal() {
	globalInstance = s
}

func (s *Store) GetStats() (*leveldb.DBStats, error) {
	stats := &leveldb.DBStats{}
	err := s.db.Stats(stats)
	return stats, err
}
