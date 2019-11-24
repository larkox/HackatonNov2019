package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/model"

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
		text := fmt.Sprintf("Test alert for alert named %s\n", k)
		text += formatReview(review)
		request := model.IncomingWebhookRequest{
			Text: text,
		}

		b, err := json.Marshal(request)

		if err != nil {
			fmt.Print(err.Error())
			return
		}

		_, err = http.Post(v.webhook, "application/json", strings.NewReader(string(b)))
		if err != nil {
			fmt.Print(err.Error())
			return
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

	var text string

	reviewsMutex.Lock()
	defer func() {
		reviewsMutex.Unlock()
		alertSync <- true
	}()

	maxReviewsServed := 10
	showing := min(newReviewsCounts[alert.packageName], maxReviewsServed)

	for _, review := range localReviews[alert.packageName][:showing] {
		text += formatReview(review)
	}
	if showing > maxReviewsServed {
		text += fmt.Sprintf("and %d more not shown.", newReviewsCounts[alert.packageName]-showing)
	}

	request := model.IncomingWebhookRequest{
		Text: text,
	}

	b, err := json.Marshal(request)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	_, err = http.Post(alert.webhook, "application/json", strings.NewReader(string(b)))
	if err != nil {
		fmt.Print(err.Error())
		return
	}
}
