package storage

import (
	"sort"
	"time"
)

func findIndexByKey(key string, items []*Item) int {
	for i := 0; i < len(items); i++ {
		item := items[i]
		if item.Key == key {
			return i
		}
	}
	return -1
}

// with preserving ordering
func deleteElemFromItems(index int, items []*Item) []*Item {
	copy(items[index:], items[index+1:])
	items[len(items)-1] = nil
	return items[:len(items)-1]
}

func deleteByKey(key string, items []*Item) []*Item {
	if i := findIndexByKey(key, items); i != -1 {
		return deleteElemFromItems(i, items)
	}
	return items
}

func push(items []*Item, item *Item) []*Item {
	items = append(items, item)
	sort.Slice(items, func(i, j int) bool {
		return items[i].TTL.Unix() > items[j].TTL.Unix()
	})
	return items
}

func deleteExpiredItems(items []*Item, m map[string]*Item) ([]*Item, map[string]*Item) {
	now := time.Now().UTC().Unix()
	i := sort.Search(len(items), func(i int) bool {
		return items[i].TTL.Unix() <= now
	})
	if i < len(items) {
		for _, item := range items[i:] {
			delete(m, item.Key)
		}
		return items[:i], m
	}
	return items, m
}
