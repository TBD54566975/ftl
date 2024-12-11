package sha256

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strconv"
)

// SHA256 is a type-safe wrapper around a SHA256 hash.
type SHA256 [sha256.Size]byte

// Sum "data" and return the SHA256 hash.
func Sum(data []byte) SHA256 { return sha256.Sum256(data) }

// SumReader "r" and return the SHA256 hash.
func SumReader(r io.Reader) (SHA256, error) {
	h := sha256.New()
	_, err := io.Copy(h, r)
	var out SHA256
	copy(out[:], h.Sum(nil))
	return out, err
}

func SumFile(path string) (SHA256, error) {
	f, err := os.Open(path)
	if err != nil {
		return SHA256{}, err
	}
	defer f.Close() //nolint:gosec
	return SumReader(f)
}

// FromBytes converts a SHA256 in []byte form to a SHA256.
func FromBytes(data []byte) SHA256 {
	var out SHA256
	copy(out[:], data)
	return out
}

// ParseSHA256 parses a hex-ecndoded SHA256 hash from a string.
func ParseSHA256(s string) (SHA256, error) {
	var out SHA256
	err := out.UnmarshalText([]byte(s))
	return out, err
}

// MustParseSHA256 parses a hex-ecndoded SHA256 hash from a string, panicing on error.
func MustParseSHA256(s string) SHA256 {
	out, err := ParseSHA256(s)
	if err != nil {
		panic(err)
	}
	return out
}

func (s *SHA256) UnmarshalText(text []byte) error {
	_, err := hex.Decode(s[:], text)
	return err
}
func (s SHA256) MarshalText() ([]byte, error) { return []byte(hex.EncodeToString(s[:])), nil }
func (s SHA256) String() string               { return hex.EncodeToString(s[:]) }
func (s SHA256) GoString() string             { return strconv.Quote(hex.EncodeToString(s[:])) }
