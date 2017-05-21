package main

import "sync"

type SCMap struct {
	table map[string]*SoundCollection
	lock  *sync.RWMutex
}

func NewSCMap() *SCMap {
	return &SCMap{table: make(map[string]*SoundCollection), lock: new(sync.RWMutex)}
}

func (m *SCMap) Lock() {
	m.lock.RLock()
}

func (m *SCMap) Unlock() {
	m.lock.RUnlock()
}

func (m *SCMap) Get(key string) *SoundCollection {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.table[key]
}

func (m *SCMap) Write(key string, value *SoundCollection) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.table[key] = value
}
