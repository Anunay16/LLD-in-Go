package database

import "sync"

type Database interface {
	GetUserName(id int) string
}

type singletonDatabase struct {
	db map[int]string
}

var (
	once     sync.Once
	instance Database
)

func GetSingletonDatabaseInstance() Database {
	once.Do(func() {
		instance = &singletonDatabase{
			db: map[int]string{
				1: "alice",
				2: "bob",
			},
		}
	})
	return instance
}

func (s *singletonDatabase) GetUserName(id int) string {
	return s.db[id]
}
