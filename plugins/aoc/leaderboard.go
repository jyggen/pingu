package main

import (
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"time"
)

type leaderboardList []*leaderboard

type leaderboard struct {
	Year    int
	Members memberList
}

type memberList []*member

type member struct {
	Id          int
	GlobalScore int
	LastStarAt  time.Time
	LocalScore  int
	Name        string
	Position    int
	Stars       starList
	TotalStars  int
}

type starList []*star

type star struct {
	CompletedAt time.Time
	Day         int
	Part        int
	Year        int
}

func (l *leaderboard) Refresh(c *client) error {
	body, err := c.GetLeaderboard(l.Year)

	if err != nil {
		return errors.WithMessage(err, "")
	}

	members := make(memberList, len(body.Members))
	i := 0

	for _, m := range body.Members {
		memberId, _ := strconv.Atoi(m.Id)

		var lastStarTs int

		switch v := m.LastStarTs.(type) {
		case int:
			lastStarTs = v
		case string:
			lastStarTs, _ = strconv.Atoi(v)
		}

		lastStarAt := time.Unix(int64(lastStarTs), 0)
		stars := make(starList, 0)

		for day, parts := range m.CompletionDayLevel {
			for part, data := range parts {
				if _, ok := data["get_star_ts"]; !ok {
					continue
				}

				stars = append(stars, &star{
					CompletedAt: time.Unix(int64(data["get_star_ts"]), 0),
					Day:         day,
					Part:        part,
					Year:        l.Year,
				})
			}
		}

		members[i] = &member{
			Id:          memberId,
			GlobalScore: m.GlobalScore,
			LastStarAt:  lastStarAt,
			LocalScore:  m.LocalScore,
			Name:        m.Name,
			Stars:       stars,
			TotalStars:  m.Stars,
		}

		i++
	}

	l.Members = members

	l.Sort()

	return nil
}

func (l *leaderboard) Sort() {
	sort.Slice(l.Members, func(i, j int) bool {
		if l.Members[j].TotalStars == l.Members[i].TotalStars {
			if l.Members[j].LastStarAt.Equal(l.Members[i].LastStarAt) {
				if l.Members[j].Name < l.Members[i].Name {
					return false
				} else {
					return true
				}
			}

			if l.Members[j].LastStarAt.Before(l.Members[i].LastStarAt) {
				return false
			} else {
				return true
			}
		}

		if l.Members[j].TotalStars > l.Members[i].TotalStars {
			return false
		} else {
			return true
		}
	})

	position := 1
	skip := 0

	for i, m := range l.Members {
		if i != 0 && m.TotalStars < l.Members[i-1].TotalStars {
			position += skip
			skip = 0
		}

		m.Position = position
		skip++
	}
}

type ByDate starList

func (a ByDate) Len() int      { return len(a) }
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool {
	return a[i].Year < a[j].Year || (a[i].Year == a[j].Year && a[i].Day < a[j].Day) || (a[i].Year == a[j].Year && a[i].Day == a[j].Day && a[i].Part < a[j].Part)
}
