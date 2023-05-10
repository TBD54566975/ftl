package sha256

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strconv"

	"github.com/alecthomas/errors"
)

type SHA256 [sha256.Size]byte

func Sum(data []byte) SHA256 { return sha256.Sum256(data) }
func SumReader(r io.Reader) (SHA256, error) {
	h := sha256.New()
	_, err := io.Copy(h, r)
	var out SHA256
	copy(out[:], h.Sum(nil))
	return out, errors.WithStack(err)
}

func FromBytes(data []byte) SHA256 {
	var out SHA256
	copy(out[:], data)
	return out
}

func ParseSHA256(s string) (SHA256, error) {
	var out SHA256
	err := out.UnmarshalText([]byte(s))
	return out, err
}

func MustParseSHA256(s string) SHA256 {
	out, err := ParseSHA256(s)
	if err != nil {
		panic(err)
	}
	return out
}

func (s *SHA256) UnmarshalText(text []byte) error {
	_, err := hex.Decode(s[:], text)
	return errors.WithStack(err)
}
func (s SHA256) MarshalText() ([]byte, error) { return []byte(hex.EncodeToString(s[:])), nil }
func (s SHA256) String() string               { return hex.EncodeToString(s[:]) }
func (s SHA256) GoString() string             { return strconv.Quote(hex.EncodeToString(s[:])) }
