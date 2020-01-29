# Google Play Reviews on Mattermost

This application lets you fetch the reviews from your apps on Google Play and show them on mattermost.

## Disclaimers
This is still a work-in-progress application, with many instabilities and security risks, so please only use it on controlled environments.

## Requirements

In order to connect to Google Play, you will need to create an OAuth client.

## Usage

With the current version you can:
- Add apps to the application from mattermost (Usage: /gpreviews add app packageId)
- Set aliases for your apps (Usage: /gpreviews set alias aliasName packageId)
- List your registered apps on the application (Usage: /gpreviews list apps)
- List your most recent reviews from all your apps (Usage: /gpreviews list reviews)
- Configure an alert on a incoming webhook to tell you when there are new reivews (Usage: /gpreviews add alert newReviews name webhook packageId_or_alias frequency_in_seconds)
  - List these alerts (Usage: /gpreviews list alert newReviews)
  - Remove these alerts (Usage: /gpreviews remove alert newReviews alertName)
- Change server configuration (Usage: /gpreviews set config configField configValue)

The application on background is fetching periodically the latest reviews. This is used as cache and for alerts.

## TODO List:

- Unit tests
- Move alerts from webhooks to channel IDs
- Move persistency to KVStore
- Create alerts from mattermost
  - Configure alerts based on star rating
  - Configure alerts for reviews updates
  - Configure "do not disturb" time for alerts
- Answer reviews from mattermost
- Improve style on messages sent to mattermost
- Search reviews

## Acknowledgments
This project was started as a project for Mattermost Hackaton 2019.