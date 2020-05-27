package storage

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"encoding/gob"
	"sync/atomic"
)

// Storage - tiny key-value db
type Storage struct {
	Done        chan bool
	mut         sync.RWMutex
	dumpFile    string
	expiredKeys []*Item // защищён sync.RWMutex
	m           map[string]*Item
	Length      int64 // защищён sync/atomic
}

// Open - load storage from file, and set
func Open(fileName string, dumpInterval time.Duration) (s *Storage, err error) {
	if fileExists(fileName) {
		log.Printf("Open existing file: %s", fileName)
		s = &Storage{Done: make(chan bool, 1)}
		if err := s.Load(fileName); err != nil {
			return nil, err
		}
	} else {
		log.Printf("Create new file: %s", fileName)
		s = &Storage{
			Done:        make(chan bool, 1),
			mut:         sync.RWMutex{},
			dumpFile:    fileName,
			expiredKeys: []*Item{},
			m:           map[string]*Item{},
			Length:      0,
		}
		s.Dump() // create file
	}
	dumpTicker := time.NewTicker(dumpInterval)
	expireTicker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-s.Done:
				return
			case <-expireTicker.C:
				s.Clean()
			case <-dumpTicker.C:
				s.Dump()
			}
		}
	}()
	return s, nil
}

func (s *Storage) Load(fileName string) error {
	s.dumpFile = fileName
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	var decoder = gob.NewDecoder(file)
	if err := decoder.Decode(&s.Length); err != nil {
		return err
	}
	if err := decoder.Decode(&s.m); err != nil {
		return err
	}
	if err := decoder.Decode(&s.expiredKeys); err != nil {
		return err
	}

	return nil
}

// Dump - save storage to disk
func (s *Storage) Dump() error {
	file, err := os.OpenFile(s.dumpFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
		return err
	}
	defer file.Close()

	s.mut.RLock()
	defer s.mut.RUnlock()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(&s.Length); err != nil {
		panic(err)
		return err
	}
	if err := encoder.Encode(&s.m); err != nil {
		panic(err)
		return err
	}
	if err := encoder.Encode(&s.expiredKeys); err != nil {
		panic(err)
		return err
	}

	return nil
}

// Close - dump storage to disk and stop ticker
func (s *Storage) Close() error {
	s.Done <- true
	return s.Dump()
}

// Get - load value from kv
func (s *Storage) Get(key string) (string, error) {
	log.Printf("Get [%s] request", key)
	s.mut.RLock()
	defer s.mut.RUnlock()
	val, ok := s.m[key]
	if !ok {
		return "", fmt.Errorf("%s not in storage", key)
	}
	return val.Value, nil
}

// GetItem -
func (s *Storage) GetItem(key string) (*Item, error) {
	log.Printf("Get [%s] TTL request", key)
	s.mut.RLock()
	defer s.mut.RUnlock()
	val, ok := s.m[key]
	if !ok {
		return nil, fmt.Errorf("%s not in storage", key)
	}
	return val, nil
}

// Values - get all values from kv
func (s *Storage) Values() []string {
	log.Println("Get values request")
	s.mut.RLock()
	defer s.mut.RUnlock()
	var values = make([]string, 0, s.Length)
	for _, value := range s.m {
		values = append(values, value.Value)
	}
	return values
}

func (s *Storage) putNew(item *Item) {
	atomic.AddInt64(&s.Length, 1)
	s.m[item.Key] = item
}

// Put - store value to kv (without TTL)
func (s *Storage) Put(key, value string) {
	log.Printf("Put %s: %s", key, value)
	s.mut.Lock()
	defer s.mut.Unlock()
	val, ok := s.m[key]
	if !ok {
		s.putNew(NewItem(key, value))
		return
	}
	val.Value = value // змаеняем Value в мапе и списке (сохраняем TTL, если был)
}

// PutWithTTL - store kv pair with expired time
func (s *Storage) PutWithTTL(key, value string, ttl time.Duration) {
	log.Printf("Put with TTL - %s: %s [duration: %s]", key, value, ttl)
	s.mut.Lock()
	defer s.mut.Unlock()
	item, ok := s.m[key]
	if !ok {
		item := NewItem(key, value)
		item.SetTTL(ttl)
		s.putNew(item)
		return
	}
	if item.TTL != nil {
		s.expiredKeys = deleteByKey(item.Key, s.expiredKeys)
	}
	item.Value = value
	item.SetTTL(ttl)
	s.m[key] = item
	s.expiredKeys = push(s.expiredKeys, item)
}

// Delete - delete item from storage and expiredKeys
func (s *Storage) Delete(key string) error {
	log.Printf("Delete %s request", key)
	s.mut.Lock()
	defer s.mut.Unlock()
	if _, ok := s.m[key]; !ok {
		return fmt.Errorf("%s not in storage", key)
	}
	delete(s.m, key)
	s.expiredKeys = deleteByKey(key, s.expiredKeys)
	return nil
}

// SetTTL - set ttl to item by key
func (s *Storage) SetTTL(key string, ttl time.Duration) error {
	log.Printf("Set TTL %s [duration %s]", key, ttl)
	s.mut.Lock()
	defer s.mut.Unlock()
	item, ok := s.m[key]
	if !ok {
		return fmt.Errorf("%s not in storage")
	}

	if item.TTL == nil {
		s.expiredKeys = push(s.expiredKeys, item)
		item.SetTTL(ttl)
		return nil
	}
	item.SetTTL(ttl)
	s.expiredKeys = push(deleteByKey(key, s.expiredKeys), item)
	return nil
}

// Clean - delete all expired items from storage
func (s *Storage) Clean() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.expiredKeys, s.m = deleteExpiredItems(s.expiredKeys, s.m)
}
