package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type apiResponse struct {
	Event   string                 `json:"event"`
	Members map[int]memberResponse `json:"members"`
	OwnerId int                    `json:"owner_id"`
}

type memberResponse struct {
	Id                 int                            `json:"id"`
	GlobalScore        int                            `json:"global_score"`
	LastStarTs         interface{}                    `json:"last_star_ts"`
	LocalScore         int                            `json:"local_score"`
	Name               string                         `json:"name"`
	CompletionDayLevel map[int]map[int]map[string]int `json:"completion_day_level"`
	Stars              int                            `json:"stars"`
}

type client struct {
	httpClient *http.Client
	ownerId    int
	session    string
}

func (c *client) GetLeaderboard(year int) (apiResponse, error) {
	var jsonData apiResponse

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://adventofcode.com/%d/leaderboard/private/view/%d.json", year, c.ownerId), nil)

	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: c.session,
	})

	res, err := c.httpClient.Do(req)

	if err != nil {
		return jsonData, errors.WithMessage(err, "http request failed")
	}

	if res.Header.Get("Content-Type") != "application/json" {
		return jsonData, errors.New("authenticated failed")
	}

	body, err := ioutil.ReadAll(res.Body)

	defer res.Body.Close()

	if err != nil {
		return jsonData, errors.WithMessage(err, "unable to read response body")
	}

	err = json.Unmarshal(body, &jsonData)

	if err != nil {
		return jsonData, errors.WithMessage(err, "unable to unmarshal json")
	}

	return jsonData, nil
}
