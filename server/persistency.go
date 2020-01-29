package main

import (
	"google.golang.org/api/androidpublisher/v3"
)

type persistencyInt interface {
	Init()
	SavePackages([]PackageInfo)
	LoadPackages(*[]PackageInfo) bool
	SaveAliases(map[string]map[string]string)
	LoadAliases(*map[string]map[string]string) bool
	SaveReviews(map[string]map[string][]*androidpublisher.Review)
	LoadReviews(*map[string]map[string][]*androidpublisher.Review) bool
	SaveAlerts(AlertsContainer)
	LoadAlerts(*AlertsContainer) bool
}

// SavePackages stores the package list on the persistant space
func (p *Plugin) SavePackages() {
	p.persistency.SavePackages(p.packageList)
}

// LoadPackages loads the package list from the persistant space
func (p *Plugin) LoadPackages() {
	p.persistency.LoadPackages(&p.packageList)
}

// SaveAliases stores the package aliases on the persistant space
func (p *Plugin) SaveAliases() {
	p.persistency.SaveAliases(p.aliases)
}

// LoadAliases loads the package aliases from the persistant space
func (p *Plugin) LoadAliases() {
	p.persistency.LoadAliases(&p.aliases)
}

// SaveReviews stores the reviews on the persistant space
func (p *Plugin) SaveReviews() {
	p.persistency.SaveReviews(p.localReviews)
}

// LoadReviews loads the reviews from the persistant space
func (p *Plugin) LoadReviews() {
	p.persistency.LoadReviews(&p.localReviews)
}

// SaveAlerts stores the alerts on the persistant space
func (p *Plugin) SaveAlerts() {
	p.persistency.SaveAlerts(p.alerts)
}

// LoadAlerts loads the alerts from the persistant space
func (p *Plugin) LoadAlerts() {
	p.persistency.LoadAlerts(&p.alerts)
}

// SaveAll stores all information (packages, aliases, alerts and reviews) on the persistant space
func (p *Plugin) SaveAll() {
	p.SavePackages()
	p.SaveAlerts()
	p.SaveAliases()
	p.SaveReviews()
}

// LoadAll loads all information (packages, aliases, alerts and reviews) from the persistant space
func (p *Plugin) LoadAll() {
	p.LoadPackages()
	p.LoadAlerts()
	p.LoadAliases()
	p.LoadReviews()
}
