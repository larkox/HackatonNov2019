package main

import (
	"context"
	"fmt"
	"sync"
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
	persistency      persistencyInt
	control          ControlUtils
}

//ControlUtils contains all the mutex used for flow control
type ControlUtils struct {
	reviewsMutex sync.Mutex
}

// AlertsContainer contains the maps to the different kinds of alerts
type AlertsContainer struct {
	NewReviewsAlerts map[string]NewReviewsAlert
}

// ServerConfig contains all the configuration values for the server
type ServerConfig struct {
	GetListTime      time.Duration
	MaxReviewsServed int
}

func newServer() *server {
	newServer := server{
		packageList:      []string{},
		aliases:          make(map[string]string),
		localReviews:     make(map[string][]*androidpublisher.Review),
		newReviewsCounts: make(map[string]int),
		persistency:      &plainJSONPersistency{},
		config: ServerConfig{
			GetListTime:      30 * time.Second,
			MaxReviewsServed: 10,
		},
		alerts: AlertsContainer{
			NewReviewsAlerts: make(map[string]NewReviewsAlert),
		},
	}
	if err := newServer.initService(); err != nil {
		fmt.Println("Error initializing the service:", err.Error())
		return nil
	}

	newServer.routes()
	go newServer.getAllReviews()

	newServer.persistency.Init()
	newServer.LoadAll()

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
