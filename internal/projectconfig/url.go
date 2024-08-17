package projectconfig

import (
	"fmt"
	"net/url"
)

// A URL that supports marshalling and unmarshalling which is, very
// frustratingly, not possible with the stdlib url.URL.
type URL url.URL

func ParseURL(rawurl string) (*URL, error) {
	parsed, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	return (*URL)(parsed), nil
}

func MustParseURL(rawurl string) *URL {
	parsed, err := ParseURL(rawurl)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (u *URL) UnmarshalText(text []byte) error {
	parsed, err := url.Parse(string(text))
	if err != nil {
		return err
	}
	*u = URL(*parsed)
	return nil
}

func (u URL) MarshalText() ([]byte, error) {
	return []byte((*url.URL)(&u).String()), nil
}

func (u URL) GoString() string {
	return fmt.Sprintf("projectconfig.MustParseURL(%q)", (*url.URL)(&u).String())
}
