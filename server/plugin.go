package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"golang.org/x/oauth2"
	"google.golang.org/api/androidpublisher/v3"
)

const (
	// GooglePlayTokenKey denotes the key on the KVStore for the GooglePlay Token
	GooglePlayTokenKey = "_googleplaytoken"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	configuration *configuration
	persistency   persistencyInt
	control       ControlUtils

	// persistent data
	// Newer reviews will always be on the lower ids of the slice
	localReviews map[string]map[string][]*androidpublisher.Review
	packageList  []PackageInfo
	aliases      map[string]map[string]string
	alerts       AlertsContainer
	token        *oauth2.Token
}

// PackageInfo stores all needed information to process each package
type PackageInfo struct {
	Name   string
	UserID string
}

//ControlUtils contains all the mutex used for flow control
type ControlUtils struct {
	reviewsMutex sync.RWMutex

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex
}

// AlertsContainer contains the maps to the different kinds of alerts
type AlertsContainer struct {
	NewReviewsAlerts map[string]map[string]*NewReviewsAlert
	NewUpdatesAlerts map[string]map[string]*NewUpdatesAlert
}

// OnActivate executes whenever the plugin is activated.
func (p *Plugin) OnActivate() error {
	if err := p.configuration.IsValid(); err != nil {
		return fmt.Errorf("error validating the configuration: %v", err)
	}

	if err := p.API.RegisterCommand(getCommand()); err != nil {
		return fmt.Errorf("failed to register command: %v", err)
	}

	p.init()

	go p.getAllReviews()
	go p.watchAlerts()

	p.persistency.Init()
	p.LoadAll()

	return nil
}

func (p *Plugin) init() {
	p.packageList = []PackageInfo{}
	p.aliases = make(map[string]map[string]string)
	p.localReviews = make(map[string]map[string][]*androidpublisher.Review)
	p.persistency = &dummyPersistency{}
	p.alerts = AlertsContainer{
		NewReviewsAlerts: make(map[string]map[string]*NewReviewsAlert),
		NewUpdatesAlerts: make(map[string]map[string]*NewUpdatesAlert),
	}
}

func (p *Plugin) getService(userID string) *androidpublisher.ReviewsService {
	config := p.getOAuthConfig(userID)

	var userInfo *GooglePlayUserInfo
	var err error

	if userInfo, err = p.getGooglePlayUserInfo(userID); err != nil {
		return nil
	}
	ctx := context.Background()
	tc := config.Client(ctx, userInfo.Token)

	service, err := androidpublisher.New(tc)
	if err != nil {
		return nil
	}

	return androidpublisher.NewReviewsService(service)
}
