package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const commandHelp = `* |/gpreviews connect| - Connect your Mattermost account to your Google Play Developer account
* |/gpreviews disconnect| - Disconnect your Mattermost account from your Google Play Developer account
* |/gpreviews add app packageId| - Add a packageId to the plugin
* |/gpreviews add alias aliasName packageId| - Add aliases for your apps
* |/gpreviews list apps| - List your registered apps on the plugin
* |/gpreviews list reviews [packageId_or_alias] - List your most recent reviews. If no package is stated, show from all packages registered
* |/gpreviews add alert alert_type name webhook packageId_or_alias frequency_in_seconds| - Configure an alert on a incoming webhook for the alert type
  * |alert_type| is the type of alert you want to add
	* newReviews - tell you when there are new reviews
* |/gpreviews list alert alert_type| - List alerts of alert_type
  * |alert_type| is the type of alert you want to add
    * newReviews - tell you when there are new reviews
* |/gpreviews remove alert alert_type alertName| - Remove one alert
  * |alert_type| is the type of alert you want to add
    * newReviews - tell you when there are new reviews`

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "gpreviews",
		DisplayName:      "Google Play Reviews",
		Description:      "Integration with Google Play Reviews.",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect, disconnect, add, list, remove",
		AutoCompleteHint: "[command]",
	}
}

// ExecuteCommand triggers when a command is executed on Mattermost
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	return p.routeRoot(split, args.UserId)
}

func (p *Plugin) routeRoot(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	availableCommands := "Available commands are:\n* `list`\n* `set`\n* `add`\n* `remove`"
	if len(args) < 2 {
		message := fmt.Sprintf(":x:Program `\"%s\"` needs a command. %s", args[0], availableCommands)
		return commandErrorResponse(message)
	}

	switch args[1] {
	case "list":
		return p.routeList(args, userID)
	case "add":
		return p.routeAdd(args, userID)
	case "remove":
		return p.routeRemove(args, userID)
	case "connect":
		return p.connect(userID)
	case "disconnect":
		return p.disconnect(userID)
	default:
		message := fmt.Sprintf(":x:Command `\"%s\"` not found. %s", args[1], availableCommands)
		return commandErrorResponse(message)
	}
}

func (p *Plugin) routeList(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	availableLists := "Available lists are:\n* `apps`\n* `alerts`\n* `reviews`"
	if len(args) < 3 {
		message := fmt.Sprintf(":x:Command `\"%s %s\"` need something to list. %s", args[0], args[1], availableLists)
		return commandErrorResponse(message)
	}
	switch args[2] {
	case "apps":
		return p.serveAppList(args, userID)
	case "alerts":
		return p.routeListAlerts(args, userID)
	case "reviews":
		return p.serveList(args, userID)
	default:
		message := fmt.Sprintf(":x:Nothing called `\"%s\"` can be listed. %s", args[2], availableLists)
		return commandErrorResponse(message)
	}
}

func (p *Plugin) routeListAlerts(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	alertTypes := "Available types are:\n* `newReviews`"
	if len(args) < 4 {
		message := fmt.Sprintf(":x:Command `\"%s %s %s\"` needs the alert type. %s", args[0], args[1], args[2], alertTypes)
		return commandErrorResponse(message)
	}
	switch args[3] {
	case "newReviews":
		return p.serveListNewReviewsAlerts(args, userID)
	default:
		message := fmt.Sprintf(":x:Alert type `\"%s\"` not found. %s", args[3], alertTypes)
		return commandErrorResponse(message)
	}
}

func (p *Plugin) routeAdd(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	availableAdds := "Available things to add are:\n* `app`\n* `alert`\n* `alias`"
	if len(args) < 3 {
		message := fmt.Sprintf(":x:Command `\"%s %s\"` need something to add. %s", args[0], args[1], availableAdds)
		return commandErrorResponse(message)
	}
	switch args[2] {
	case "app":
		return p.addApp(args, userID)
	case "alert":
		return p.routeAddAlert(args, userID)
	case "alias":
		return p.addAlias(args, userID)
	default:
		message := fmt.Sprintf(":x:Nothing named `\"%s\"` can be added. %s", args[2], availableAdds)
		return commandErrorResponse(message)
	}
}

func (p *Plugin) routeAddAlert(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	alertTypes := "Available types are:\n* `newReviews`"
	if len(args) < 4 {
		message := fmt.Sprintf(":x:Command `\"%s %s %s\"` needs the alert type. %s", args[0], args[1], args[2], alertTypes)
		return commandErrorResponse(message)
	}
	switch args[3] {
	case "newReviews":
		return p.addNewReviewsAlert(args, userID)
	default:
		message := fmt.Sprintf(":x:Alert type `\"%s\"` not found. %s", args[3], alertTypes)
		return commandErrorResponse(message)
	}
}

func (p *Plugin) routeRemove(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	availableRemoves := "Available things to remove are:\n* `alert`"
	if len(args) < 3 {
		message := fmt.Sprintf(":x:Command `\"%s %s\"` needs the alert type. %s", args[0], args[1], availableRemoves)
		return commandErrorResponse(message)
	}
	switch args[2] {
	case "alert":
		return p.routeRemoveAlert(args, userID)
	default:
		message := fmt.Sprintf(":x:Nothing named `\"%s\"` can be removed. %s", args[2], availableRemoves)
		return commandErrorResponse(message)
	}
}

func (p *Plugin) routeRemoveAlert(args []string, userID string) (*model.CommandResponse, *model.AppError) {
	alertTypes := "Available types are:\n* `newReviews`"
	if len(args) < 4 {
		message := fmt.Sprintf(":x:Command `\"%s %s %s\"` needs the alert type. %s", args[0], args[1], args[2], alertTypes)
		return commandErrorResponse(message)
	}
	switch args[3] {
	case "newReviews":
		return p.removeNewReviewsAlert(args, userID)
	default:
		message := fmt.Sprintf(":x:Alert type `\"%s\"` not found. %s", args[3], alertTypes)
		return commandErrorResponse(message)
	}
}

func commandErrorResponse(message string) (*model.CommandResponse, *model.AppError) {
	response := &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         message,
		Username:     "GPReviews",
		//IconURL:
	}
	return response, nil
}
