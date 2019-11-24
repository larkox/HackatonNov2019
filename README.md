This application lets you fetch the reviews from your apps on Google Play and show them on mattermost.

This is still a work-in-progress application, with many instabilities and security risks, so please only use it on controlled environments.

In order to connect to Google Play, you will need to create a Service Account. You can find information on how to do this on the followin link: https://developer.android.com/google/play/developer-api?#using

Once created, you will have a json file with the credentials. Put those credentials on a safe place on the same machine as this program, and set the environment varible GOOGLE_APPLICATION_CREDENTIALS. For example:

export GOOGLE_APPLICATION_CREDENTIALS=/path/to/my/file/file_name.json

With the current version you can:
- Add apps to the application from mattermost using an outgoing webhook to "/addApp" (Usage: trigger packageId)
- Set aliases for your apps using an outgoing webhook to "/setAlias" (Usage: trigger alias packageId)
- List your registered apps on the application using an outgoing webhook to "/listApps" (Usage: trigger)
- List your most recent reviews from all your apps using an outgoing webhook to "/list" (Usage: trigger)
- Configure an alert on a incoming webhook to tell you when there are new reivews using an outgoing webhook to /addNewReviewsAlert (Usage: trigger name webhook package_name frequency_in_seconds)
  - List these alerts using an outoing webhook to "/listNewReviewsAlerts" (Usage: trigger)
  - Remove these alerts using an outgoing webhook to "/removeNewReviewsAlert" (Usage: trigger alertName)

The application on background is fetching periodically the latest reviews. This is used as cache, and also on later versions it will be used for alerts.

TODO List:
- Config file, to setup everything without messing with the code
- Create alerts from mattermost
  - Configure alerts based on star rating
  - Configure alerts for reviews updates
  - Configure "do not disturb" time for alerts
- Data persistency
- Answer reviews from mattermost
- Improve style on messages sent to mattermost
- Search reviews
- Change configuration from mattermost

This project was started as a project for Mattermost Hackaton 2019.
