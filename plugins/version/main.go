package main

import (
	"fmt"
	"github.com/jyggen/pingu/pingu"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
	"regexp"
)

type plugin struct{}

var version string

func New(c *viper.Viper) pingu.Plugin {
	return pingu.Plugin(&plugin{})
}

func (pl *plugin) Author() pingu.Author {
	return pingu.Author{
		Email: "jonas@stendahl.me",
		Name:  "Jonas Stendahl",
	}
}

func (pl *plugin) Commands() pingu.Commands {
	return pingu.Commands{
		&pingu.Command{
			Description: "Reports the version of myself I'm currently running.",
			Func: func(pi *pingu.Pingu, ev *slack.MessageEvent) {
				pi.Reply(ev, fmt.Sprintf("I'm currently running Pingu %s, built at %s.", pi.FriendlyVersion(), pi.BuiltAt()))
			},
			Trigger: regexp.MustCompile("^!version$"),
		},
	}
}

func (pl *plugin) Name() string {
	return "Version"
}

func (pl *plugin) Tasks() pingu.Tasks {
	return pingu.Tasks{}
}

func (pl *plugin) Version() string {
	return version
}
