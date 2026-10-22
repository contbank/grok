package grok

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordingPushPublisher struct {
	topic string
	data  interface{}
	fail  bool
}

func (p *recordingPushPublisher) Publish(topicID string, data interface{}, _ ...map[string]string) (string, error) {
	if p.fail {
		return "", errors.New("sns unavailable")
	}
	p.topic, p.data = topicID, data
	return "msg-1", nil
}

// The JSON is the contract read by notifications: field names must not change.
func TestPushNotificationRequestedJSON(t *testing.T) {
	request := &PushNotificationRequested{
		NotificationID: "payment_approved:6ad0",
		Template:       PushTemplatePaymentApproved,
		Params:         map[string]int{"pix": 2},
		Route:          "/pagar/6ad0",
		Source:         "payments",
		Recipients: PushRecipients{
			UserIDs:        []string{"joao"},
			ExcludeUserIDs: []string{"maria"},
			AccountContext: PushAccount("acc", "30.777.482/0001-51"),
		},
	}
	raw, err := json.Marshal(request)
	require.NoError(t, err)
	require.JSONEq(t, `{"notification_id":"payment_approved:6ad0","template":"PAYMENT_APPROVED","params":{"pix":2},
		"route":"/pagar/6ad0","source":"payments","recipients":{"user_ids":["joao"],"exclude_user_ids":["maria"],
		"account_context":{"account_id":"acc","identity":"30777482000151"}}}`, string(raw))
}

// Amount in cents and texts for the complete text kept in the notification center (never on the push).
func TestPushNotificationRequestedDetailsJSON(t *testing.T) {
	request := &PushNotificationRequested{
		NotificationID: "pix_received:E123",
		Template:       PushTemplatePixReceived,
		Params:         map[string]int{PushParamAmountCents: 20000},
		Details:        map[string]string{PushDetailCounterparty: "CLIENTE Z"},
	}
	raw, err := json.Marshal(request)
	require.NoError(t, err)
	require.JSONEq(t, `{"notification_id":"pix_received:E123","template":"PIX_RECEIVED","params":{"amount_cents":20000},
		"details":{"counterparty":"CLIENTE Z"},"recipients":{}}`, string(raw))
}

func TestPushAccount(t *testing.T) {
	require.Equal(t, &PushAccountRecipient{AccountID: "acc", Identity: "30777482000151"}, PushAccount(" acc ", "30.777.482/0001-51"))
	require.Nil(t, PushAccount("acc", ""))
}

func TestRequestPush(t *testing.T) {
	publisher := &recordingPushPublisher{}
	request := &PushNotificationRequested{NotificationID: "x", Template: PushTemplateNewAppAccess}

	id, err := RequestPush(publisher, "push_notification_requested", request)
	require.NoError(t, err)
	require.Equal(t, "msg-1", id)
	require.Equal(t, "push_notification_requested", publisher.topic)
	require.Equal(t, request, publisher.data)

	other := &recordingPushPublisher{}
	_, err = RequestPush(other, "", request)
	require.NoError(t, err, "push off")
	require.Empty(t, other.topic)
	_, err = RequestPush(other, "t", nil)
	require.NoError(t, err)

	_, err = RequestPush(&recordingPushPublisher{fail: true}, "t", request)
	require.Error(t, err)
}
