package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Status int

const (
	StatusSuccess Status = 1
	StatusInfo    Status = 2
	StatusError   Status = 3
)

type Configuration struct {
	Enabled   bool   `yaml:"enabled"`
	Url       string `yaml:"url"`
	Channel   string `yaml:"channel"`
	Username  string `yaml:"username"`
	IconUrl   string `yaml:"icon-url"`
	IconEmoji string `yaml:"icon-emoji"`
}

type mattermostAttachmentField struct {
	ShortField bool   `json:"short"`
	Title      string `json:"title"`
	Value      string `json:"value"`
}

type mattermostMessageAttachment struct {
	Fallback string                      `json:"fallback"`
	Color    string                      `json:"color"`
	Text     string                      `json:"text"`
	Fields   []mattermostAttachmentField `json:"fields"`
}

type mattermostNotification struct {
	Text        string                        `json:"text"`
	Channel     string                        `json:"channel"`
	Username    string                        `json:"username"`
	IconUrl     string                        `json:"icon_url"`
	IconEmoji   string                        `json:"icon_emoji"`
	Attachments []mattermostMessageAttachment `json:"attachments"`
}

///////////////////////////////////////////////////////////////////////////////

type Notifier struct {
	Configuration Configuration
}

func (notifier *Notifier) Notify(status Status, dumperName string, message string) {
	if !notifier.Configuration.Enabled {
		return
	}

	attachment := mattermostMessageAttachment{
		Fallback: fmt.Sprintf("%s: %s", dumperName, message),
		Color:    statusColor(status),
		Text:     message,
		Fields: []mattermostAttachmentField{
			{
				Title: "Backup",
				Value: dumperName,
			},
		},
	}

	notification := mattermostNotification{
		Channel:     notifier.Configuration.Channel,
		Username:    notifier.Configuration.Username,
		IconUrl:     notifier.Configuration.IconUrl,
		IconEmoji:   notifier.Configuration.IconEmoji,
		Attachments: []mattermostMessageAttachment{attachment},
	}

	notificationJson, err := json.Marshal(notification)
	if err != nil {
		log.Errorf("unable to marshal notification: %s", err)
		return
	}

	response, err := http.Post(notifier.Configuration.Url, "application/json", bytes.NewBuffer(notificationJson))
	if err != nil {
		log.Errorf("unable to send notification: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Errorf("notification status: %d", response.StatusCode)
	}
}

///////////////////////////////////////////////////////////////////////////////

func statusColor(status Status) string {
	switch status {
	case StatusSuccess:
		return "#00AA00"
	case StatusInfo:
		return "#0000AA"
	case StatusError:
		return "#AA0000"
	default:
		return "#888888"
	}
}
