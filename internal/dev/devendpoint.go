package dev

import (
	"net/url"

	"github.com/alecthomas/types/optional"
)

type LocalEndpoint struct {
	Module         string
	Endpoint       url.URL
	DebugPort      int
	Language       string
	RunnerInfoFile optional.Option[string]
}
