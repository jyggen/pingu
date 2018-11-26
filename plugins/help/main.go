package main

import (
	"fmt"
	"github.com/jyggen/pingu/pingu"
	"github.com/nlopes/slack"
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
			Description: "Lists all available commands.",
			Func: func(pi *pingu.Pingu, ev *slack.MessageEvent) {
				pi.Reply(ev, generateHelpOutput(pi))
			},
			Trigger: regexp.MustCompile("^!help$"),
		},
	}
}

func (pl *plugin) Name() string {
	return "Help"
}

func (pl *plugin) Tasks() pingu.Tasks {
	return pingu.Tasks{}
}

func (pl *plugin) Version() string {
	return version
}

func generateHelpOutput(pi *pingu.Pingu) string {
	output := "Here's a list of all available commands:\n\n```\n"

	for _, pl := range pi.Plugins() {
		output += fmt.Sprintf("%s (%s):\n", pl.Name(), pl.Version())

		for _, cmd := range pl.Commands() {
			trigger := cmd.Trigger.String()
			output += fmt.Sprintf("%s: %s\n", trigger[1:len(trigger)-1], cmd.Description)
		}

		output += "\n"
	}

	return output[:len(output)-1] + "```\n"
}
