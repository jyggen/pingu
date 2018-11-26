package pingu

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"regexp"
	"time"
)

type Command struct {
	Description string
	Func        func(pi *Pingu, ev *slack.MessageEvent)
	Trigger     *regexp.Regexp
}

type Commands []*Command

type Pingu struct {
	builtAt     time.Time
	connectedAt time.Time
	config      *viper.Viper
	latency time.Duration
	logger      *logrus.Logger
	name        string
	plugins     Plugins
	rtm         *slack.RTM
	startedAt   time.Time
	version     string
}

type Task struct {
	Func        func(pi *Pingu)
	Interval    time.Duration
}

type Tasks []*Task

func New(config *viper.Viper, logger *logrus.Logger) *Pingu {
	factories, err := LoadPlugins(config.GetString("pingu.plugin_path"))

	if err != nil {
		logger.Fatal(err)
	}

	plugins := make(Plugins, len(factories))

	for i, factory := range factories {
		plugins[i] = factory(config)

		logger.WithFields(logrus.Fields{
			"author":  plugins[i].Author(),
			"name":    plugins[i].Name(),
			"version": plugins[i].Version(),
		}).Info("Plugin loaded")
	}

	api := slack.New(config.GetString("slack.token"))
	rtm := api.NewRTM()

	return &Pingu{
		config:    config,
		logger:    logger,
		name:      "Pingu",
		plugins:   plugins,
		rtm:       rtm,
		startedAt: time.Now(),
	}
}

func (p *Pingu) BuiltAt() time.Time {
	return p.builtAt
}

func (p *Pingu) ConnectedAt() time.Time {
	return p.connectedAt
}

func (p *Pingu) Latency() time.Duration {
	return p.latency
}

func (p *Pingu) Logger() *logrus.Logger {
	return p.logger
}

func (p *Pingu) Name() string {
	return p.name
}

func (p *Pingu) Plugins() Plugins {
	return p.plugins
}

func (p *Pingu) Reply(ev *slack.MessageEvent, msg string) {
	p.Say(fmt.Sprintf("<@%s>: %s", ev.User, msg), ev.Channel)
}

func (p *Pingu) Run() {
	c := cron.New()
	w := p.logger.Writer()

	defer w.Close()

	c.ErrorLog = log.New(w, "", 0)

	for _, plugin := range p.plugins {
		plugin := plugin
		for _, task := range plugin.Tasks() {
			task := task
			if err := c.AddFunc(fmt.Sprintf("@every %s", task.Interval.String()), func () {
				task.Func(p)
				p.logger.WithFields(logrus.Fields{
					"plugin":  plugin.Name(),
				}).Info("Task executed")
			}); err != nil {
				p.logger.Fatal(err)
			}
		}
	}

	go p.rtm.ManageConnection()

	for msg := range p.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			p.connectedAt = time.Now()
			p.logger.Info("Connection established")

			for _, plugin := range p.plugins {
				plugin := plugin
				for _, task := range plugin.Tasks() {
					task := task
					task.Func(p)
					p.logger.WithFields(logrus.Fields{
						"plugin":  plugin.Name(),
					}).Info("Task executed")
				}
			}

			c.Start()
			for _, e := range c.Entries() {
				fmt.Printf("%+v\n", e)
			}
		case *slack.DisconnectedEvent:
			p.logger.Info("Connection lost")
			c.Stop()
		case *slack.LatencyReport:
			p.latency = ev.Value
		case *slack.InvalidAuthEvent:
			p.logger.Fatal("Authentication failed")
		case *slack.MessageEvent:
			for _, plugin := range p.plugins {
				plugin := plugin
				for _, command := range plugin.Commands() {
					command := command
					if command.Trigger.MatchString(ev.Text) {
						p.logger.WithFields(logrus.Fields{
							"plugin":  plugin.Name(),
							"trigger": command.Trigger.String(),
						}).Info("Command triggered")
						command.Func(p, ev)
					}
				}
			}
		}
	}
}

func (p *Pingu) Say(msg string, ch string) {
	p.rtm.SendMessage(p.rtm.NewOutgoingMessage(msg, ch))
}

func (p *Pingu) SendAttachments(attachments []slack.Attachment, msg string, ch string) {
	_, _, err := p.rtm.PostMessage(ch, "", slack.PostMessageParameters{
		AsUser:      true,
		LinkNames:   1,
		UnfurlLinks: false,
		UnfurlMedia: true,
		Attachments: attachments,
	})

	if err != nil {
		p.logger.Error(err)
	}
}

func (p *Pingu) StartedAt() time.Time {
	return p.startedAt
}

func (p *Pingu) Version() string {
	return p.version
}
