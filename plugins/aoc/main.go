package main

import (
	"fmt"
	"github.com/jyggen/pingu/pingu"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type plugin struct {
	channel      string
	client       *client
	global       *leaderboard
	leaderboards leaderboardList
}

var leaderboardRegex *regexp.Regexp
var version string

func init() {
	leaderboardRegex = regexp.MustCompile("^!leaderboard(?: ([\\d]{4}))?$")
}

func New(c *viper.Viper) pingu.Plugin {
	return pingu.Plugin(&plugin{
		channel: c.GetString("aoc.channel"),
		client: &client{
			httpClient: &http.Client{
				Timeout: c.GetDuration("aoc.timeout") * time.Second,
			},
			ownerId: c.GetInt("aoc.owner"),
			session: c.GetString("aoc.session"),
		},
		global: &leaderboard{},
	})
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
			Description: "Prints either the global leaderboard, or the leaderboard for a specific year.",
			Func:        pl.postLeaderboard,
			Trigger:     leaderboardRegex,
		},
		&pingu.Command{
			Description: "Forces a refresh of all leaderboards.",
			Func: func(pi *pingu.Pingu, ev *slack.MessageEvent) {
				if ev.Channel != pl.channel {
					pi.Reply(ev, fmt.Sprintf("Noot! Noot! That command is only available in <#%s>!", pl.channel))
					return
				}

				pl.refreshLeaderboards(pi)
			},
			Trigger: regexp.MustCompile("^!refresh$"),
		},
	}
}

func (pl *plugin) Name() string {
	return "Advent of Code"
}

func (pl *plugin) Tasks() pingu.Tasks {
	return pingu.Tasks{
		&pingu.Task{
			Func:     pl.refreshLeaderboards,
			Interval: time.Minute * 15,
		},
		&pingu.Task{
			Func: pl.announceNewDay,
			Spec: "0 5 1-25 DEC *",
		},
	}
}

func (pl *plugin) Version() string {
	return version
}

func (pl *plugin) announceJoined(pi *pingu.Pingu, a leaderboard, b leaderboard) {
	availableStars := calculateAvailableStars(time.Now(), b.Year)

Loop:
	for _, our := range b.Members {
		for _, their := range a.Members {
			if our.Id == their.Id {
				continue Loop
			}
		}

		var starsLabel string

		if our.TotalStars != 1 {
			starsLabel = "stars"
		} else {
			starsLabel = "star"
		}

		pi.Say(fmt.Sprintf(
			"Noot! Noot! _%s_ has joined our ranks, starting at position *%d* with *%d %s* (%.2f%%) collected.",
			our.Name,
			our.Position,
			our.TotalStars,
			starsLabel,
			float64(our.TotalStars)/float64(availableStars)*100,
		), pl.channel)
	}
}

func (pl *plugin) announceLeft(pi *pingu.Pingu, a leaderboard, b leaderboard) {
Loop:
	for _, their := range a.Members {
		for _, our := range b.Members {
			if their.Id == our.Id {
				continue Loop
			}
		}

		pi.Say(fmt.Sprintf(
			"Noot! Noot! It seems like _%s_ has left our ranks. The leaderboards have been recalculated.",
			their.Name,
		), pl.channel)
	}
}

func (pl *plugin) announceNewDay(pi *pingu.Pingu) {
	year, _, day := time.Now().Date()

	pi.Say(fmt.Sprintf(
		"Noot! Noot! <https://adventofcode.com/%[1]d/day/%[2]d|Day %[2]d of %[1]d is now available!> Please keep spoilers to a minimum and instead use a thread on this very message to discuss today's challenge. Happy coding!",
		year,
		day,
	), pl.channel)
}

func (pl *plugin) announceChanges(pi *pingu.Pingu, a leaderboard, b leaderboard) {
	availableStars := calculateAvailableStars(time.Now(), b.Year)

OurLoop:
	for _, our := range b.Members {
		for _, their := range a.Members {
			if our.Id == their.Id {
				if our.TotalStars != their.TotalStars {
					diff := calculateDifference(our.Stars, their.Stars)
					diffLen := len(diff)

					if diffLen != 0 {
						starsMessage := pl.buildChangeMessage(diff)

						var starsLabel string

						if diffLen != 1 {
							starsLabel = "stars"
						} else {
							starsLabel = "star"
						}

						message := fmt.Sprintf("_%s_ earned *%d %s* by completing _%s_", our.Name, diffLen, starsLabel, starsMessage)

						if their.Position > our.Position {
							message += fmt.Sprintf(", moving up to *position %d*", our.Position)
						} else if their.Position < our.Position {
							message += fmt.Sprintf(", moving down to *position %d*", our.Position)
						} else {
							message += fmt.Sprintf(", staying at *position %d*", our.Position)
						}

						if our.TotalStars != 1 {
							starsLabel = "stars"
						} else {
							starsLabel = "star"
						}

						pi.Say(message+fmt.Sprintf(
							" with *%d %s* (%.2f%%)!",
							our.TotalStars,
							starsLabel,
							float64(our.TotalStars)/float64(availableStars)*100,
						), pl.channel)
					}
				}

				continue OurLoop
			}
		}
	}
}

func (pl *plugin) buildChangeMessage(l starList) string {
	numOfStars := len(l)

	if numOfStars == 0 {
		return ""
	}

	stars := make([]string, numOfStars)

	sort.Sort(ByDate(l))

	for i, s := range l {
		stars[i] = fmt.Sprintf("Day %d Part %d (%d)", s.Day, s.Part, s.Year)
	}

	starsString := strings.Join([]string{
		strings.Join(stars[:numOfStars-1], "_, _"),
		stars[numOfStars-1],
	}, "_ and _")

	if starsString[0:7] == "_ and _" {
		starsString = starsString[7:]
	}

	return starsString
}

func (pl *plugin) buildLeaderboard(l *leaderboard) string {
	availableStars := calculateAvailableStars(time.Now(), l.Year)
	message := ""

	for _, m := range l.Members {
		var starsLabel string

		if m.TotalStars != 1 {
			starsLabel = "stars"
		} else {
			starsLabel = "star"
		}

		message += fmt.Sprintf(
			"*%d.* _%s_ with *%d %s* (%.2f%%) collected.\n",
			m.Position,
			m.Name,
			m.TotalStars,
			starsLabel,
			float64(m.TotalStars)/float64(availableStars)*100,
		)
	}

	return message
}

func (pl *plugin) postLeaderboard(pi *pingu.Pingu, ev *slack.MessageEvent) {
	if ev.Channel != pl.channel {
		pi.Reply(ev, fmt.Sprintf("Noot! Noot! That command is only available in <#%s>!", pl.channel))
		return
	}

	var board *leaderboard

	match := leaderboardRegex.FindStringSubmatch(ev.Text)

	if match[1] != "" {
		year, _ := strconv.Atoi(match[1])

		for _, l := range pl.leaderboards {
			if l.Year == year {
				board = l
			}
		}

		if board == nil {
			pi.Reply(ev, fmt.Sprintf("Noot! Noot! %d does not have a leaderboard!", year))
			return
		}
	} else {
		board = pl.global
	}

	pi.Say(pl.buildLeaderboard(board), ev.Channel)
}

func (pl *plugin) refreshGlobalLeaderboard() {
	members := make(map[int]*member)

	for _, l := range pl.leaderboards {
		for _, m := range l.Members {
			if _, ok := members[m.Id]; !ok {
				members[m.Id] = &member{
					Id:   m.Id,
					Name: m.Name,
				}
			}

			members[m.Id].GlobalScore += m.GlobalScore
			members[m.Id].LocalScore += m.LocalScore
			members[m.Id].Stars = append(members[m.Id].Stars, m.Stars...)
			members[m.Id].TotalStars += m.TotalStars

			if members[m.Id].LastStarAt.Before(m.LastStarAt) {
				members[m.Id].LastStarAt = m.LastStarAt
			}
		}
	}

	pl.global.Members = make(memberList, len(members))
	i := 0

	for _, m := range members {
		pl.global.Members[i] = m
		i++
	}

	pl.global.Sort()
}

func (pl *plugin) refreshLeaderboards(pi *pingu.Pingu) {
Loop:
	for _, year := range getValidYears(time.Now()) {
		for _, l := range pl.leaderboards {
			if l.Year == year {
				continue Loop
			}
		}

		pl.leaderboards = append(pl.leaderboards, &leaderboard{
			Year: year,
		})
	}

	var wg sync.WaitGroup

	wg.Add(len(pl.leaderboards))

	for _, l := range pl.leaderboards {
		go func(l *leaderboard) {
			before := *l
			err := l.Refresh(pl.client)
			after := *l

			if err != nil {
				pi.Logger().Error(err)
			} else if len(before.Members) != 0 {
				pl.announceChanges(pi, before, after)
			}

			wg.Done()
		}(l)
	}

	wg.Wait()
	before := *pl.global
	pl.refreshGlobalLeaderboard()
	after := *pl.global

	if len(before.Members) != 0 {
		pl.announceJoined(pi, before, after)
		pl.announceLeft(pi, before, after)
	}
}
