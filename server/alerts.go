package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"

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

func (p *Plugin) watchAlerts() {
	for {
		//s.testAlert(&mockReview)
		config := p.getConfiguration()
		time.Sleep(time.Duration(config.AlertWatcherTime) * time.Second)
		p.alertNewReviews()
		p.alertNewUpdates()
	}
}

func (p *Plugin) updateAlerts(packageName string, userID string, updatedReviews []*androidpublisher.Review, newReviews []*androidpublisher.Review) {
	for _, v := range p.alerts.NewReviewsAlerts[userID] {
		if v.PackageName == packageName {
			v.count += len(newReviews)
		}
	}
	for _, v := range p.alerts.NewUpdatesAlerts[userID] {
		if v.PackageName == packageName {
			v.updatedReviews = append(updatedReviews, v.updatedReviews...)
		}
	}
}

func (p *Plugin) testAlert(review *androidpublisher.Review) {
	for _, alerts := range p.alerts.NewReviewsAlerts {
		for k, v := range alerts {
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
}

func (p *Plugin) alertNewUpdates() {
	for _, alerts := range p.alerts.NewUpdatesAlerts {
		for _, v := range alerts {
			go p.sendUpdatedAlert(v)
		}
	}
}

func (p *Plugin) sendUpdatedAlert(alert *NewUpdatesAlert) {
	if alert.lastAlerted.Unix()+alert.Frequency > time.Now().Unix() {
		return
	}

	if len(alert.updatedReviews) == 0 {
		return
	}

	text := fmt.Sprintf("## Some reviews has been updated:\n")

	p.control.reviewsMutex.Lock()
	defer func() {
		alert.updatedReviews = []*androidpublisher.Review{}
		p.control.reviewsMutex.Unlock()
	}()

	config := p.getConfiguration()
	showing := min(len(alert.updatedReviews), config.MaxReviewsServed)

	for _, review := range alert.updatedReviews[:showing] {
		text += formatReview(review)
	}
	if showing > config.MaxReviewsServed {
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

func (p *Plugin) alertNewReviews() {
	for userID, alerts := range p.alerts.NewReviewsAlerts {
		for _, v := range alerts {
			go p.sendReviewsAlert(v, userID)
		}
	}
}

func (p *Plugin) sendReviewsAlert(alert *NewReviewsAlert, userID string) {
	if alert.lastAlerted.Unix()+alert.Frequency > time.Now().Unix() {
		return
	}

	if alert.count == 0 {
		return
	}

	text := "## You have new reviews:\n"

	p.control.reviewsMutex.Lock()
	defer func() {
		alert.count = 0
		p.control.reviewsMutex.Unlock()
	}()

	config := p.getConfiguration()
	showing := min(alert.count, config.MaxReviewsServed)

	for _, review := range p.localReviews[userID][alert.PackageName][:showing] {
		text += formatReview(review)
	}
	if showing > config.MaxReviewsServed {
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
