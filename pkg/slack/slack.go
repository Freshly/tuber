package slack

import (
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

var MsgOptionDisableLinkUnfurl = slack.MsgOptionDisableLinkUnfurl

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

func (c *Client) Message(logger *zap.Logger, message string, channel string, opts ...slack.MsgOption) {
	messageLogger := logger.With(zap.String("slackMessage", message), zap.String("slackChannel", channel))
	messageLogger.Debug("slack message triggered")

	if !c.enabled {
		messageLogger.Debug("slack message would have sent but slack is not enabled")
		return
	}

	if channel == "" {
		c.send(messageLogger, c.catchAllChannel, message, opts...)
		return
	}

	c.send(messageLogger, channel, message)
}

func (c *Client) send(logger *zap.Logger, channel string, message string, opts ...slack.MsgOption) {
	channelLogger := logger.With(zap.String("slackChannel", channel))
	channelLogger.Debug("sending slack message")

	opts = append(opts, slack.MsgOptionText(message, false))

	_, _, err := c.client.PostMessage(channel, opts...)
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
