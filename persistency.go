package main

import (
	"google.golang.org/api/androidpublisher/v3"
)

type persistencyInt interface {
	Init()
	SaveConfig(ServerConfig)
	LoadConfig(*ServerConfig)
	SavePackages([]string)
	LoadPackages(*[]string)
	SaveAliases(map[string]string)
	LoadAliases(*map[string]string)
	SaveReviews(map[string][]*androidpublisher.Review)
	LoadReviews(*map[string][]*androidpublisher.Review)
	SaveAlerts(AlertsContainer)
	LoadAlerts(*AlertsContainer)
	SaveAll(server)
	LoadAll(*server)
}

func (s *server) SaveConfig() {
	s.persistency.SaveConfig(s.config)
}

func (s *server) LoadConfig() {
	s.persistency.LoadConfig(&s.config)
}

func (s *server) SavePackages() {
	s.persistency.SavePackages(s.packageList)
}

func (s *server) LoadPackages() {
	s.persistency.LoadPackages(&s.packageList)
}

func (s *server) SaveAliases() {
	s.persistency.SaveAliases(s.aliases)
}

func (s *server) LoadAliases() {
	s.persistency.LoadAliases(&s.aliases)
}

func (s *server) SaveReviews() {
	s.persistency.SaveReviews(s.localReviews)
}

func (s *server) LoadReviews() {
	s.persistency.LoadReviews(&s.localReviews)
}

func (s *server) SaveAlerts() {
	s.persistency.SaveAlerts(s.alerts)
}

func (s *server) LoadAlerts() {
	s.persistency.LoadAlerts(&s.alerts)
}

func (s *server) SaveAll() {
	s.persistency.SaveAll(*s)
}

func (s *server) LoadAll() {
	s.persistency.LoadAll(s)
}
