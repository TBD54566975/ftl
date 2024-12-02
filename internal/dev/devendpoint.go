package dev

import "net/url"

type LocalEndpoint struct {
	Module    string
	Endpoint  url.URL
	DebugPort int
	Language  string
}
