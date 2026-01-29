package notifier

// RegisterAll registers all built-in notifier types to the DefaultRegistry
func RegisterAll() {
	Register("mail", wrapMail)
	Register("webhook", wrap(NewWebhook))
	Register("feishu", wrap(NewFeishu))
	Register("dingtalk", wrap(NewDingtalk))
	Register("discord", wrap(NewDiscord))
	Register("slack", wrap(NewSlack))
	Register("github", wrap(NewGitHub))
	Register("telegram", wrap(NewTelegram))
	Register("postmark", wrap(NewPostmark))
	Register("sendgrid", wrap(NewSendGrid))
	Register("ses", wrapSES)
	Register("resend", wrap(NewResend))
	Register("wxwork", wrap(NewWxWork))
	Register("googlechat", wrapGoogleChat)
	Register("healthchecks", wrapHealthchecks)
}

// wrap converts a simple factory func to the Factory signature
func wrap(fn func(*Base) *Webhook) Factory {
	return func(base *Base) (Notifier, error) {
		return fn(base), nil
	}
}

// wrapMail handles NewMail which returns (Notifier, error)
func wrapMail(base *Base) (Notifier, error) {
	return NewMail(base)
}

// wrapSES wraps NewSES to match Factory signature
func wrapSES(base *Base) (Notifier, error) {
	return NewSES(base), nil
}

// wrapGoogleChat wraps NewGoogleChat to match Factory signature
func wrapGoogleChat(base *Base) (Notifier, error) {
	return NewGoogleChat(base), nil
}

// wrapHealthchecks wraps NewHealthchecks to match Factory signature
func wrapHealthchecks(base *Base) (Notifier, error) {
	return NewHealthchecks(base), nil
}
