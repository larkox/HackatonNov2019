package main

import (
	"google.golang.org/api/androidpublisher/v3"
)

type dummyPersistency struct{}

// Init initializes the persistency system
func (p *dummyPersistency) Init() {
}

// SavePackages stores the package list on the persistant space
func (p *dummyPersistency) SavePackages(packageList []PackageInfo) {
}

// LoadPackages loads the package list from the persistant space
func (p *dummyPersistency) LoadPackages(packageList *[]PackageInfo) bool {
	return true
}

// SaveAliases stores the package aliases on the persistant space
func (p *dummyPersistency) SaveAliases(aliases map[string]map[string]string) {
}

// LoadAliases loads the package aliases from the persistant space
func (p *dummyPersistency) LoadAliases(aliases *map[string]map[string]string) bool {
	return true
}

// SaveReviews stores the reviews on the persistant space
func (p *dummyPersistency) SaveReviews(reviews map[string]map[string][]*androidpublisher.Review) {
}

// LoadReviews loads the reviews from the persistant space
func (p *dummyPersistency) LoadReviews(reviews *map[string]map[string][]*androidpublisher.Review) bool {
	return true
}

// SaveAlerts stores the alerts on the persistant space
func (p *dummyPersistency) SaveAlerts(alerts AlertsContainer) {
}

// LoadAlerts loads the alerts from the persistant space
func (p *dummyPersistency) LoadAlerts(alerts *AlertsContainer) bool {
	return true
}
