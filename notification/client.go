package notification

import "github.com/tbalthazar/onesignal-go"

type Client struct {
	appID           string
	onesignalClient *onesignal.Client
}

func NewClient(appID, appKey string) *Client {
	client := onesignal.NewClient(nil)
	client.AppKey = appKey

	return &Client{
		appID:           appID,
		onesignalClient: client,
	}
}
func (c *Client) NotifyActiveUsers(headings, contents map[string]string) error {
	req := &onesignal.NotificationRequest{
		AppID:            c.appID,
		Headings:         headings,
		Contents:         contents,
		IncludedSegments: []string{"Active Users"},
	}
	_, _, err := c.onesignalClient.Notifications.Create(req)
	return err
}
