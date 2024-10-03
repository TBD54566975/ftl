package main

import (
	"fmt"
)

type releaseCmd struct {
	Describe releaseDescribeCmd `cmd:"" help:"Describes the specified release."`
	Publish  releasePublishCmd  `cmd:"" help:"Packages the project into a release and publishes it."`
	List     releaseListCmd     `cmd:"" help:"Lists all published releases."`
}

type releaseDescribeCmd struct {
}

func (d *releaseDescribeCmd) Run() error {
	return fmt.Errorf("release describe not implemented")
}

type releasePublishCmd struct {
}

func (d *releasePublishCmd) Run() error {
	return fmt.Errorf("release publish not implemented")
}

type releaseListCmd struct {
}

func (d *releaseListCmd) Run() error {
	return fmt.Errorf("release list not implemented")
}
