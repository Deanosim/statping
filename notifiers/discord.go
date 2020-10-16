package notifiers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/statping/statping/types/failures"
	"github.com/statping/statping/types/notifications"
	"github.com/statping/statping/types/notifier"
	"github.com/statping/statping/types/null"
	"github.com/statping/statping/types/services"
	"github.com/statping/statping/utils"
	"strings"
	"time"
)

var _ notifier.Notifier = (*discord)(nil)

type discord struct {
	*notifications.Notification
}

var Discorder = &discord{&notifications.Notification{
	Method:      "discord",
	Title:       "Discord",
	Description: "Send notifications to your discord channel using discord webhooks. Insert your discord channel Webhook URL to receive notifications. Based on the <a href=\"https://discordapp.com/developers/docs/resources/Webhook\">discord webhooker API</a>.",
	Author:      "Hunter Long",
	AuthorUrl:   "https://github.com/hunterlong",
	Delay:       time.Duration(5 * time.Second),
	Icon:        "fab fa-discord",
	SuccessData: null.NewNullString(`{
  "embeds": [
    {
      "title": "{{.Service.Name}} is back up",
      "description": "Your service ['{{.Service.Name}}']({{.Service.Domain}}) is currently back online and was down for {{.Service.Downtime.Human}}.",
      "url": "{{.Service.Domain}}",
      "color": 8311585,
      "footer": {
        "icon_url": "https://avatars1.githubusercontent.com/u/61949049?s=200&v=4",
        "text": "Statping Version {{.Core.Version}}"
      },
      "author": {
      "name": "{{.Core.Name}}",
      "url": "{{.Core.Domain}}",
      "icon_url": "https://avatars1.githubusercontent.com/u/61949049?s=200&v=4"
    },
      "thumbnail": {
        "url": "https://avatars1.githubusercontent.com/u/61949049?s=200&v=4"
      },
      "fields": [
        {
          "name": "Last Online",
          "value": "{{.Service.LastOnline}}",
          "inline": true
        },
        {
          "name": "Last Offline",
          "value": "{{.Service.LastOffline}}",
          "inline": true
        },
        {
          "name": "Failures 24 Hours",
          "value": "{{.Service.FailuresLast24Hours}}",
          "inline": true
        }
      ]
    }
  ]
}`),
	FailureData: null.NewNullString(`{
  "embeds": [
    {
      "title": "Your service '{{.Service.Name}}' is failing",
      "description": "Your service ['{{.Service.Name}}']({{.Service.Domain}}) is currently offline for {{.Service.Downtime.Human}}!",
      "url": "{{.Service.Domain}}",
      "color": 13632027,
      "footer": {
        "icon_url": "https://avatars1.githubusercontent.com/u/61949049?s=200&v=4",
        "text": "Statping Version {{.Core.Version}}"
      },
      "author": {
      "name": "{{.Core.Name}}",
      "url": "{{.Core.Domain}}",
      "icon_url": "https://avatars1.githubusercontent.com/u/61949049?s=200&v=4"
    },
      "thumbnail": {
        "url": "https://avatars1.githubusercontent.com/u/61949049?s=200&v=4"
      },
      "fields": [
        {
          "name": "Downtime Start",
          "value": "{{.Failure.DowntimeAgo}}"
        },
        {
          "name": "Reason",
          "value": "{{.Failure.Issue}}",
          "inline": true
        },
        {
          "name": "Ping",
          "value": "{{.Failure.PingTime}}",
          "inline": true
        },
        {
          "name": "Failures 24 Hours",
          "value": "{{.Service.FailuresLast24Hours}}",
          "inline": true
        }
      ]
    }
  ]
}`),
	DataType:    "json",
	Limits:      60,
	Form: []notifications.NotificationForm{{
		Type:        "text",
		Title:       "discord webhooker URL",
		Placeholder: "https://discordapp.com/api/webhooks/****/*****",
		DbField:     "host",
	}}},
}

// Send will send a HTTP Post to the discord API. It accepts type: []byte
func (d *discord) sendRequest(msg string) (string, error) {
	out, _, err := utils.HttpRequest(d.Host.String, "POST", "application/json", nil, strings.NewReader(msg), time.Duration(10*time.Second), true, nil)
	return string(out), err
}

func (d *discord) Select() *notifications.Notification {
	return d.Notification
}

func (d *discord) Valid(values notifications.Values) error {
	return nil
}

// OnFailure will trigger failing service
func (d *discord) OnFailure(s services.Service, f failures.Failure) (string, error) {
	out, err := d.sendRequest(ReplaceVars(d.FailureData.String, s, f))
	return out, err
}

// OnSuccess will trigger successful service
func (d *discord) OnSuccess(s services.Service) (string, error) {
	out, err := d.sendRequest(ReplaceVars(d.SuccessData.String, s, failures.Failure{}))
	return out, err
}

// OnSave triggers when this notifier has been saved
func (d *discord) OnTest() (string, error) {
	outError := errors.New("incorrect discord URL, please confirm URL is correct")
	message := `{"content": "Testing the discord notifier"}`
	contents, _, err := utils.HttpRequest(Discorder.Host.String, "POST", "application/json", nil, bytes.NewBuffer([]byte(message)), time.Duration(10*time.Second), true, nil)
	if string(contents) == "" {
		return "", nil
	}
	var dtt discordTestJson
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(contents, &dtt); err != nil {
		return string(contents), outError
	}
	if dtt.Code == 0 {
		return string(contents), outError
	}
	return string(contents), nil
}

// OnSave will trigger when this notifier is saved
func (d *discord) OnSave() (string, error) {
	return "", nil
}

type discordTestJson struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
