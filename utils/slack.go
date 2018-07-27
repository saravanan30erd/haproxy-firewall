package utils

import (
	"github.com/ashwanthkumar/slack-go-webhook"
	"os"
)

func SlackNotify(message string) {
	host, _ := os.Hostname()
	attachment1 := slack.Attachment{}
	attachment1.AddField(slack.Field{Title: "Host", Value: host}).AddField(slack.Field{Title: "Message", Value: message})
	payload := slack.Payload{
		Text:        "Notification from FE-DOS-Signin",
		Username:    "FE-DOS-Signin",
		IconEmoji:   ":red_circle:",
		Attachments: []slack.Attachment{attachment1},
	}
	slackUrl := os.Getenv("SLACK_URL")
	err := slack.Send(slackUrl, "", payload)
	if len(err) > 0 {
		Errorlog.Println("Slack error: ", err)
	}
}
