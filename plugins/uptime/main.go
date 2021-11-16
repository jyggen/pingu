package main

import (
	"fmt"
	"github.com/hako/durafmt"
	"github.com/jyggen/pingu/pingu"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
	"regexp"
	"time"
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
			Description: "Reports my current uptime.",
			Func: func(pi *pingu.Pingu, ev *slack.MessageEvent) {
				pi.Reply(ev, fmt.Sprintf(
					"My current uptime is %s, and I've been connected for %s.",
					durafmt.ParseShort(time.Since(pi.StartedAt())),
					durafmt.ParseShort(time.Since(pi.ConnectedAt())),
				))
			},
			Trigger: regexp.MustCompile("^!uptime$"),
		},
	}
}

func (pl *plugin) Name() string {
	return "Uptime"
}

func (pl *plugin) Tasks() pingu.Tasks {
	return pingu.Tasks{}
}

func (pl *plugin) Version() string {
	return version
}
