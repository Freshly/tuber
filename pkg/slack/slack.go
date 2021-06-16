package slack

import (
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

type Client struct {
	client          *slack.Client
	enabled         bool
	catchAllChannel string
}

func New(key string, enabled bool, catchAllChannel string) *Client {
	return &Client{
		client:          slack.New(key),
		enabled:         enabled,
		catchAllChannel: catchAllChannel,
	}
}

func (c *Client) Message(logger *zap.Logger, message string, channels ...string) {
	messageLogger := logger.With(zap.String("slackMessage", message), zap.Strings("slackChannels", channels))
	messageLogger.Debug("slack message triggered")
	var presentChannels []string
	for _, ch := range channels {
		if ch != "" {
			presentChannels = append(presentChannels, ch)
		}
	}
	if c.enabled {
		if len(presentChannels) == 0 {
			c.send(messageLogger, c.catchAllChannel, message)
			return
		}
		for _, channel := range presentChannels {
			c.send(messageLogger, channel, message)
			return
		}
	}

	messageLogger.Debug("slack message would have sent but slack is not enabled")
}

func (c *Client) send(logger *zap.Logger, channel string, message string) {
	channelLogger := logger.With(zap.String("slackChannel", channel))
	channelLogger.Debug("sending slack message")
	_, _, err := c.client.PostMessage(channel, slack.MsgOptionText(message, false))
	if err != nil {
		if err.Error() == "channel_not_found" {
			channelLogger.Error("channel not found, check configured channel and ensure tuber is a member", zap.Error(err))
			return
		}
		channelLogger.Error("error sending slack message", zap.Error(err))
		return
	}

	channelLogger.Debug("posted slack message without error")
}
