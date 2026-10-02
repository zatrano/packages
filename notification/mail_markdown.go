package notification

import "github.com/zatrano/framework/v3/core/ssr"

// SetTemplate attaches an SSR Engine for template-based mail bodies.
func (m *MailManager) SetTemplate(engine ssr.Engine) {
	if m == nil {
		return
	}
	m.view = engine
}

// SetView is deprecated; use SetTemplate.
func (m *MailManager) SetView(engine ssr.Engine) {
	m.SetTemplate(engine)
}
