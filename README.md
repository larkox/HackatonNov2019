# Google Play Reviews on Mattermost

This application lets you fetch the reviews from your apps on Google Play and show them on mattermost.

## Disclaimers
This is still a work-in-progress application, with many instabilities and security risks, so please only use it on controlled environments.

## Requirements

In order to connect to Google Play, you will need to create a Service Account. You can find information on how to do this on the followin link: https://developer.android.com/google/play/developer-api?#using

Once created, you will have a json file with the credentials. Put those credentials on a safe place on the same machine as this program, and set the environment varible GOOGLE_APPLICATION_CREDENTIALS. For example:

```
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/my/file/file_name.json
```

## Integration

You just need to create a outgoing webhook from mattermost to the / of this application. Make sure the webhook is application/json.

## Usage

With the current version you can:
- Add apps to the application from mattermost (Usage: trigger add app packageId)
- Set aliases for your apps (Usage: trigger set alias aliasName packageId)
- List your registered apps on the application (Usage: trigger list apps)
- List your most recent reviews from all your apps (Usage: trigger list reviews)
- Configure an alert on a incoming webhook to tell you when there are new reivews (Usage: trigger add alert newReviews name webhook packageId_or_alias frequency_in_seconds)
  - List these alerts (Usage: trigger list alert newReviews)
  - Remove these alerts (Usage: trigger remove alert newReviews alertName)

The application on background is fetching periodically the latest reviews. This is used as cache and for alerts.

## TODO List:

- Create alerts from mattermost
  - Configure alerts based on star rating
  - Configure alerts for reviews updates
  - Configure "do not disturb" time for alerts
- Answer reviews from mattermost
- Improve style on messages sent to mattermost
- Search reviews
- Change configuration from mattermost

## Acknowledgments
This project was started as a project for Mattermost Hackaton 2019.