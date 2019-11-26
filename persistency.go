package main

import (
	"google.golang.org/api/androidpublisher/v3"
)

type persistencyInt interface {
	Init()
	SaveConfig(ServerConfig)
	LoadConfig(*ServerConfig) bool
	LoadDefaultConfig(*ServerConfig) bool
	SavePackages([]string)
	LoadPackages(*[]string) bool
	SaveAliases(map[string]string)
	LoadAliases(*map[string]string) bool
	SaveReviews(map[string][]*androidpublisher.Review)
	LoadReviews(*map[string][]*androidpublisher.Review) bool
	SaveAlerts(AlertsContainer)
	LoadAlerts(*AlertsContainer) bool
}

func (s *server) SaveConfig() {
	if s.config.SaveConfig {
		s.persistency.SaveConfig(s.config)
	}
}

func (s *server) LoadConfig() {
	if !s.persistency.LoadConfig(&s.config) {
		s.LoadDefaultConfig()
	}
}

func (s *server) LoadDefaultConfig() {
	if !s.persistency.LoadDefaultConfig(&s.config) {
		s.config = ServerConfig{
			GetListTime:      30,
			MaxReviewsServed: 10,
			AlertWatcherTime: 30,
		}
	}
}

func (s *server) SavePackages() {
	if s.config.SavePackages {
		s.persistency.SavePackages(s.packageList)
	}
}

func (s *server) LoadPackages() {
	s.persistency.LoadPackages(&s.packageList)
}

func (s *server) SaveAliases() {
	if s.config.SaveAliases {
		s.persistency.SaveAliases(s.aliases)
	}
}

func (s *server) LoadAliases() {
	s.persistency.LoadAliases(&s.aliases)
}

func (s *server) SaveReviews() {
	if s.config.SaveReviews {
		s.persistency.SaveReviews(s.localReviews)
	}
}

func (s *server) LoadReviews() {
	s.persistency.LoadReviews(&s.localReviews)
}

func (s *server) SaveAlerts() {
	if s.config.SaveAlerts {
		s.persistency.SaveAlerts(s.alerts)
	}
}

func (s *server) LoadAlerts() {
	s.persistency.LoadAlerts(&s.alerts)
}

func (s *server) SaveAll() {
	s.SavePackages()
	s.SaveConfig()
	s.SaveAlerts()
	s.SaveAliases()
	s.SaveReviews()
}

func (s *server) LoadAll() {
	s.LoadPackages()
	s.LoadConfig()
	s.LoadAlerts()
	s.LoadAliases()
	s.LoadReviews()
}
