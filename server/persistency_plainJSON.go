package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"google.golang.org/api/androidpublisher/v3"
)

type plainJSONPersistency struct {
	defaultConfigFilename string
	configFilename        string
	packagesFilename      string
	aliasesFilename       string
	reviewsFilename       string
	alertsFilename        string
}

func (p *plainJSONPersistency) Init() {
	p.defaultConfigFilename = "default_config.json"
	p.configFilename = "data/config.json"
	p.packagesFilename = "data/packages.json"
	p.aliasesFilename = "data/aliases.json"
	p.reviewsFilename = "data/reviews.json"
	p.alertsFilename = "data/alerts.json"
}

func (p *plainJSONPersistency) SavePackages(packageList []PackageInfo) {
	data, err := json.MarshalIndent(packageList, "", "    ")
	if err != nil {
		fmt.Println("Error saving packages: " + err.Error())
		return
	}

	err = ioutil.WriteFile(p.packagesFilename, data, 0644)
	if err != nil {
		fmt.Println("Error saving packages: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadPackages(packageList *[]PackageInfo) bool {
	data, err := ioutil.ReadFile(p.packagesFilename)
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

func (p *plainJSONPersistency) SaveAliases(aliases map[string]map[string]string) {
	data, err := json.MarshalIndent(aliases, "", "    ")
	if err != nil {
		fmt.Println("Error saving aliases: " + err.Error())
		return
	}

	err = ioutil.WriteFile(p.aliasesFilename, data, 0644)
	if err != nil {
		fmt.Println("Error saving aliases: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadAliases(aliases *map[string]map[string]string) bool {
	data, err := ioutil.ReadFile(p.aliasesFilename)
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

func (p *plainJSONPersistency) SaveReviews(reviews map[string]map[string][]*androidpublisher.Review) {
	data, err := json.MarshalIndent(reviews, "", "    ")
	if err != nil {
		fmt.Println("Error saving reviews: " + err.Error())
		return
	}

	err = ioutil.WriteFile(p.reviewsFilename, data, 0644)
	if err != nil {
		fmt.Println("Error saving reviews: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadReviews(reviews *map[string]map[string][]*androidpublisher.Review) bool {
	data, err := ioutil.ReadFile(p.reviewsFilename)
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

	err = ioutil.WriteFile(p.alertsFilename, data, 0644)
	if err != nil {
		fmt.Println("Error saving alerts: " + err.Error())
	}
}

func (p *plainJSONPersistency) LoadAlerts(alerts *AlertsContainer) bool {
	data, err := ioutil.ReadFile(p.alertsFilename)
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
