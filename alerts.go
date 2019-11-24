package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"
)

type newReviewsAlert = struct {
	webhook     string
	packageName string
	frequency   int64
	lastAlerted time.Time
}

var newReviewsAlerts = make(map[string]newReviewsAlert)

func testAlert(review *androidpublisher.Review) {
	for k, v := range newReviewsAlerts {
		message := "Test alert for alert named " + k
		payload := "{\"text\": \"" + message + "\\n" + formatReview(review) + "\"}"
		_, err := http.Post(v.webhook, "application/json", strings.NewReader(payload))
		if err != nil {
			fmt.Print(err.Error())
		}
	}
}

func alertNewReviews(packageName string) {
	alertSync := make(chan bool)
	waitFor := 0
	for _, v := range newReviewsAlerts {
		if v.packageName == packageName {
			go sendReviewsAlert(v, alertSync)
			waitFor++
		}
	}
	for i := 0; i < waitFor; i++ {
		_ = <-alertSync
	}
	reviewsMutex.Lock()
	defer reviewsMutex.Unlock()

	newReviewsCounts[packageName] = 0
}

func sendReviewsAlert(alert newReviewsAlert, alertSync chan bool) {
	if alert.lastAlerted.Unix()+alert.frequency > time.Now().Unix() {
		return
	}

	payload := "{\"text\": \""
	reviewsMutex.Lock()
	defer func() {
		reviewsMutex.Unlock()
		alertSync <- true
	}()

	maxReviewsServed := 10
	showing := min(newReviewsCounts[alert.packageName], maxReviewsServed)

	for _, review := range localReviews[alert.packageName][:showing] {
		payload += formatReview(review)
	}
	if showing > maxReviewsServed {
		payload += fmt.Sprintf("and %d more not shown.", newReviewsCounts[alert.packageName]-showing)
	}

	payload += "\"}"
	_, err := http.Post(alert.webhook, "application/json", strings.NewReader(payload))
	if err != nil {
		fmt.Print(err.Error())
		return
	}
}
