package storage

import "errors"

var (
	ErrURLNotFound = errors.New("shortener not found")
	ErrURLExists   = errors.New("shortener already exists")
)
