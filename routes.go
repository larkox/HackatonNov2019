package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/model"
)

func (s *server) routes() {
	http.HandleFunc("/", s.routeRoot)
}

func (s *server) routeRoot(w http.ResponseWriter, r *http.Request) {
	args, errMessage := getArgs(r)
	if args == nil {
		printRouteError(w, errMessage)
		return
	}

	availableCommands := "Available commands are:\nlist\nset\nadd\nremove"
	if len(args) < 2 {
		message := fmt.Sprintf("Program \"%s\" needs a command. %s", args[0], availableCommands)
		printRouteError(w, message)
		return
	}

	switch args[1] {
	case "list":
		s.routeList(w, r, args)
	case "set":
		s.routeSet(w, r, args)
	case "add":
		s.routeAdd(w, r, args)
	case "remove":
		s.routeRemove(w, r, args)
	default:
		message := fmt.Sprintf("Command \"%s\" not found. %s", args[1], availableCommands)
		printRouteError(w, message)
	}
}

func (s *server) routeList(w http.ResponseWriter, r *http.Request, args []string) {
	availableLists := "Available lists are:\napps\nalerts\nreviews\nremove"
	if len(args) < 3 {
		message := fmt.Sprintf("Command \"%s %s\" need something to list. %s", args[0], args[1], availableLists)
		printRouteError(w, message)
		return
	}
	switch args[2] {
	case "apps":
		s.serveAppList(w, r, args)
	case "alerts":
		s.routeListAlerts(w, r, args)
	case "reviews":
		s.serveList(w, r, args)
	default:
		message := fmt.Sprintf("Nothing called \"%s\" not found. %s", args[2], availableLists)
		printRouteError(w, message)
	}
}

func (s *server) routeListAlerts(w http.ResponseWriter, r *http.Request, args []string) {
	alertTypes := "Available types are:\nnewReviews"
	if len(args) < 4 {
		message := fmt.Sprintf("Command \"%s %s %s\" needs the alert type. %s", args[0], args[1], args[2], alertTypes)
		printRouteError(w, message)
		return
	}
	switch args[3] {
	case "newReviews":
		s.serveListNewReviewsAlerts(w, r, args)
	default:
		message := fmt.Sprintf("Alert type \"%s\" not found. %s", args[3], alertTypes)
		printRouteError(w, message)
	}
}

func (s *server) routeSet(w http.ResponseWriter, r *http.Request, args []string) {
	availableSets := "Available things to set are:\nalias"
	if len(args) < 3 {
		message := fmt.Sprintf("Command \"%s %s\" need something to set. %s", args[0], args[1], availableSets)
		printRouteError(w, message)
		return
	}
	switch args[2] {
	case "alias":
		s.setAlias(w, r, args)
	default:
		message := fmt.Sprintf("Nothing named \"%s\" can be set. %s", args[2], availableSets)
		printRouteError(w, message)
	}
}

func (s *server) routeAdd(w http.ResponseWriter, r *http.Request, args []string) {
	availableAdds := "Available things to add are:\napp\nalert"
	if len(args) < 3 {
		message := fmt.Sprintf("Command \"%s %s\" need something to add. %s", args[0], args[1], availableAdds)
		printRouteError(w, message)
		return
	}
	switch args[2] {
	case "app":
		s.addApp(w, r, args)
	case "alert":
		s.routeAddAlert(w, r, args)
	default:
		message := fmt.Sprintf("Nothing named \"%s\" can be added. %s", args[2], availableAdds)
		printRouteError(w, message)
	}
}

func (s *server) routeAddAlert(w http.ResponseWriter, r *http.Request, args []string) {
	alertTypes := "Available types are:\nnewReviews"
	if len(args) < 4 {
		message := fmt.Sprintf("Command \"%s %s %s\" needs the alert type. %s", args[0], args[1], args[2], alertTypes)
		printRouteError(w, message)
		return
	}
	switch args[3] {
	case "newReviews":
		s.addNewReviewsAlert(w, r, args)
	default:
		message := fmt.Sprintf("Alert type \"%s\" not found. %s", args[3], alertTypes)
		printRouteError(w, message)
	}
}

func (s *server) routeRemove(w http.ResponseWriter, r *http.Request, args []string) {
	availableRemoves := "Available things to remove are:\nalert"
	if len(args) < 3 {
		message := fmt.Sprintf("Command \"%s %s\" needs the alert type. %s", args[0], args[1], availableRemoves)
		printRouteError(w, message)
		return
	}
	switch args[2] {
	case "alert":
		s.routeRemoveAlert(w, r, args)
	default:
		message := fmt.Sprintf("Nothing named \"%s\" can be removed. %s", args[2], availableRemoves)
		printRouteError(w, message)
	}
}

func (s *server) routeRemoveAlert(w http.ResponseWriter, r *http.Request, args []string) {
	alertTypes := "Available types are:\nnewReviews"
	if len(args) < 4 {
		message := fmt.Sprintf("Command \"%s %s %s\" needs the alert type. %s", args[0], args[1], args[2], alertTypes)
		printRouteError(w, message)
		return
	}
	switch args[3] {
	case "newReviews":
		s.removeNewReviewsAlert(w, r, args)
	default:
		message := fmt.Sprintf("Alert type \"%s\" not found. %s", args[3], alertTypes)
		printRouteError(w, message)
	}
}

func printRouteError(w http.ResponseWriter, message string) {
	text := message
	response := model.OutgoingWebhookResponse{
		Text: &text,
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}
