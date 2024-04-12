package model

import (
	"errors"
	"strings"
)

type CronJobKey = KeyType[CronJobPayload, *CronJobPayload]

func NewCronJobKey(module, verb string) CronJobKey {
	return newKey[CronJobPayload](strings.Join([]string{module, verb}, "-"))
}

func ParseCronJobKey(key string) (CronJobKey, error) { return parseKey[CronJobPayload](key) }

type CronJobPayload struct {
	Ref string
}

var _ KeyPayload = (*CronJobPayload)(nil)

func (d *CronJobPayload) Kind() string   { return "crn" }
func (d *CronJobPayload) String() string { return d.Ref }
func (d *CronJobPayload) Parse(parts []string) error {
	if len(parts) == 0 {
		return errors.New("expected <module>-<verb> but got empty string")
	}
	d.Ref = strings.Join(parts, "-")
	return nil
}
func (d *CronJobPayload) RandomBytes() int { return 10 }
