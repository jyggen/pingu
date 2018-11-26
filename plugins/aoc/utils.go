package main

import (
	"time"
)

const firstYear = 2015

func calculateAvailableStars(now time.Time, onlyYear int) int {
	var years []int

	if onlyYear != 0 {
		years = []int{onlyYear}
	} else {
		years = getValidYears(now)
	}

	stars := 0

	for _, year := range years {
		if year < now.Year() {
			stars += 50
			continue
		}

		if now.Month() < time.December {
			continue
		}

		stars += now.Day() * 2

		dayUnlockedAt := time.Date(now.Year(), now.Month(), now.Day(), 5, 0, 0, 0, time.UTC)

		if dayUnlockedAt.After(now) {
			stars -= 2
		}
	}

	return stars
}

func calculateDifference(a starList, b starList) starList {
	var diff starList

	for _, s1 := range a {
		found := false

		for _, s2 := range b {
			if s1.Year == s2.Year && s1.Day == s2.Day && s1.Part == s2.Part {
				found = true
				break
			}
		}

		if !found {
			diff = append(diff, s1)
		}
	}

	return diff
}

func getValidYears(now time.Time) []int {
	max := now.Year()
	startDate := time.Date(max, time.December, 1, 5, 0, 0, 0, time.UTC)

	if now.Before(startDate) {
		max--
	}

	if max < firstYear {
		return []int{}
	}

	years := make([]int, max-firstYear+1)

	for i := range years {
		years[i] = firstYear + i
	}

	return years
}
