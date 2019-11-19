package main

import (
	"regexp"
	"time"
)

// TransformedHoraroResponse is the modified response from horaro to be more accessible
type TransformedHoraroResponse struct {
	Meta struct {
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
	} `json:"meta"`
	Data []eventData `json:"data"`
}

type eventData struct {
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

// TransformHoraro transforms the response from the official horaro to a better format
func TransformHoraro(horaro *HoraroResponse) TransformedHoraroResponse {
	response := TransformedHoraroResponse{}

	// Meta

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

	// Data

	gameColumnIndex := IndexOfCaseInsensitive("Game", horaro.Schedule.Columns)
	playersColumnIndex := IndexOfCaseInsensitive("Player(s)", horaro.Schedule.Columns)
	platformColumnIndex := IndexOfCaseInsensitive("Platform", horaro.Schedule.Columns)
	categoryColumnIndex := IndexOfCaseInsensitive("Category", horaro.Schedule.Columns)
	noteColumnIndex := IndexOfCaseInsensitive("Note", horaro.Schedule.Columns)
	layoutColumnIndex := IndexOfCaseInsensitive("Layout", horaro.Schedule.Columns)
	infoColumIndex := IndexOfCaseInsensitive("Info", horaro.Schedule.Columns)
	idColumnIndex := IndexOfCaseInsensitive("ID", horaro.Schedule.Columns)

	eventList := make([]eventData, len(horaro.Schedule.Items))

	for i, value := range horaro.Schedule.Items {
		eventList[i] = eventData{}
		eventList[i].Length = value.LengthT
		eventList[i].Scheduled = value.Scheduled
		eventList[i].Options = value.Options

		if playersColumnIndex > -1 {
			// Split on:
			// " vs. "
			// " vs "
			// ", "
			// " , "
			// " and "
			// " & "
			eventList[i].Players = regexp.MustCompile("\\s*(\\svs.\\s|\\svs\\s|\\s*,\\s|\\sand\\s|\\s&\\s)\\s*").Split(*value.Data[playersColumnIndex], -1)
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
