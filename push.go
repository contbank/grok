package grok

import "strings"

// Push notifications of the native app (workspace comunicacao, 01-catalogo-de-avisos; infrastructure in
// dispositivos-confiaveis docs 12). A service that owns a fact asks notifications for a push by publishing a
// PushNotificationRequested on the push_notification_requested topic. The service never writes text: Template
// points to the notifications catalog (title, body, route, category), and the request carries no amount, name or
// document, because the push goes through Expo, FCM and APNs.

// PushTemplate: key of the notifications push catalog. Use the constants below: a typo is a compile error instead
// of a notice dropped in production.
type PushTemplate string

const (
	// Security (cannot be turned off)
	PushTemplateNewAppAccess        PushTemplate = "NEW_APP_ACCESS"
	PushTemplateDeviceEnrolled      PushTemplate = "DEVICE_ENROLLED"
	PushTemplateDeviceRevoked       PushTemplate = "DEVICE_REVOKED"
	PushTemplatePixKeyCreated       PushTemplate = "PIX_KEY_CREATED"
	PushTemplatePixKeyDeleted       PushTemplate = "PIX_KEY_DELETED"
	PushTemplatePixKeyClaimReceived PushTemplate = "PIX_KEY_CLAIM_RECEIVED"

	// Money problems (cannot be turned off)
	PushTemplateTransferNotCompleted         PushTemplate = "TRANSFER_NOT_COMPLETED"
	PushTemplateScheduledPaymentNotCompleted PushTemplate = "SCHEDULED_PAYMENT_NOT_COMPLETED"

	// Payments
	PushTemplatePaymentPendingApproval    PushTemplate = "PAYMENT_PENDING_APPROVAL"
	PushTemplateDDAReceived               PushTemplate = "DDA_RECEIVED"
	PushTemplatePaymentApproved           PushTemplate = "PAYMENT_APPROVED"
	PushTemplatePaymentRejected           PushTemplate = "PAYMENT_REJECTED"
	PushTemplateScheduledPaymentCompleted PushTemplate = "SCHEDULED_PAYMENT_COMPLETED"
	PushTemplatePaymentsDueToday          PushTemplate = "PAYMENTS_DUE_TODAY"

	// Money in
	PushTemplateTransfersSummary PushTemplate = "TRANSFERS_SUMMARY"
	PushTemplatePixReceived      PushTemplate = "PIX_RECEIVED"
	PushTemplateTEDReceived      PushTemplate = "TED_RECEIVED"
	PushTemplateTransferReceived PushTemplate = "TRANSFER_RECEIVED"
	PushTemplateBillingsOverdue  PushTemplate = "BILLINGS_OVERDUE"
	PushTemplateInviteAccepted   PushTemplate = "INVITE_ACCEPTED"
)

// PushNotificationRequested: one notice. NotificationID is the idempotency key (same id = one notice): build it
// from the fact (e.g. "payment_approved:<payment id>"), never from a key value, name or document. Params carries
// counts for templates whose text has numbers (e.g. the daily summary).
type PushNotificationRequested struct {
	NotificationID string         `json:"notification_id"`
	Template       PushTemplate   `json:"template"`
	Params         map[string]int `json:"params,omitempty"`
	Recipients     PushRecipients `json:"recipients"`
	// Route: app screen opened on tap ("/pagar/<id>"); empty uses the template route.
	Route  string `json:"route,omitempty"`
	Source string `json:"source,omitempty"`
}

// PushRecipients: UserIDs and the admins of Account together, minus ExcludeUserIDs.
type PushRecipients struct {
	UserIDs                []string `json:"user_ids,omitempty"`
	ExcludeUserIDs         []string `json:"exclude_user_ids,omitempty"`
	ExcludeInstallationIDs []string `json:"exclude_installation_ids,omitempty"`
	// Account: company notice; authorization resolves who receives (admins, or who has Permission).
	Account *PushAccountRecipient `json:"account,omitempty"`
	// AccountContext: personal notice about a company; only gives the app the account to switch to.
	AccountContext *PushAccountRecipient `json:"account_context,omitempty"`
}

// PushAccountRecipient: Identity (CPF/CNPJ) is only used to ask authorization; it never goes to the push.
type PushAccountRecipient struct {
	AccountID             string `json:"account_id"`
	Identity              string `json:"identity"`
	Permission            string `json:"permission,omitempty"`
	IncludeBusinessAgents bool   `json:"include_business_agents,omitempty"`
}

// PushAccount builds the company recipient with the identity in digits. nil when there is no identity.
func PushAccount(accountID, identity string) *PushAccountRecipient {
	identity = OnlyDigits(identity)
	if identity == "" {
		return nil
	}
	return &PushAccountRecipient{AccountID: strings.TrimSpace(accountID), Identity: identity}
}

// PushPublisher: what RequestPush needs (MessageBrokerProducer satisfies it).
type PushPublisher interface {
	Publish(topicID string, data interface{}, attributes ...map[string]string) (string, error)
}

var _ PushPublisher = (*MessageBrokerProducer)(nil)

// RequestPush publishes the request on the push topic. An empty topic means push is off in this environment:
// nothing is published and no error is returned. A nil request is ignored.
func RequestPush(publisher PushPublisher, topic string, request *PushNotificationRequested) (string, error) {
	if publisher == nil || strings.TrimSpace(topic) == "" || request == nil {
		return "", nil
	}
	return publisher.Publish(topic, request)
}
