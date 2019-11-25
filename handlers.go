package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mattermost/mattermost-server/model"
)

func (s *server) removeNewReviewsAlert(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	args, errMessage := getArgs(r)
	if args == nil {
		text += errMessage
		return
	}
	if len(args) != 2 {
		text += fmt.Sprintf("Wrong use: %s unique_name.", args[0])
		return
	}
	if _, ok := s.alerts.NewReviewsAlerts[args[1]]; !ok {
		text += fmt.Sprintf("There no alert named %s.", args[1])
		return
	}

	delete(s.alerts.NewReviewsAlerts, args[1])
	s.SaveAlerts()
	text += fmt.Sprintf("Alert %s removed.", args[1])
}

func (s *server) serveListNewReviewsAlerts(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	for k, v := range s.alerts.NewReviewsAlerts {
		text += fmt.Sprintf("Alert \"%s\": From package %s every %v seconds at most on webhook %s\n", k, v.PackageName, v.Frequency, v.Webhook)
	}
}

func (s *server) addNewReviewsAlert(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	args, errMessage := getArgs(r)
	if args == nil {
		text += errMessage
		return
	}

	if len(args) != 5 {
		text += fmt.Sprintf("Wrong use: %s unique_name webhook packageName_or_alias minimum_frequency_in_seconds", args[0])
		return
	}
	if _, ok := s.alerts.NewReviewsAlerts[args[1]]; ok {
		text += fmt.Sprintf("There is already an alert named %s.", args[1])
		return
	}

	packageName, ok := getPackageNameFromArgs(args[3], s.packageList, s.aliases)
	if !ok {
		text += fmt.Sprintf("Package %s is not yet registered.", args[3])
		return
	}

	frequency, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil || frequency <= 0 {
		text += fmt.Sprintf("%s is not a well formed frequency. Please use a positive number.", args[4])
		return
	}

	s.alerts.NewReviewsAlerts[args[1]] = NewReviewsAlert{
		Webhook:     args[2],
		PackageName: packageName,
		Frequency:   frequency,
		LastAlerted: time.Now(),
	}
	s.SaveAlerts()

	text += fmt.Sprintf("Alert %s registered.", args[1])
}

func (s *server) addApp(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	args, errMessage := getArgs(r)
	if args == nil {
		text += errMessage
		return
	}
	if len(args) != 2 {
		text += fmt.Sprintf("Wrong use: %s packageName", args[0])
		return
	}

	if contains(s.packageList, args[1]) {
		text += fmt.Sprintf("Package %s already registered.", args[1])
		return
	}

	_, err := s.service.Reviews.List(args[1]).Do()
	if err != nil {
		text += fmt.Sprintf("Error registering the app %s: %v", args[1], err.Error())
		return
	}

	s.packageList = append(s.packageList, args[1])
	s.SavePackages()
	text += fmt.Sprintf("Package %s added to the system.", args[1])
}

func (s *server) setAlias(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	args, errMessage := getArgs(r)
	if args == nil {
		text += errMessage
		return
	}

	if len(args) != 3 {
		text += fmt.Sprintf("Wrong use: %s alias packageName", args[0])
		return
	}
	if !contains(s.packageList, args[2]) {
		fmt.Fprintf(w, "App %s not registered.", args[2])
		return
	}

	if _, ok := s.aliases[args[1]]; ok {
		text += fmt.Sprintf("Alias %s already set for app %s.", args[1], s.aliases[args[1]])
		return
	}

	s.aliases[args[1]] = args[2]
	s.SaveAliases()
	text += fmt.Sprintf("Alias %s set for app %s.", args[1], args[2])
}

func (s *server) serveAppList(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	text += "Here are all the apps you have registered:\n"
	for _, packageName := range s.packageList {
		text += packageName
		if al := getAliasesForPackage(packageName, s.aliases); len(al) > 0 {
			text += " AKA"
			for _, alias := range al {
				text += fmt.Sprintf(" %s", alias)
			}
		}
		text += "\n"
	}
}

func (s *server) serveList(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	text += fmt.Sprintf("Here are the %d latest reviews from each app:\n", s.config.MaxReviewsServed)
	s.control.reviewsMutex.Lock()
	defer s.control.reviewsMutex.Unlock()
	for key, reviewList := range s.localReviews {
		text += fmt.Sprintf("Package Id: %s\n", key)
		for i, review := range reviewList {
			if i >= s.config.MaxReviewsServed {
				break
			}
			text += formatReview(review)
		}
	}
}
