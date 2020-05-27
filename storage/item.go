package storage

import "time"

type Item struct {
	Key   string
	Value string
	TTL   *time.Time
}

func NewItem(key, value string) *Item {
	return &Item{
		Key:   key,
		Value: value,
	}
}

func (item *Item) SetTTL(ttl time.Duration) {
	expired := time.Now().UTC().Add(ttl)
	item.TTL = &expired
}

func (item *Item) ExpiredAfter() time.Duration {
	return item.TTL.Sub(time.Now().UTC())
}
