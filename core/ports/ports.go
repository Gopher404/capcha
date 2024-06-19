package ports

import "time"

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value string, duration time.Duration) error
	Delete(key string) error
}
