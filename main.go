// This application lets you fetch the reviews from your apps on Google Play and show them on mattermost.
package main

import (
	"fmt"
	"net/http"
)

func main() {
	mainServer := newServer()
	if mainServer == nil {
		fmt.Println("Error initializing the server. Shutting down.")
		return
	}
	http.ListenAndServe(":8080", nil)
}
