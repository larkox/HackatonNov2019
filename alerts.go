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

func (s *server) testAlert(review *androidpublisher.Review) {
	for k, v := range s.alerts.newReviewsAlerts {
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

func (s *server) alertNewReviews(packageName string) {
	alertSync := make(chan bool)
	waitFor := 0
	for _, v := range s.alerts.newReviewsAlerts {
		if v.packageName == packageName {
			go s.sendReviewsAlert(v, alertSync)
			waitFor++
		}
	}
	for i := 0; i < waitFor; i++ {
		_ = <-alertSync
	}
	reviewsMutex.Lock()
	defer reviewsMutex.Unlock()

	s.newReviewsCounts[packageName] = 0
}

func (s *server) sendReviewsAlert(alert newReviewsAlert, alertSync chan bool) {
	if alert.lastAlerted.Unix()+alert.frequency > time.Now().Unix() {
		return
	}

	var text string

	reviewsMutex.Lock()
	defer func() {
		reviewsMutex.Unlock()
		alertSync <- true
	}()

	showing := min(s.newReviewsCounts[alert.packageName], s.config.maxReviewsServed)

	for _, review := range s.localReviews[alert.packageName][:showing] {
		text += formatReview(review)
	}
	if showing > s.config.maxReviewsServed {
		text += fmt.Sprintf("and %d more not shown.", s.newReviewsCounts[alert.packageName]-showing)
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
