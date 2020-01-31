package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) removeNewReviewsAlert(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	var message string

	alertName := args[4]

	if len(args) != 5 {
		message += fmt.Sprintf(":x:Wrong use: `%s %s %s %s unique_name`.", args[0], args[1], args[2], args[3])
		return commandErrorResponse(message)
	}
	if _, ok := p.alerts.NewReviewsAlerts[userID][alertName]; !ok {
		message += fmt.Sprintf(":x:There no alert named **%s**.", alertName)
		return commandErrorResponse(message)
	}

	delete(p.alerts.NewReviewsAlerts[userID], alertName)
	p.SaveAlerts()
	message += fmt.Sprintf(":white_check_mark:Alert **%s** removed.", alertName)
	return commandStatusResponse(message)
}

func (p *Plugin) serveListNewReviewsAlerts(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	var message string

	message += "## Here are all the alerts you have registered:\n"
	for k, v := range p.alerts.NewReviewsAlerts[userID] {
		message += fmt.Sprintf("* Alert **\"%s\"**: From package **%s** every **%v seconds** at most on webhook **%s**\n", k, v.PackageName, v.Frequency, v.Webhook)
	}
	return commandStatusResponse(message)
}

func (p *Plugin) addNewReviewsAlert(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	var message string

	if len(args) != 8 {
		message += fmt.Sprintf(":x:Wrong use: `%s %s %s %s unique_name webhook packageName_or_alias minimum_frequency_in_seconds`", args[0], args[1], args[2], args[3])
		return commandErrorResponse(message)
	}
	uniqueName := args[4]
	webhook := args[5]
	packageNameOrAlias := args[6]
	minimumFrequency := args[7]

	if _, ok := p.alerts.NewReviewsAlerts[userID][uniqueName]; ok {
		message += fmt.Sprintf(":x:There is already an alert named **%s**.", uniqueName)
		return commandErrorResponse(message)
	}

	packageName, ok := getPackageNameFromArgs(packageNameOrAlias, userID, p.packageList, p.aliases[userID])
	if !ok {
		message += fmt.Sprintf(":x:Package **%s** is not yet registered.", packageNameOrAlias)
		return commandErrorResponse(message)
	}

	frequency, err := strconv.ParseInt(minimumFrequency, 10, 64)
	if err != nil || frequency <= 0 {
		message += fmt.Sprintf(":x:**%s** is not a well formed frequency. Please use a positive number.", minimumFrequency)
		return commandErrorResponse(message)
	}

	if _, ok := p.alerts.NewReviewsAlerts[userID]; !ok {
		p.alerts.NewReviewsAlerts[userID] = make(map[string]*NewReviewsAlert)
	}

	p.alerts.NewReviewsAlerts[userID][uniqueName] = &NewReviewsAlert{
		Alert: Alert{
			Webhook:     webhook,
			PackageName: packageName,
			Frequency:   frequency,
			lastAlerted: time.Now(),
		},
	}
	p.SaveAlerts()

	message += fmt.Sprintf(":white_check_mark:Alert **%s** registered.", uniqueName)
	return commandStatusResponse(message)
}

func (p *Plugin) addApp(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	var message string

	if len(args) != 4 {
		message += fmt.Sprintf(":x:Wrong use: `%s %s %s packageName`", args[0], args[1], args[2])
		return commandErrorResponse(message)
	}

	packageName := args[3]
	packageInfo := PackageInfo{Name: packageName, UserID: userID}

	if contains(p.packageList, packageInfo) {
		message += fmt.Sprintf(":x:Package **%s** already registered.", packageName)
		return commandErrorResponse(message)
	}

	service := p.getService(userID)

	_, err := service.List(packageName).Do()
	if err != nil {
		message += fmt.Sprintf(":x:Error registering the app **%s**: **%v**", packageName, err.Error())
		return commandErrorResponse(message)
	}

	p.packageList = append(p.packageList, packageInfo)
	p.SavePackages()
	message += fmt.Sprintf(":white_check_mark:Package **%s** added to the system.", packageName)
	return commandStatusResponse(message)
}

func (p *Plugin) addAlias(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	var message string

	if len(args) != 5 {
		message += fmt.Sprintf(":x:Wrong use: `%s %s %s aliasName packageName`", args[0], args[1], args[2])
		return commandErrorResponse(message)
	}

	aliasName := args[3]
	packageName := args[4]

	if !contains(p.packageList, PackageInfo{Name: packageName, UserID: userID}) {
		message += fmt.Sprintf(":x:App **%s** not registered.", packageName)
		return commandErrorResponse(message)
	}

	if packageName, ok := p.aliases[aliasName]; ok {
		message += fmt.Sprintf(":x:Alias **%s** already set for app **%s**.", aliasName, packageName)
		return commandErrorResponse(message)
	}

	if _, ok := p.aliases[userID]; !ok {
		p.aliases[userID] = *new(map[string]string)
	}

	p.aliases[userID][aliasName] = packageName
	p.SaveAliases()
	message += fmt.Sprintf(":white_check_mark:Alias **%s** added for app **%s**.", aliasName, packageName)
	return commandStatusResponse(message)
}

func (p *Plugin) serveAppList(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	var message string

	message += "## Here are all the apps you have registered:\n"
	for _, packageInfo := range p.packageList {
		if userID == packageInfo.UserID {
			message += fmt.Sprintf("* **%s**", packageInfo.Name)
			if al := getAliasesForPackage(packageInfo.Name, p.aliases[userID]); len(al) > 0 {
				message += " AKA"
				for _, alias := range al {
					message += fmt.Sprintf(" **_%s_**", alias)
				}
			}
			message += "\n"
		}
	}
	return commandStatusResponse(message)
}

func (p *Plugin) serveList(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	var message string

	config := p.getConfiguration()

	message += fmt.Sprintf("## Here are the %d latest reviews from each app:\n", config.MaxReviewsServed)
	p.control.reviewsMutex.RLock()
	defer p.control.reviewsMutex.RUnlock()
	for key, reviewList := range p.localReviews[userID] {
		message += fmt.Sprintf("* Package Id: %s\n", key)
		for i, review := range reviewList {
			if i >= config.MaxReviewsServed {
				break
			}
			message += formatReview(review)
		}
	}
	return commandStatusResponse(message)
}

func (p *Plugin) connect(userID string) (*model.CommandResponse, *model.AppError) {
	config := p.API.GetConfig()
	if config.ServiceSettings.SiteURL == nil {
		return commandErrorResponse("Encountered an error connecting to Google Play: SiteURL is not set-up.")
	}

	siteURL := *config.ServiceSettings.SiteURL
	if i := len(siteURL); i > 0 && siteURL[i-1] == '/' {
		siteURL = siteURL[:i-1]
	}
	userInfo, _ := p.getGooglePlayUserInfo(userID)
	if userInfo == nil {
		return commandStatusResponse(fmt.Sprintf("[Click here to link your Google Play account.](%s/plugins/com.mattermost.google-play-reviews/oauth/connect)", siteURL))
	}

	return commandStatusResponse("Google Play Reviews connected and running.")
}

func (p *Plugin) disconnect(userID string) (*model.CommandResponse, *model.AppError) {
	userInfo, _ := p.getGooglePlayUserInfo(userID)
	if userInfo != nil {
		p.API.KVDelete(userID + GooglePlayTokenKey)
	}

	return commandStatusResponse("Correctly disconnected from Google Play.")
}

func commandStatusResponse(message string) (*model.CommandResponse, *model.AppError) {
	response := &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         message,
		Username:     "GPReviews",
		//IconURL:
	}
	return response, nil
}
