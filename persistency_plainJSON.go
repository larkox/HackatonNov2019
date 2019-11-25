package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"google.golang.org/api/androidpublisher/v3"
)

type plainJSONPersistency struct {
	DEFAULT_CONFIG_FILENAME string
	CONFIG_FILENAME         string
	PACKAGES_FILENAME       string
	ALIASES_FILENAME        string
	REVIEWS_FILENAME        string
	ALERTS_FILENAME         string
}

func (p *plainJSONPersistency) Init() {
	p.DEFAULT_CONFIG_FILENAME = "default_config.json"
	p.CONFIG_FILENAME = "data/config.json"
	p.PACKAGES_FILENAME = "data/packages.json"
	p.ALIASES_FILENAME = "data/aliases.json"
	p.REVIEWS_FILENAME = "data/reviews.json"
	p.ALERTS_FILENAME = "data/alerts.json"
}

func (p *plainJSONPersistency) SaveConfig(config ServerConfig) {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		fmt.Println("Error saving config: " + err.Error())
		return
	}

	err = ioutil.WriteFile(p.CONFIG_FILENAME, data, 0644)
	if err != nil {
		fmt.Println("Error saving config: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadConfig(config *ServerConfig) bool {
	data, err := ioutil.ReadFile(p.CONFIG_FILENAME)
	if err != nil {
		fmt.Println("Error loading config: " + err.Error())
		return false
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		fmt.Println("Error loading config: " + err.Error())
		return false
	}
	return true
}

func (p *plainJSONPersistency) LoadDefaultConfig(config *ServerConfig) bool {
	data, err := ioutil.ReadFile(p.DEFAULT_CONFIG_FILENAME)
	if err != nil {
		fmt.Println("Error loading default config: " + err.Error())
		return false
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		fmt.Println("Error loading default config: " + err.Error())
		return false
	}
	return true
}

func (p *plainJSONPersistency) SavePackages(packageList []string) {
	data, err := json.MarshalIndent(packageList, "", "    ")
	if err != nil {
		fmt.Println("Error saving packages: " + err.Error())
		return
	}

	err = ioutil.WriteFile(p.PACKAGES_FILENAME, data, 0644)
	if err != nil {
		fmt.Println("Error saving packages: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadPackages(packageList *[]string) bool {
	data, err := ioutil.ReadFile(p.PACKAGES_FILENAME)
	if err != nil {
		fmt.Println("Error loading packages: " + err.Error())
		return false
	}

	err = json.Unmarshal(data, packageList)
	if err != nil {
		fmt.Println("Error loading packages: " + err.Error())
		return false
	}
	return true
}

func (p *plainJSONPersistency) SaveAliases(aliases map[string]string) {
	data, err := json.MarshalIndent(aliases, "", "    ")
	if err != nil {
		fmt.Println("Error saving aliases: " + err.Error())
		return
	}

	err = ioutil.WriteFile(p.ALIASES_FILENAME, data, 0644)
	if err != nil {
		fmt.Println("Error saving aliases: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadAliases(aliases *map[string]string) bool {
	data, err := ioutil.ReadFile(p.ALIASES_FILENAME)
	if err != nil {
		fmt.Println("Error loading aliases: " + err.Error())
		return false
	}

	err = json.Unmarshal(data, aliases)
	if err != nil {
		fmt.Println("Error loading aliases: " + err.Error())
		return false
	}
	return true
}

func (p *plainJSONPersistency) SaveReviews(reviews map[string][]*androidpublisher.Review) {
	data, err := json.MarshalIndent(reviews, "", "    ")
	if err != nil {
		fmt.Println("Error saving reviews: " + err.Error())
		return
	}

	err = ioutil.WriteFile(p.REVIEWS_FILENAME, data, 0644)
	if err != nil {
		fmt.Println("Error saving reviews: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadReviews(reviews *map[string][]*androidpublisher.Review) bool {
	data, err := ioutil.ReadFile(p.REVIEWS_FILENAME)
	if err != nil {
		fmt.Println("Error loading reviews: " + err.Error())
		return false
	}

	err = json.Unmarshal(data, reviews)
	if err != nil {
		fmt.Println("Error loading reviews: " + err.Error())
		return false
	}
	return true
}

func (p *plainJSONPersistency) SaveAlerts(alerts AlertsContainer) {
	data, err := json.MarshalIndent(alerts, "", "    ")
	if err != nil {
		fmt.Println("Error saving alerts: " + err.Error())
		return
	}

	err = ioutil.WriteFile(p.ALERTS_FILENAME, data, 0644)
	if err != nil {
		fmt.Println("Error saving alerts: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadAlerts(alerts *AlertsContainer) bool {
	data, err := ioutil.ReadFile(p.ALERTS_FILENAME)
	if err != nil {
		fmt.Println("Error loading alerts: " + err.Error())
		return false
	}

	err = json.Unmarshal(data, alerts)
	if err != nil {
		fmt.Println("Error loading alerts: " + err.Error())
		return false
	}
	return true
}

func (p *plainJSONPersistency) SaveAll(s server) {
	p.SavePackages(s.packageList)
	p.SaveConfig(s.config)
	p.SaveAlerts(s.alerts)
	p.SaveAliases(s.aliases)
	p.SaveReviews(s.localReviews)
}

func (p *plainJSONPersistency) LoadAll(s *server) bool {
	ok := p.LoadPackages(&s.packageList)
	ok = ok && p.LoadConfig(&s.config)
	ok = ok && p.LoadAlerts(&s.alerts)
	ok = ok && p.LoadAliases(&s.aliases)
	ok = ok && p.LoadReviews(&s.localReviews)
	return ok
}
