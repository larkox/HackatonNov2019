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

// Alert stores the important information about what to alert and how often.
type Alert = struct {
	Webhook     string
	PackageName string
	Frequency   int64
	lastAlerted time.Time
}

// NewReviewsAlert declares an alert for new reviews on the system
type NewReviewsAlert = struct {
	Alert
	count int
}

// NewUpdatesAlert declares an alert for updates in the user reviews
type NewUpdatesAlert = struct {
	Alert
	updatedReviews []*androidpublisher.Review
}

func (s *server) watchAlerts() {
	for {
		//s.testAlert(&mockReview)
		time.Sleep(time.Duration(s.config.AlertWatcherTime) * time.Second)
		s.alertNewReviews()
		s.alertNewUpdates()
	}
}

func (s *server) updateAlerts(packageName string, updatedReviews []*androidpublisher.Review, newReviews []*androidpublisher.Review) {
	for _, v := range s.alerts.NewReviewsAlerts {
		if v.PackageName == packageName {
			v.count += len(newReviews)
		}
	}
	for _, v := range s.alerts.NewUpdatesAlerts {
		if v.PackageName == packageName {
			v.updatedReviews = append(updatedReviews, v.updatedReviews...)
		}
	}
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

func (s *server) alertNewUpdates() {
	for _, v := range s.alerts.NewUpdatesAlerts {
		go s.sendUpdatedAlert(v)
	}
}

func (s *server) sendUpdatedAlert(alert *NewUpdatesAlert) {
	if alert.lastAlerted.Unix()+alert.Frequency > time.Now().Unix() {
		return
	}

	if len(alert.updatedReviews) == 0 {
		return
	}

	text := fmt.Sprintf("## Some reviews has been updated:\n")

	s.control.reviewsMutex.Lock()
	defer func() {
		alert.updatedReviews = []*androidpublisher.Review{}
		s.control.reviewsMutex.Unlock()
	}()

	showing := min(len(alert.updatedReviews), s.config.MaxReviewsServed)

	for _, review := range alert.updatedReviews[:showing] {
		text += formatReview(review)
	}
	if showing > s.config.MaxReviewsServed {
		text += fmt.Sprintf("and **%d** more not shown.", len(alert.updatedReviews)-showing)
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
	alert.lastAlerted = time.Now()
}

func (s *server) alertNewReviews() {
	for _, v := range s.alerts.NewReviewsAlerts {
		go s.sendReviewsAlert(v)
	}
}

func (s *server) sendReviewsAlert(alert *NewReviewsAlert) {
	if alert.lastAlerted.Unix()+alert.Frequency > time.Now().Unix() {
		return
	}

	if alert.count == 0 {
		return
	}

	text := "## You have new reviews:\n"

	s.control.reviewsMutex.Lock()
	defer func() {
		alert.count = 0
		s.control.reviewsMutex.Unlock()
	}()

	showing := min(alert.count, s.config.MaxReviewsServed)

	for _, review := range s.localReviews[alert.PackageName][:showing] {
		text += formatReview(review)
	}
	if showing > s.config.MaxReviewsServed {
		text += fmt.Sprintf("and **%d** more not shown.", alert.count-showing)
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
	alert.lastAlerted = time.Now()
}
