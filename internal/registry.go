package internal

import (
	"time"
)

type HttpBasicCredentials struct {
	Username string
	Password string
}

type Registry interface {
	GetInterval() time.Duration
	FetchVersions(resource string) (*[]string, error)
}
