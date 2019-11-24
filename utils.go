package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/androidpublisher/v3"
)

func getArgs(r *http.Request) (args []string, errMessage string) {
	requestJSON := OutgoingWebhookJSON{}
	err := json.NewDecoder(r.Body).Decode(&requestJSON)
	if err != nil {
		errMessage = fmt.Sprintf("Error on request:%s", err.Error())
		return nil, errMessage
	}
	if len(strings.Split(requestJSON.Text, "\n")) != 1 {
		errMessage = fmt.Sprint("Please, put all the commands on the same line.")
		return nil, errMessage
	}
	if len(strings.Split(requestJSON.Text, "\"")) != 1 {
		errMessage = fmt.Sprint("Please, do not use \".")
		return nil, errMessage
	}
	args = strings.Split(requestJSON.Text, " ")
	return args, ""
}

func contains(slice []string, value string) bool {
	for _, elem := range slice {
		if elem == value {
			return true
		}
	}
	return false
}

func getAliasesForPackage(packageName string) []string {
	result := []string{}
	for k, v := range aliases {
		if v == packageName {
			result = append(result, k)
		}
	}
	return result
}

func removeElement(list []*androidpublisher.Review, index int) []*androidpublisher.Review {
	var newList []*androidpublisher.Review
	if index == len(list)-1 {
		newList = list[:index]
	} else {
		newList = append(list[:index], list[index+1:]...)
	}
	return newList
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}