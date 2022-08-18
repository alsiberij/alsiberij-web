package database

import "errors"

var (
	ErrRedisNotInitialized    = errors.New("redis not initialized")
	ErrPostgresNotInitialized = errors.New("postgres not initialized")
)
