package main

import (
	"reflect"
	"testing"
	"time"
)

func TestCalculateAvailableStars(t *testing.T) {
	newYork, err := time.LoadLocation("America/New_York")

	if err != nil {
		t.Error(err)
	}

	testCases := []struct {
		now      time.Time
		only     int
		expected int
	}{
		{time.Date(2009, time.October, 25, 0, 0, 0, 0, newYork), 0, 0},
		{time.Date(2015, time.October, 25, 0, 0, 0, 0, newYork), 0, 0},
		{time.Date(2015, time.December, 1, 0, 0, 0, 0, time.UTC), 0, 0},
		{time.Date(2015, time.December, 1, 0, 0, 0, 0, newYork), 0, 2},
		{time.Date(2016, time.October, 25, 0, 0, 0, 0, newYork), 0, 50},
		{time.Date(2018, time.December, 1, 0, 0, 0, 0, time.UTC), 0, 150},
		{time.Date(2018, time.December, 1, 0, 0, 0, 0, newYork), 0, 152},
		{time.Date(2018, time.December, 10, 0, 0, 0, 0, time.UTC), 0, 168},
		{time.Date(2018, time.December, 10, 0, 0, 0, 0, newYork), 0, 170},
		{time.Date(2018, time.December, 25, 0, 0, 0, 0, newYork), 0, 200},
		{time.Date(2020, time.December, 26, 0, 0, 0, 0, newYork), 0, 300},
		{time.Date(2020, time.December, 10, 0, 0, 0, 0, newYork), 2020, 20},
		{time.Date(2020, time.December, 26, 0, 0, 0, 0, newYork), 2020, 50},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.now.String(), func(t *testing.T) {
			t.Parallel()

			if actual := calculateAvailableStars(testCase.now, testCase.only); testCase.expected != actual {
				t.Errorf("calculateAvailableStars() was incorrect, got: %v, want %v.", actual, testCase.expected)
			}
		})
	}
}

func TestCalculateDifference(t *testing.T) {
	star := &star{
		CompletedAt: time.Now(),
		Day:         10,
		Part:        2,
		Year:        2017,
	}

	testCases := []struct {
		a        starList
		b        starList
		expected starList
	}{
		{
			a:        starList{star},
			b:        starList{},
			expected: starList{star},
		},
		{
			a:        starList{},
			b:        starList{star},
			expected: starList{},
		},
		{
			a:        starList{star},
			b:        starList{star},
			expected: starList{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run("", func(t *testing.T) {
			t.Parallel()

			if actual := calculateDifference(testCase.a, testCase.b); !reflect.DeepEqual(testCase.expected, actual) {
				t.Errorf("calculateDifference() was incorrect, got: %+v, want %+v.", actual, testCase.expected)
			}
		})
	}
}

func TestGetValidYears(t *testing.T) {
	newYork, err := time.LoadLocation("America/New_York")

	if err != nil {
		t.Error(err)
	}

	testCases := []struct {
		now      time.Time
		expected []int
	}{
		{time.Date(2009, time.October, 25, 0, 0, 0, 0, newYork), []int{}},
		{time.Date(2015, time.October, 25, 0, 0, 0, 0, newYork), []int{}},
		{time.Date(2015, time.December, 1, 0, 0, 0, 0, time.UTC), []int{}},
		{time.Date(2015, time.December, 1, 0, 0, 0, 0, newYork), []int{2015}},
		{time.Date(2016, time.October, 25, 0, 0, 0, 0, newYork), []int{2015}},
		{time.Date(2018, time.December, 1, 0, 0, 0, 0, time.UTC), []int{2015, 2016, 2017}},
		{time.Date(2018, time.December, 1, 0, 0, 0, 0, newYork), []int{2015, 2016, 2017, 2018}},
		{time.Date(2018, time.December, 25, 0, 0, 0, 0, newYork), []int{2015, 2016, 2017, 2018}},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.now.String(), func(t *testing.T) {
			t.Parallel()

			if actual := getValidYears(testCase.now); !reflect.DeepEqual(testCase.expected, actual) {
				t.Errorf("getValidYears() was incorrect, got: %v, want %v.", actual, testCase.expected)
			}
		})
	}
}
