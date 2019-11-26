package main

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/api/androidpublisher/v3"
)

type server struct {
	config      ServerConfig
	packageList []string
	aliases     map[string]string
	service     *androidpublisher.Service
	alerts      AlertsContainer
	// Newer reviews will always be on the lower ids of the slice
	localReviews map[string][]*androidpublisher.Review
	persistency  persistencyInt
	control      ControlUtils
}

//ControlUtils contains all the mutex used for flow control
type ControlUtils struct {
	reviewsMutex sync.Mutex
}

// AlertsContainer contains the maps to the different kinds of alerts
type AlertsContainer struct {
	NewReviewsAlerts map[string]NewReviewsAlert
	NewUpdatesAlerts map[string]NewUpdatesAlert
}

// ServerConfig contains all the configuration values for the server
type ServerConfig struct {
	GetListTime      int
	AlertWatcherTime int
	MaxReviewsServed int
	SaveConfig       bool
	SavePackages     bool
	SaveAliases      bool
	SaveReviews      bool
	SaveAlerts       bool
}

func newServer() *server {
	newServer := server{
		packageList:  []string{},
		aliases:      make(map[string]string),
		localReviews: make(map[string][]*androidpublisher.Review),
		persistency:  &plainJSONPersistency{},
		config:       ServerConfig{},
		alerts: AlertsContainer{
			NewReviewsAlerts: make(map[string]NewReviewsAlert),
			NewUpdatesAlerts: make(map[string]NewUpdatesAlert),
		},
	}
	if err := newServer.initService(); err != nil {
		fmt.Println("Error initializing the service:", err.Error())
		return nil
	}

	newServer.routes()
	go newServer.getAllReviews()
	go newServer.watchAlerts()

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
