// This application lets you fetch the reviews from your apps on Google Play and show them on mattermost.
package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/api/androidpublisher/v3"
)

var service *androidpublisher.Service

var packageList = []string{}
var aliases = make(map[string]string)

// OutgoingWebhookJSON Schema of the payload of Outgoing Webhooks received from Mattermost
type OutgoingWebhookJSON = struct {
	Token       string `json:"token"`
	TeamID      string `json:"team_id"`
	TeamDomain  string `json:"team_domain"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Timestamp   int64  `json:"timestamp"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	PostID      string `json:"post_id"`
	Text        string `json:"text"`
	TriggerWord string `json:"trigger_word"`
	FileIds     string `json:"file_ids"`
}

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
	http.HandleFunc("/setAlias", setAlias)
	http.HandleFunc("/addApp", addApp)
	http.HandleFunc("/addNewReviewsAlert", addNewReviewsAlert)
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

func addNewReviewsAlert(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "{\"text\":\"\\n")
	args, errMessage := getArgs(r)
	if args == nil {
		fmt.Fprintf(w, "%s\"}", errMessage)
		return
	}
	if len(args) != 5 {
		fmt.Fprintf(w, "Wrong use: addNewReviewsAlert unique_name webhook packageName minimum_frequency_in_seconds\"}")
		fmt.Print(args, len(args))
		return
	}
	if _, ok := newReviewsAlerts[args[1]]; ok {
		fmt.Fprintf(w, "There is already an alert named %s.\"}", args[1])
		return
	}

	if !contains(packageList, args[3]) {
		fmt.Fprintf(w, "Package %s is not yet registered.\"}", args[3])
		return
	}
	frequency, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil || frequency <= 0 {
		fmt.Fprintf(w, "%s is not a well formed frequency. Please use a positive number.\"}", args[3])
		return
	}

	newReviewsAlerts[args[1]] = newReviewsAlert{
		webhook:     args[2],
		packageName: args[3],
		frequency:   frequency,
		lastAlerted: time.Now(),
	}
	fmt.Fprintf(w, "Alert %s registered.\"}", args[1])
}

func addApp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "{\"text\":\"\\n")
	args, errMessage := getArgs(r)
	if args == nil {
		fmt.Fprintf(w, "%s\"}", errMessage)
		return
	}
	if len(args) != 2 {
		fmt.Fprintf(w, "Wrong use: addApp packageName\"}")
		return
	}
	if contains(packageList, args[1]) {
		fmt.Fprintf(w, "Package %s already registered.\"}", args[1])
		return
	}

	_, err := service.Reviews.List(args[1]).Do()
	if err != nil {
		fmt.Fprintf(w, "Error registering the app %s: %v\"}", args[1], err.Error())
		return
	}

	packageList = append(packageList, args[1])
	fmt.Fprintf(w, "Package %s added to the system.\"}", args[1])
}

func setAlias(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "{\"text\":\"\\n")
	args, errMessage := getArgs(r)
	if args == nil {
		fmt.Fprintf(w, "%s\"}", errMessage)
		return
	}
	if len(args) != 3 {
		fmt.Fprintf(w, "Wrong use: setAlias alias packageName\"}")
		return
	}
	if contains(packageList, args[2]) {
		if _, ok := aliases[args[1]]; ok {
			fmt.Fprintf(w, "Alias %s already set for app %s\"}", args[1], aliases[args[1]])
			return
		}
		aliases[args[1]] = args[2]
		fmt.Fprintf(w, "Alias %s set for app %s.\"}", args[1], args[2])
		return
	}
	fmt.Fprintf(w, "App %s not registered.\"}", args[2])
}

func serveAppList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "{\"text\":\"\\n")
	fmt.Fprint(w, "Here are all the apps you have registered:\\n")
	for _, packageName := range packageList {
		fmt.Fprint(w, packageName)
		if al := getAliasesForPackage(packageName); len(al) > 0 {
			fmt.Fprint(w, " AKA")
			for _, alias := range al {
				fmt.Fprintf(w, " %s", alias)
			}
		}
		fmt.Fprint(w, "\\n")
	}
	fmt.Fprintf(w, "\"}")
}

func serveList(w http.ResponseWriter, r *http.Request) {
	maxReviewsServed := 10
	fmt.Fprint(w, "{\"text\":\"\\n")
	fmt.Fprintf(w, "Here are the %d latest reviews from each app:\\n", maxReviewsServed)
	reviewsMutex.Lock()
	for key, reviewList := range localReviews {
		fmt.Fprintf(w, "Package Id: %s\\n", key)
		for i, review := range reviewList {
			if i >= maxReviewsServed {
				break
			}
			fmt.Fprintf(w, formatReview(review))
		}
	}
	reviewsMutex.Unlock()
	fmt.Fprintf(w, "\"}")
}
