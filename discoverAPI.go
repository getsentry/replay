package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"
)

type DiscoverAPI struct {
	Data     []EventMetadata
	endpoint string
}

type EventMetadata struct {
	Id       string
	Project  string
	Platform string
}

// Events from last 24HrPeriod events for selected Projects
// Returns event metadata (e.g. Id, Project) but not the entire Event itself, which gets queried separately.
// n of 200 may not work
func (d DiscoverAPI) latestEventMetadata(org string, n int) []EventMetadata {
	fmt.Printf("\n> ORG %v\n", org)

	// query := "platform.name%3Ajavascript+OR+platform.name%3Apython"
	query := "platform.name%3Ajavascript+OR+platform.name%3Apython+OR+platform.name%3Ajava+OR+platform.name%3Aruby+OR+platform.name%3Ago+OR+platform.name%3Anode+OR+platform.name%3Aphp"

	// 0 project names specified
	endpoint := fmt.Sprintf("https://sentry.io/api/0/organizations/%v/eventsv2/?statsPeriod=24h&field=event.type&field=project&field=platform&per_page=%v&query=%v", org, strconv.Itoa(n), query)

	request, _ := http.NewRequest("GET", endpoint, nil)
	request.Header.Set("content-type", "application/json")
	request.Header.Set("Authorization", fmt.Sprint("Bearer ", os.Getenv("SENTRY_AUTH_TOKEN")))

	var httpClient = &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
	}
	body, errResponse := ioutil.ReadAll(response.Body)
	if errResponse != nil {
		sentry.CaptureException(errResponse)
		log.Fatal(errResponse)
	}

	json.Unmarshal(body, &d)
	fmt.Printf("> %v Discover.Data length: %v\n", org, len(d.Data))

	for _, e := range d.Data {
		fmt.Printf("> %v %v\n", org, e.Project)
	}
	return d.Data
}

// Consider chaining
// func (d DiscoverAPI) execute() {
// //
// }