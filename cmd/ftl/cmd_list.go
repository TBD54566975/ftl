package main

import "context"

type listCmd struct{}

func (l *listCmd) Run(ctx context.Context) error {
	return nil
}
