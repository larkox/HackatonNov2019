package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mattermost/mattermost-server/model"
)

func (s *server) removeNewReviewsAlert(w http.ResponseWriter, r *http.Request, args []string) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	alertName := args[4]

	if len(args) != 5 {
		text += fmt.Sprintf("Wrong use: %s %s %s %s unique_name.", args[0], args[1], args[2], args[3])
		return
	}
	if _, ok := s.alerts.NewReviewsAlerts[alertName]; !ok {
		text += fmt.Sprintf("There no alert named %s.", alertName)
		return
	}

	delete(s.alerts.NewReviewsAlerts, alertName)
	s.SaveAlerts()
	text += fmt.Sprintf("Alert %s removed.", alertName)
}

func (s *server) serveListNewReviewsAlerts(w http.ResponseWriter, r *http.Request, args []string) {
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

func (s *server) addNewReviewsAlert(w http.ResponseWriter, r *http.Request, args []string) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	if len(args) != 8 {
		text += fmt.Sprintf("Wrong use: %s %s %s %s unique_name webhook packageName_or_alias minimum_frequency_in_seconds", args[0], args[1], args[2], args[3])
		return
	}
	uniqueName := args[4]
	webhook := args[5]
	packageNameOrAlias := args[6]
	minimumFrequency := args[7]

	if _, ok := s.alerts.NewReviewsAlerts[uniqueName]; ok {
		text += fmt.Sprintf("There is already an alert named %s.", uniqueName)
		return
	}

	packageName, ok := getPackageNameFromArgs(packageNameOrAlias, s.packageList, s.aliases)
	if !ok {
		text += fmt.Sprintf("Package %s is not yet registered.", packageNameOrAlias)
		return
	}

	frequency, err := strconv.ParseInt(minimumFrequency, 10, 64)
	if err != nil || frequency <= 0 {
		text += fmt.Sprintf("%s is not a well formed frequency. Please use a positive number.", minimumFrequency)
		return
	}

	s.alerts.NewReviewsAlerts[uniqueName] = NewReviewsAlert{
		Webhook:     webhook,
		PackageName: packageName,
		Frequency:   frequency,
		LastAlerted: time.Now(),
	}
	s.SaveAlerts()

	text += fmt.Sprintf("Alert %s registered.", uniqueName)
}

func (s *server) addApp(w http.ResponseWriter, r *http.Request, args []string) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	if len(args) != 4 {
		text += fmt.Sprintf("Wrong use: %s %s %s packageName", args[0], args[1], args[2])
		return
	}

	packageName := args[3]

	if contains(s.packageList, packageName) {
		text += fmt.Sprintf("Package %s already registered.", packageName)
		return
	}

	_, err := s.service.Reviews.List(packageName).Do()
	if err != nil {
		text += fmt.Sprintf("Error registering the app %s: %v", packageName, err.Error())
		return
	}

	s.packageList = append(s.packageList, packageName)
	s.SavePackages()
	text += fmt.Sprintf("Package %s added to the system.", packageName)
}

func (s *server) setAlias(w http.ResponseWriter, r *http.Request, args []string) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	if len(args) != 5 {
		text += fmt.Sprintf("Wrong use: %s %s %s aliasName packageName", args[0], args[1], args[2])
		return
	}

	aliasName := args[3]
	packageName := args[4]

	if !contains(s.packageList, packageName) {
		fmt.Fprintf(w, "App %s not registered.", packageName)
		return
	}

	if _, ok := s.aliases[aliasName]; ok {
		text += fmt.Sprintf("Alias %s already set for app %s.", aliasName, s.aliases[aliasName])
		return
	}

	s.aliases[aliasName] = packageName
	s.SaveAliases()
	text += fmt.Sprintf("Alias %s set for app %s.", aliasName, packageName)
}

func (s *server) serveAppList(w http.ResponseWriter, r *http.Request, args []string) {
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

func (s *server) serveList(w http.ResponseWriter, r *http.Request, args []string) {
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
