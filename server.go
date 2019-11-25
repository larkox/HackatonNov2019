package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/androidpublisher/v3"
)

type server struct {
	config      ServerConfig
	packageList []string
	aliases     map[string]string
	service     *androidpublisher.Service
	alerts      AlertsContainer
	// Newer reviews will always be on the lower ids of the slice
	localReviews     map[string][]*androidpublisher.Review
	newReviewsCounts map[string]int
}

// AlertsContainer contains the maps to the different kinds of alerts
type AlertsContainer = struct {
	newReviewsAlerts map[string]newReviewsAlert
}

// ServerConfig contains all the configuration values for the server
type ServerConfig = struct {
	getListTime      time.Duration
	maxReviewsServed int
}

func newServer() *server {
	newServer := server{
		packageList:      []string{},
		aliases:          make(map[string]string),
		localReviews:     make(map[string][]*androidpublisher.Review),
		newReviewsCounts: make(map[string]int),
		config: ServerConfig{
			getListTime:      30 * time.Second,
			maxReviewsServed: 10,
		},
		alerts: AlertsContainer{
			newReviewsAlerts: make(map[string]newReviewsAlert),
		},
	}
	if err := newServer.initService(); err != nil {
		fmt.Println("Error initializing the service:", err.Error())
		return nil
	}

	newServer.routes()
	go newServer.getAllReviews()

	return &newServer
}

func (s *server) initService() error {
	ctx := context.Background()
	service, err := androidpublisher.NewService(ctx)
	if err != nil {
		return err
	}

	s.service = service
	return nil
}
