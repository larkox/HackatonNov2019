package main

import "net/http"

func (s *server) routes() {
	http.HandleFunc("/list", s.serveList)
	http.HandleFunc("/listApps", s.serveAppList)
	http.HandleFunc("/listNewReviewsAlerts", s.serveListNewReviewsAlerts)
	http.HandleFunc("/setAlias", s.setAlias)
	http.HandleFunc("/addApp", s.addApp)
	http.HandleFunc("/addNewReviewsAlert", s.addNewReviewsAlert)
	http.HandleFunc("/removeNewReviewsAlert", s.removeNewReviewsAlert)
}
