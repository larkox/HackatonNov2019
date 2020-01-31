package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if err := p.configuration.IsValid(); err != nil {
		http.Error(w, "This plugin is not configured.", http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch path := r.URL.Path; path {
	case "/oauth/connect":
		p.connectUserToGooglePlay(w, r)
	case "/oauth/complete":
		p.completeConnectUserToGooglePlay(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) connectUserToGooglePlay(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	conf := p.getOAuthConfig(userID)

	state := fmt.Sprintf("%v_%v", model.NewId()[0:15], userID)

	p.API.KVSet(state, []byte(state))

	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	http.Redirect(w, r, url, http.StatusFound)
}

func (p *Plugin) getOAuthConfig(userID string) *oauth2.Config {
	apiConfig := p.API.GetConfig()
	siteURL := *apiConfig.ServiceSettings.SiteURL
	if i := len(siteURL); i > 0 && siteURL[i-1] == '/' {
		siteURL = siteURL[:i-1]
	}
	redirectURL := siteURL + "/plugins/com.mattermost.google-play-reviews/oauth/complete"

	pluginConfig := p.getConfiguration()
	conf := &oauth2.Config{
		ClientID:     pluginConfig.GooglePlayOAuthClientID,
		ClientSecret: pluginConfig.GooglePlayOAuthClientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/androidpublisher",
		},
		Endpoint: google.Endpoint,
	}

	return conf
}

func (p *Plugin) completeConnectUserToGooglePlay(w http.ResponseWriter, r *http.Request) {
	authedUserID := r.Header.Get("Mattermost-User-ID")
	if authedUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	conf := p.getOAuthConfig(authedUserID)

	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")

	if storedState, err := p.API.KVGet(state); err != nil {
		fmt.Println(err.Error())
		http.Error(w, "missing stored state", http.StatusBadRequest)
		return
	} else if string(storedState) != state {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	userID := strings.Split(state, "_")[1]

	p.API.KVDelete(state)

	if userID != authedUserID {
		http.Error(w, "Not authorized, incorrect user", http.StatusUnauthorized)
		return
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo := &GooglePlayUserInfo{
		UserID: userID,
		Token:  tok,
	}

	if err := p.storeGooglePlayUserInfo(userInfo); err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Unable to connect user to GooglePlay:"+err.Error(), http.StatusInternalServerError)
		return
	}

	html := `
		<!DOCTYPE html>
		<html>
			<head>
				<script>
					window.close();
				</script>
			</head>
			<body>
				<p>Completed connecting to GooglePlay. Please close this window.</p>
			</body>
		</html>
		`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (p *Plugin) getGooglePlayUserInfo(userID string) (*GooglePlayUserInfo, error) {
	config := p.getConfiguration()

	var userInfo GooglePlayUserInfo

	if infoBytes, err := p.API.KVGet(userID + GooglePlayTokenKey); err != nil || infoBytes == nil {
		return nil, fmt.Errorf("must connect user account to GooglePlay first")
	} else if err := json.Unmarshal(infoBytes, &userInfo); err != nil {
		return nil, fmt.Errorf("unable to parse token")
	}

	unencryptedToken, err := decrypt([]byte(config.EncryptionKey), userInfo.Token.AccessToken)
	if err != nil {
		mlog.Error(err.Error())
		return nil, fmt.Errorf("unable to decrypt access token")
	}

	userInfo.Token.AccessToken = unencryptedToken

	return &userInfo, nil
}

func (p *Plugin) storeGooglePlayUserInfo(info *GooglePlayUserInfo) error {
	config := p.getConfiguration()

	encryptedToken, err := encrypt([]byte(config.EncryptionKey), info.Token.AccessToken)
	if err != nil {
		return err
	}

	info.Token.AccessToken = encryptedToken

	jsonInfo, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if err := p.API.KVSet(info.UserID+GooglePlayTokenKey, jsonInfo); err != nil {
		return err
	}

	return nil
}

// GooglePlayUserInfo stores important user information to save on the KVStore
type GooglePlayUserInfo struct {
	UserID string
	Token  *oauth2.Token
}
