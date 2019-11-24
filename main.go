// This application lets you fetch the reviews from your apps on Google Play and show them on mattermost.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mattermost/mattermost-server/model"
	"google.golang.org/api/androidpublisher/v3"
)

var service *androidpublisher.Service

var packageList = []string{}
var aliases = make(map[string]string)

const (
	getListTime = 30 * time.Second
)

func main() {
	var err error
	if service, err = initService(); err != nil {
		fmt.Println("Error initializing the service:", err.Error())
	}

	http.HandleFunc("/list", serveList)
	http.HandleFunc("/listApps", serveAppList)
	http.HandleFunc("/listNewReviewsAlerts", serveListNewReviewsAlerts)
	http.HandleFunc("/setAlias", setAlias)
	http.HandleFunc("/addApp", addApp)
	http.HandleFunc("/addNewReviewsAlert", addNewReviewsAlert)
	http.HandleFunc("/removeNewReviewsAlert", removeNewReviewsAlert)
	go getAllReviews()
	http.ListenAndServe(":8080", nil)
}

func initService() (*androidpublisher.Service, error) {
	ctx := context.Background()
	service, err := androidpublisher.NewService(ctx)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}

	return service, nil
}

func removeNewReviewsAlert(w http.ResponseWriter, r *http.Request) {
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
	if _, ok := newReviewsAlerts[args[1]]; !ok {
		text += fmt.Sprintf("There no alert named %s.", args[1])
		return
	}

	delete(newReviewsAlerts, args[1])
	text += fmt.Sprintf("Alert %s removed.", args[1])
}

func serveListNewReviewsAlerts(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	for k, v := range newReviewsAlerts {
		text += fmt.Sprintf("Alert \"%s\": From package %s every %v seconds at most on webhook %s\n", k, v.packageName, v.frequency, v.webhook)
	}
}

func addNewReviewsAlert(w http.ResponseWriter, r *http.Request) {
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
		text += fmt.Sprintf("Wrong use: %s unique_name webhook packageName minimum_frequency_in_seconds", args[0])
		return
	}
	if _, ok := newReviewsAlerts[args[1]]; ok {
		text += fmt.Sprintf("There is already an alert named %s.", args[1])
		return
	}

	if !contains(packageList, args[3]) {
		text += fmt.Sprintf("Package %s is not yet registered.", args[3])
		return
	}

	frequency, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil || frequency <= 0 {
		text += fmt.Sprintf("%s is not a well formed frequency. Please use a positive number.", args[3])
		return
	}

	newReviewsAlerts[args[1]] = newReviewsAlert{
		webhook:     args[2],
		packageName: args[3],
		frequency:   frequency,
		lastAlerted: time.Now(),
	}

	text += fmt.Sprintf("Alert %s registered.", args[1])
}

func addApp(w http.ResponseWriter, r *http.Request) {
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

	if contains(packageList, args[1]) {
		text += fmt.Sprintf("Package %s already registered.", args[1])
		return
	}

	_, err := service.Reviews.List(args[1]).Do()
	if err != nil {
		text += fmt.Sprintf("Error registering the app %s: %v", args[1], err.Error())
		return
	}

	packageList = append(packageList, args[1])
	text += fmt.Sprintf("Package %s added to the system.", args[1])
}

func setAlias(w http.ResponseWriter, r *http.Request) {
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
	if !contains(packageList, args[2]) {
		fmt.Fprintf(w, "App %s not registered.", args[2])
		return
	}

	if _, ok := aliases[args[1]]; ok {
		text += fmt.Sprintf("Alias %s already set for app %s.", args[1], aliases[args[1]])
		return
	}

	aliases[args[1]] = args[2]
	text += fmt.Sprintf("Alias %s set for app %s.", args[1], args[2])
}

func serveAppList(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	text += "Here are all the apps you have registered:\n"
	for _, packageName := range packageList {
		text += packageName
		if al := getAliasesForPackage(packageName); len(al) > 0 {
			text += " AKA"
			for _, alias := range al {
				text += fmt.Sprintf(" %s", alias)
			}
		}
		text += "\n"
	}
}

func serveList(w http.ResponseWriter, r *http.Request) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	maxReviewsServed := 10
	text += fmt.Sprintf("Here are the %d latest reviews from each app:\n", maxReviewsServed)
	reviewsMutex.Lock()
	defer reviewsMutex.Unlock()
	for key, reviewList := range localReviews {
		text += fmt.Sprintf("Package Id: %s\n", key)
		for i, review := range reviewList {
			if i >= maxReviewsServed {
				break
			}
			text += formatReview(review)
		}
	}
}
