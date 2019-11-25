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

type NewReviewsAlert = struct {
	Webhook     string
	PackageName string
	Frequency   int64
	LastAlerted time.Time
}

func (s *server) testAlert(review *androidpublisher.Review) {
	for k, v := range s.alerts.NewReviewsAlerts {
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

		_, err = http.Post(v.Webhook, "application/json", strings.NewReader(string(b)))
		if err != nil {
			fmt.Print(err.Error())
			return
		}
	}
}

func (s *server) alertNewReviews(packageName string) {
	alertSync := make(chan bool)
	waitFor := 0
	for _, v := range s.alerts.NewReviewsAlerts {
		if v.PackageName == packageName {
			go s.sendReviewsAlert(v, alertSync)
			waitFor++
		}
	}
	for i := 0; i < waitFor; i++ {
		_ = <-alertSync
	}
	s.control.reviewsMutex.Lock()
	defer s.control.reviewsMutex.Unlock()

	s.newReviewsCounts[packageName] = 0
}

func (s *server) sendReviewsAlert(alert NewReviewsAlert, alertSync chan bool) {
	if alert.LastAlerted.Unix()+alert.Frequency > time.Now().Unix() {
		return
	}

	var text string

	s.control.reviewsMutex.Lock()
	defer func() {
		s.control.reviewsMutex.Unlock()
		alertSync <- true
	}()

	showing := min(s.newReviewsCounts[alert.PackageName], s.config.MaxReviewsServed)

	for _, review := range s.localReviews[alert.PackageName][:showing] {
		text += formatReview(review)
	}
	if showing > s.config.MaxReviewsServed {
		text += fmt.Sprintf("and %d more not shown.", s.newReviewsCounts[alert.PackageName]-showing)
	}

	request := model.IncomingWebhookRequest{
		Text: text,
	}

	b, err := json.Marshal(request)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	_, err = http.Post(alert.Webhook, "application/json", strings.NewReader(string(b)))
	if err != nil {
		fmt.Print(err.Error())
		return
	}
}
