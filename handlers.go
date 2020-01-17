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
		text += fmt.Sprintf(":x:Wrong use: `%s %s %s %s unique_name`.", args[0], args[1], args[2], args[3])
		return
	}
	if _, ok := s.alerts.NewReviewsAlerts[alertName]; !ok {
		text += fmt.Sprintf(":x:There no alert named **%s**.", alertName)
		return
	}

	delete(s.alerts.NewReviewsAlerts, alertName)
	s.SaveAlerts()
	text += fmt.Sprintf(":white_check_mark:Alert **%s** removed.", alertName)
}

func (s *server) serveListNewReviewsAlerts(w http.ResponseWriter, r *http.Request, args []string) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	text += "## Here are all the alerts you have registered:\n"
	for k, v := range s.alerts.NewReviewsAlerts {
		text += fmt.Sprintf("* Alert **\"%s\"**: From package **%s** every **%v seconds** at most on webhook **%s**\n", k, v.PackageName, v.Frequency, v.Webhook)
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
		text += fmt.Sprintf(":x:Wrong use: `%s %s %s %s unique_name webhook packageName_or_alias minimum_frequency_in_seconds`", args[0], args[1], args[2], args[3])
		return
	}
	uniqueName := args[4]
	webhook := args[5]
	packageNameOrAlias := args[6]
	minimumFrequency := args[7]

	if _, ok := s.alerts.NewReviewsAlerts[uniqueName]; ok {
		text += fmt.Sprintf(":x:There is already an alert named **%s**.", uniqueName)
		return
	}

	packageName, ok := getPackageNameFromArgs(packageNameOrAlias, s.packageList, s.aliases)
	if !ok {
		text += fmt.Sprintf(":x:Package **%s** is not yet registered.", packageNameOrAlias)
		return
	}

	frequency, err := strconv.ParseInt(minimumFrequency, 10, 64)
	if err != nil || frequency <= 0 {
		text += fmt.Sprintf(":x:**%s** is not a well formed frequency. Please use a positive number.", minimumFrequency)
		return
	}

	s.alerts.NewReviewsAlerts[uniqueName] = &NewReviewsAlert{
		Alert: Alert{
			Webhook:     webhook,
			PackageName: packageName,
			Frequency:   frequency,
			lastAlerted: time.Now(),
		},
	}
	s.SaveAlerts()

	text += fmt.Sprintf(":white_check_mark:Alert **%s** registered.", uniqueName)
}

func (s *server) addApp(w http.ResponseWriter, r *http.Request, args []string) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	if len(args) != 4 {
		text += fmt.Sprintf(":x:Wrong use: `%s %s %s packageName`", args[0], args[1], args[2])
		return
	}

	packageName := args[3]

	if contains(s.packageList, packageName) {
		text += fmt.Sprintf(":x:Package **%s** already registered.", packageName)
		return
	}

	_, err := s.service.Reviews.List(packageName).Do()
	if err != nil {
		text += fmt.Sprintf(":x:Error registering the app **%s**: **%v**", packageName, err.Error())
		return
	}

	s.packageList = append(s.packageList, packageName)
	s.SavePackages()
	text += fmt.Sprintf(":white_check_mark:Package **%s** added to the system.", packageName)
}

func (s *server) addAlias(w http.ResponseWriter, r *http.Request, args []string) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	if len(args) != 5 {
		text += fmt.Sprintf(":x:Wrong use: `%s %s %s aliasName packageName`", args[0], args[1], args[2])
		return
	}

	aliasName := args[3]
	packageName := args[4]

	if !contains(s.packageList, packageName) {
		fmt.Fprintf(w, ":x:App **%s** not registered.", packageName)
		return
	}

	if _, ok := s.aliases[aliasName]; ok {
		text += fmt.Sprintf(":x:Alias **%s** already set for app **%s**.", aliasName, s.aliases[aliasName])
		return
	}

	s.aliases[aliasName] = packageName
	s.SaveAliases()
	text += fmt.Sprintf(":white_check_mark:Alias **%s** added for app **%s**.", aliasName, packageName)
}

func (s *server) setConfig(w http.ResponseWriter, r *http.Request, args []string) {
	configAvailable := "Config fields available are:\n* `GetListTime`\n* `MaxReviewsServed`\n* `SaveConfig`\n* `SavePackages`\n* `SaveAliases`\n* `SaveReviews`\n* `SaveAlerts`"
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	if len(args) != 5 {
		text += fmt.Sprintf(":x:Wrong use: `%s %s %s configField configValue`", args[0], args[1], args[2])
		return
	}

	if !isField(args[3], s.config) {
		text += fmt.Sprintf(":x:**%s** is not a config field. %s", args[3], configAvailable)
		return
	}

	errorCannotConvertFormat := ":x:**%s** cannot be converted to a number. Error: **%s**"
	errorPositiveFormat := ":x:**%s** is 0 or below. Please, use a positive number."
	errorBooleanFormat := ":x:**%s** is not valid boolean. Use true or false."
	configUpdatedFormat := ":white_check_mark:Config field **%s** updated to **%s**."

	switch args[3] {
	case "GetListTime":
		i, err := strconv.Atoi(args[4])
		if err != nil {
			text += fmt.Sprintf(errorCannotConvertFormat, args[4], err.Error())
			return
		}
		if i <= 0 {
			text += fmt.Sprintf(errorPositiveFormat, args[4])
			return
		}
		s.config.GetListTime = i
		text += fmt.Sprintf(configUpdatedFormat, args[3], args[4])
	case "MaxReviewsServed":
		i, err := strconv.Atoi(args[4])
		if err != nil {
			text += fmt.Sprintf(errorCannotConvertFormat, args[4], err.Error())
			return
		}
		if i <= 0 {
			text += fmt.Sprintf(errorPositiveFormat, args[4])
			return
		}
		s.config.MaxReviewsServed = i
		text += fmt.Sprintf(configUpdatedFormat, args[3], args[4])
	case "SaveConfig":
		value := args[4] == "true"
		if !value && args[4] != "false" {
			text += fmt.Sprintf(errorBooleanFormat, args[4])
			return
		}
		s.config.SaveConfig = value
		text += fmt.Sprintf(configUpdatedFormat, args[3], args[4])
	case "SavePackages":
		value := args[4] == "true"
		if !value && args[4] != "false" {
			text += fmt.Sprintf(errorBooleanFormat, args[4])
			return
		}
		s.config.SavePackages = value
		text += fmt.Sprintf(configUpdatedFormat, args[3], args[4])
	case "SaveAliases":
		value := args[4] == "true"
		if !value && args[4] != "false" {
			text += fmt.Sprintf(errorBooleanFormat, args[4])
			return
		}
		s.config.SaveAliases = value
		text += fmt.Sprintf(configUpdatedFormat, args[3], args[4])
	case "SaveReviews":
		value := args[4] == "true"
		if !value && args[4] != "false" {
			text += fmt.Sprintf(errorBooleanFormat, args[4])
			return
		}
		s.config.SaveReviews = value
		text += fmt.Sprintf(configUpdatedFormat, args[3], args[4])
	case "SaveAlerts":
		value := args[4] == "true"
		if !value && args[4] != "false" {
			text += fmt.Sprintf(errorBooleanFormat, args[4])
			return
		}
		s.config.SaveAlerts = value
		text += fmt.Sprintf(configUpdatedFormat, args[3], args[4])
	default:
		text += fmt.Sprintf(":x:You cannot set **%s** from Mattermost.", args[3])
		return
	}
	s.SaveConfig()
}

func (s *server) serveAppList(w http.ResponseWriter, r *http.Request, args []string) {
	var text string
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	defer encoder.Encode(response)

	text += "## Here are all the apps you have registered:\n"
	for _, packageName := range s.packageList {
		text += fmt.Sprintf("* **%s**", packageName)
		if al := getAliasesForPackage(packageName, s.aliases); len(al) > 0 {
			text += " AKA"
			for _, alias := range al {
				text += fmt.Sprintf(" **_%s_**", alias)
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

	text += fmt.Sprintf("## Here are the %d latest reviews from each app:\n", s.config.MaxReviewsServed)
	s.control.reviewsMutex.Lock()
	defer s.control.reviewsMutex.Unlock()
	for key, reviewList := range s.localReviews {
		text += fmt.Sprintf("* Package Id: %s\n", key)
		for i, review := range reviewList {
			if i >= s.config.MaxReviewsServed {
				break
			}
			text += formatReview(review)
		}
	}
}
