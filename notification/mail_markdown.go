package notification

import "github.com/zatrano/canvas"

// SetTemplate attaches a Canvas engine for template-based mail bodies.
func (m *MailManager) SetTemplate(engine *canvas.Engine) {
	if m == nil {
		return
	}
	m.template = engine
}

// SetView is deprecated; use SetTemplate.
func (m *MailManager) SetView(engine *canvas.Engine) {
	m.SetTemplate(engine)
}
