package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// JSONTime wrapper around time.Time that allows for easy date conversion for marshalling
type JSONTime struct {
	time.Time
}

// MarshalJSON converts dates to UTC and in the ISO8601 format
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t.UTC().Format(time.RFC3339))
	return []byte(stamp), nil
}

type horaroMetaV1 struct {
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Timezone    string    `json:"timezone"`
	Start       time.Time `json:"start"`
	Website     string    `json:"website"`
	Twitter     string    `json:"twitter"`
	Twitch      string    `json:"twitch"`
	Description string    `json:"description"`
	Setup       string    `json:"setup"`
	Updated     time.Time `json:"updated"`
	URL         string    `json:"url"`
	Event       struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"event"`
	Exported time.Time `json:"exported"`
}

type horaroMetaV2 struct {
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Timezone    string   `json:"timezone"`
	Start       JSONTime `json:"start"`
	Website     string   `json:"website"`
	Twitter     string   `json:"twitter"`
	Twitch      string   `json:"twitch"`
	Description string   `json:"description"`
	Setup       string   `json:"setup"`
	Updated     JSONTime `json:"updated"`
	URL         string   `json:"url"`
	Event       struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"event"`
	Exported JSONTime `json:"exported"`
}

// TransformedHoraroResponseV1 is the modified response from horaro in list form
type TransformedHoraroResponseV1 struct {
	Meta horaroMetaV1  `json:"meta"`
	Data []eventDataV1 `json:"data"`
}

// TransformedHoraroResponseV2 is the modified response from horaro in list form
type TransformedHoraroResponseV2 struct {
	Meta horaroMetaV2  `json:"meta"`
	Data []eventDataV2 `json:"data"`
}

// ScheduleHoraroResponseV1 is the modified response from horaro spaced out in days
type ScheduleHoraroResponseV1 struct {
	Meta horaroMetaV1             `json:"meta"`
	Data map[string][]eventDataV1 `json:"data"`
}

// ScheduleHoraroResponseV2 is the modified response from horaro spaced out in days
type ScheduleHoraroResponseV2 struct {
	Meta horaroMetaV2             `json:"meta"`
	Data map[string][]eventDataV2 `json:"data"`
}

type eventDataV1 struct {
	Length    int         `json:"length"`
	Scheduled time.Time   `json:"scheduled"`
	Game      *string     `json:"game"`
	Players   []string    `json:"players"`
	Platform  *string     `json:"platform"`
	Category  *string     `json:"category"`
	Note      *string     `json:"note"`
	Layout    *string     `json:"layout"`
	Info      *string     `json:"info"`
	ID        *string     `json:"id"`
	Options   interface{} `json:"options"`
}

type eventDataV2 struct {
	Length    int         `json:"length"`
	Scheduled JSONTime    `json:"scheduled"`
	Game      *string     `json:"game"`
	Players   []string    `json:"players"`
	Platform  *string     `json:"platform"`
	Category  *string     `json:"category"`
	Note      *string     `json:"note"`
	Layout    *string     `json:"layout"`
	Info      *string     `json:"info"`
	ID        *string     `json:"id"`
	Options   interface{} `json:"options"`
}

// indexOf gets the index of an element in a list ignoring casing
func indexOf(element string, data []string, compareFunc func(s, t string) bool) int {
	for i, v := range data {
		if compareFunc(element, v) {
			return i
		}
	}

	return -1
}

// Matches the following:
// " vs. "
// " vs "
// ", "
// " , "
// " and "
// " & "
var playersPattern = regexp.MustCompile(`\s*(\svs.\s|\svs\s|\s*,\s|\sand\s|\s&\s)\s*`)

// OrganizeHoraro organizes the response from horaro into days
func OrganizeHoraro(list TransformedHoraroResponseV1) ScheduleHoraroResponseV1 {
	schedule := ScheduleHoraroResponseV1{}
	schedule.Meta = list.Meta

	schedule.Data = make(map[string][]eventDataV1)

	for _, value := range list.Data {
		// why, google, why would this be the format format
		key := value.Scheduled.Format("2006-01-02")

		if arr, ok := schedule.Data[key]; ok {
			schedule.Data[key] = append(arr, value)
		} else {
			schedule.Data[key] = []eventDataV1{value}
		}
	}

	return schedule
}

// UpcomingHoraroV1 gets the upcoming values from horaro
func UpcomingHoraroV1(list TransformedHoraroResponseV1, amount int) TransformedHoraroResponseV1 {
	upcoming := TransformedHoraroResponseV1{}
	upcoming.Meta = list.Meta

	upcoming.Data = []eventDataV1{}

	now := time.Now()

	for _, value := range list.Data {
		if len(upcoming.Data) >= amount {
			break
		}

		start := value.Scheduled
		end := value.Scheduled.Add(time.Second * time.Duration(value.Length))
		if start.After(now) || (start.Before(now) && end.After(now)) {
			upcoming.Data = append(upcoming.Data, value)
		}
	}

	return upcoming
}

// UpcomingHoraroV2 gets the upcoming values from horaro
func UpcomingHoraroV2(list TransformedHoraroResponseV2, amount int) TransformedHoraroResponseV2 {
	upcoming := TransformedHoraroResponseV2{}
	upcoming.Meta = list.Meta

	upcoming.Data = []eventDataV2{}

	now := time.Now()

	for _, value := range list.Data {
		if len(upcoming.Data) >= amount {
			break
		}

		start := value.Scheduled
		end := value.Scheduled.Add(time.Second * time.Duration(value.Length))
		if start.After(now) || (start.Before(now) && end.After(now)) {
			upcoming.Data = append(upcoming.Data, value)
		}
	}

	return upcoming
}

// TransformHoraroV1 transforms the response from the official horaro to a better format
func TransformHoraroV1(horaro *HoraroResponse) TransformedHoraroResponseV1 {
	response := TransformedHoraroResponseV1{}

	// Format response Meta
	response.Meta.Name = horaro.Schedule.Name
	response.Meta.Slug = horaro.Schedule.Slug
	response.Meta.Timezone = horaro.Schedule.Timezone
	response.Meta.Start = horaro.Schedule.Start
	response.Meta.Website = horaro.Schedule.Website
	response.Meta.Twitter = horaro.Schedule.Twitter
	response.Meta.Twitch = horaro.Schedule.Twitch
	response.Meta.Description = horaro.Schedule.Description
	response.Meta.Setup = horaro.Schedule.Setup
	response.Meta.Updated = horaro.Schedule.Updated
	response.Meta.URL = horaro.Schedule.URL
	response.Meta.Event = horaro.Schedule.Event
	response.Meta.Exported = horaro.Meta.Exported

	// Format response Data
	gameColumnIndex := indexOf("Game", horaro.Schedule.Columns, strings.EqualFold)
	playersColumnIndex := indexOf("Player(s)", horaro.Schedule.Columns, strings.EqualFold)
	platformColumnIndex := indexOf("Platform", horaro.Schedule.Columns, strings.EqualFold)
	categoryColumnIndex := indexOf("Category", horaro.Schedule.Columns, strings.EqualFold)
	noteColumnIndex := indexOf("Note", horaro.Schedule.Columns, strings.EqualFold)
	layoutColumnIndex := indexOf("Layout", horaro.Schedule.Columns, strings.EqualFold)
	infoColumIndex := indexOf("Info", horaro.Schedule.Columns, strings.EqualFold)
	idColumnIndex := indexOf("ID", horaro.Schedule.Columns, strings.EqualFold)

	eventList := make([]eventDataV1, len(horaro.Schedule.Items))

	for i, value := range horaro.Schedule.Items {
		eventList[i] = eventDataV1{}
		eventList[i].Length = value.LengthT
		eventList[i].Scheduled = value.Scheduled
		eventList[i].Options = value.Options

		if playersColumnIndex > -1 {
			eventList[i].Players = playersPattern.Split(*value.Data[playersColumnIndex], -1)
		}
		if gameColumnIndex > -1 {
			eventList[i].Game = value.Data[gameColumnIndex]
		}
		if platformColumnIndex > -1 {
			eventList[i].Platform = value.Data[platformColumnIndex]
		}
		if categoryColumnIndex > -1 {
			eventList[i].Category = value.Data[categoryColumnIndex]
		}
		if noteColumnIndex > -1 {
			eventList[i].Note = value.Data[noteColumnIndex]
		}
		if layoutColumnIndex > -1 {
			eventList[i].Layout = value.Data[layoutColumnIndex]
		}
		if infoColumIndex > -1 {
			eventList[i].Info = value.Data[infoColumIndex]
		}
		if idColumnIndex > -1 {
			eventList[i].ID = value.Data[idColumnIndex]
		}
	}

	response.Data = eventList

	return response
}

// TransformHoraroV2 transforms the response from the official horaro to a better format
func TransformHoraroV2(horaro *HoraroResponse) TransformedHoraroResponseV2 {
	response := TransformedHoraroResponseV2{}

	// Format response Meta
	response.Meta.Name = horaro.Schedule.Name
	response.Meta.Slug = horaro.Schedule.Slug
	response.Meta.Timezone = horaro.Schedule.Timezone
	response.Meta.Start = JSONTime{horaro.Schedule.Start}
	response.Meta.Website = horaro.Schedule.Website
	response.Meta.Twitter = horaro.Schedule.Twitter
	response.Meta.Twitch = horaro.Schedule.Twitch
	response.Meta.Description = horaro.Schedule.Description
	response.Meta.Setup = horaro.Schedule.Setup
	response.Meta.Updated = JSONTime{horaro.Schedule.Updated}
	response.Meta.URL = horaro.Schedule.URL
	response.Meta.Event = horaro.Schedule.Event
	response.Meta.Exported = JSONTime{horaro.Meta.Exported}

	// Format response Data
	gameColumnIndex := indexOf("Game", horaro.Schedule.Columns, strings.EqualFold)
	playersColumnIndex := indexOf("Player(s)", horaro.Schedule.Columns, strings.EqualFold)
	platformColumnIndex := indexOf("Platform", horaro.Schedule.Columns, strings.EqualFold)
	categoryColumnIndex := indexOf("Category", horaro.Schedule.Columns, strings.EqualFold)
	noteColumnIndex := indexOf("Note", horaro.Schedule.Columns, strings.EqualFold)
	layoutColumnIndex := indexOf("Layout", horaro.Schedule.Columns, strings.EqualFold)
	infoColumIndex := indexOf("Info", horaro.Schedule.Columns, strings.EqualFold)
	idColumnIndex := indexOf("ID", horaro.Schedule.Columns, strings.EqualFold)

	eventList := make([]eventDataV2, len(horaro.Schedule.Items))

	for i, value := range horaro.Schedule.Items {
		eventList[i] = eventDataV2{}
		eventList[i].Length = value.LengthT
		eventList[i].Scheduled = JSONTime{value.Scheduled}
		eventList[i].Options = value.Options

		if playersColumnIndex > -1 {
			eventList[i].Players = playersPattern.Split(*value.Data[playersColumnIndex], -1)
		}
		if gameColumnIndex > -1 {
			eventList[i].Game = value.Data[gameColumnIndex]
		}
		if platformColumnIndex > -1 {
			eventList[i].Platform = value.Data[platformColumnIndex]
		}
		if categoryColumnIndex > -1 {
			eventList[i].Category = value.Data[categoryColumnIndex]
		}
		if noteColumnIndex > -1 {
			eventList[i].Note = value.Data[noteColumnIndex]
		}
		if layoutColumnIndex > -1 {
			eventList[i].Layout = value.Data[layoutColumnIndex]
		}
		if infoColumIndex > -1 {
			eventList[i].Info = value.Data[infoColumIndex]
		}
		if idColumnIndex > -1 {
			eventList[i].ID = value.Data[idColumnIndex]
		}
	}

	response.Data = eventList

	return response
}
