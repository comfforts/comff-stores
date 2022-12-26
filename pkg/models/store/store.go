package store

import "time"

type Store struct {
	ID        string    `json:"id,omitempty"`
	StoreId   uint64    `json:"store_id"`
	Name      string    `json:"name"`
	Org       string    `json:"org"`
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	Created   time.Time `json:"created"`
}

type StoreGeo struct {
	Store    *Store
	Distance float64
}

type StoreStats struct {
	Count     int
	HashCount int
	Ready     bool
}
