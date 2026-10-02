package notification

import "github.com/zatrano/packages/template"

// SetTemplate attaches a template Engine for template-based mail bodies.
func (m *MailManager) SetTemplate(engine template.Engine) {
	if m == nil {
		return
	}
	m.view = engine
}

// SetView is deprecated; use SetTemplate.
func (m *MailManager) SetView(engine template.Engine) {
	m.SetTemplate(engine)
}
