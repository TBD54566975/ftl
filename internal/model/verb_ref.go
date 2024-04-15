package model

import "fmt"

type VerbRef struct {
	Module string
	Name   string
}

func (v VerbRef) String() string {
	return fmt.Sprintf("%s.%s", v.Module, v.Name)
}
