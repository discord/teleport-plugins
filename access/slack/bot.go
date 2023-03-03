/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gravitational/teleport-plugins/access/common"
	"github.com/gravitational/teleport-plugins/lib"
	pd "github.com/gravitational/teleport-plugins/lib/plugindata"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/trace"

	"github.com/go-resty/resty/v2"
)

const slackMaxConns = 100
const slackHTTPTimeout = 10 * time.Second

// SlackBot is a slack client that works with AccessRequest.
// It's responsible for formatting and posting a message on Slack when an
// action occurs with an access request: a new request popped up, or a
// request is processed/updated.
type SlackBot struct {
	client      *resty.Client
	clusterName string
	webProxyURL *url.URL
}

// onAfterResponseSlack resty error function for Slack
func onAfterResponseSlack(_ *resty.Client, resp *resty.Response) error {
	if !resp.IsSuccess() {
		return trace.Errorf("slack api returned unexpected code %v", resp.StatusCode())
	}

	var result SlackResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func (b SlackBot) CheckHealth(ctx context.Context) error {
	return nil
}

// Broadcast posts request info to Slack with action buttons.
func (b SlackBot) Broadcast(ctx context.Context, recipients []common.Recipient, reqID string, reqData pd.AccessRequestData) (common.SentMessages, error) {
	var data common.SentMessages
	var errors []error

	for _, recipient := range recipients {
		var result ChatMsgResponse
		_, err := b.client.NewRequest().
			SetContext(ctx).
			SetBody(SlackMsg{Text: b.slackMsgSections(reqID, reqData)}).
			SetResult(&result).
			Post(recipient.ID + "?wait=true")
		if err != nil {
			errors = append(errors, trace.Wrap(err))
			continue
		}
		data = append(data, common.MessageData{ChannelID: recipient.ID, MessageID: result.Id})
	}

	return data, trace.NewAggregate(errors...)
}

func (b SlackBot) PostReviewReply(ctx context.Context, channelID, timestamp string, review types.AccessReview) error {
	text, err := common.MsgReview(review)
	if err != nil {
		return trace.Wrap(err)
	}

	_, err = b.client.NewRequest().
		SetContext(ctx).
		SetBody(SlackMsg{Text: text}).
		Patch(channelID + "/messages/" + timestamp)
	return trace.Wrap(err)
}

// Expire updates request's Slack post with EXPIRED status and removes action buttons.
func (b SlackBot) UpdateMessages(ctx context.Context, reqID string, reqData pd.AccessRequestData, slackData common.SentMessages, reviews []types.AccessReview) error {
	var errors []error
	for _, msg := range slackData {
		_, err := b.client.NewRequest().
			SetContext(ctx).
			SetBody(SlackMsg{Text: b.slackMsgSections(reqID, reqData)}).
			Patch(msg.ChannelID + "/messages/" + msg.MessageID)
		if err != nil {
			switch err.Error() {
			case "message_not_found":
				err = trace.Wrap(err, "cannot find message with timestamp %s in channel %s", msg.MessageID, msg.ChannelID)
			default:
				err = trace.Wrap(err)
			}
			errors = append(errors, trace.Wrap(err))
		}
	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func (b SlackBot) FetchRecipient(ctx context.Context, recipient string) (*common.Recipient, error) {
	if lib.IsEmail(recipient) {
		return nil, trace.Errorf("Cannot handle individual user recipients")
	}
	// TODO: check if channel exists ?
	return &common.Recipient{
		Name: recipient,
		ID:   recipient,
		Kind: "Channel",
		Data: nil,
	}, nil
}

// msgSection builds a Slack message section (obeys markdown).
func (b SlackBot) slackMsgSections(reqID string, reqData pd.AccessRequestData) string {
	fields := common.MsgFields(reqID, reqData, b.clusterName, b.webProxyURL)
	statusText := common.MsgStatusText(reqData.ResolutionTag, reqData.ResolutionReason)

	return fmt.Sprintf("You have a new Role Request:\n%s\n%s", fields, statusText)
}
