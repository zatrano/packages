package markdown

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

var (
	reHeading    = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
	reBold       = regexp.MustCompile(`\*\*(.+?)\*\*|__(.+?)__`)
	reItalic     = regexp.MustCompile(`\*(.+?)\*|_(.+?)_`)
	reItalicStar = regexp.MustCompile(`(^|[^A-Za-z0-9])\*([A-Za-z0-9](?:[^*\n]*[A-Za-z0-9])?)\*([^A-Za-z0-9]|$)`)
	reItalicU    = regexp.MustCompile(`(^|[^A-Za-z0-9])_([^_\n]+?)_([^A-Za-z0-9]|$)`)
	reCode       = regexp.MustCompile("`([^`]+)`")
	reLink       = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reInlineSlot = regexp.MustCompile("\x00INL(\\d+)\x00")
	reHR         = regexp.MustCompile(`(?m)^-{3,}$`)
	reUL         = regexp.MustCompile(`(?m)^(?:[-*]\s+.+\n?)+`)
	reOL         = regexp.MustCompile(`(?m)^(?:\d+\.\s+.+\n?)+`)
	reQuote      = regexp.MustCompile(`(?m)^>\s?(.+)$`)
	reFence      = regexp.MustCompile("(?s)```([a-zA-Z0-9_-]*)\\n(.*?)```")
	reTableRow   = regexp.MustCompile(`(?m)^\|.+\|$`)
)

// ToHTML converts a small Markdown subset to HTML.
func ToHTML(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	codeBlocks := make([]string, 0)
	input = reFence.ReplaceAllStringFunc(input, func(m string) string {
		match := reFence.FindStringSubmatch(m)
		if len(match) != 3 {
			return html.EscapeString(m)
		}
		lang := strings.TrimSpace(match[1])
		body := strings.TrimRight(match[2], "\n")
		class := ""
		if lang != "" {
			class = ` class="language-` + html.EscapeString(lang) + `"`
		}
		htmlBlock := "<pre><code" + class + ">" + html.EscapeString(body) + "</code></pre>"
		codeBlocks = append(codeBlocks, htmlBlock)
		return "\n\n%%CODEBLOCK" + strconv.Itoa(len(codeBlocks)-1) + "%%\n\n"
	})

	blocks := strings.Split(input, "\n\n")
	parts := make([]string, 0, len(blocks))
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		if strings.HasPrefix(block, "%%CODEBLOCK") && strings.HasSuffix(block, "%%") {
			idx := strings.TrimSuffix(strings.TrimPrefix(block, "%%CODEBLOCK"), "%%")
			if n, err := strconv.Atoi(idx); err == nil && n >= 0 && n < len(codeBlocks) {
				parts = append(parts, codeBlocks[n])
				continue
			}
		}
		parts = append(parts, renderBlock(block))
	}
	return strings.Join(parts, "\n")
}

// ToText strips Markdown markers for a plain-text fallback.
func ToText(input string) string {
	out := input
	out = reFence.ReplaceAllString(out, "$2")
	out = reHeading.ReplaceAllString(out, "$2")
	out = reBold.ReplaceAllString(out, "$1$2")
	out = reItalic.ReplaceAllString(out, "$1$2")
	out = reCode.ReplaceAllString(out, "$1")
	out = reLink.ReplaceAllString(out, "$1 ($2)")
	out = reHR.ReplaceAllString(out, "")
	out = reQuote.ReplaceAllString(out, "$1")
	out = regexp.MustCompile(`(?m)^[-*]\s+`).ReplaceAllString(out, "• ")
	out = regexp.MustCompile(`(?m)^\d+\.\s+`).ReplaceAllString(out, "")
	return strings.TrimSpace(out)
}

func renderBlock(block string) string {
	if reHR.MatchString(block) && !strings.Contains(block, "\n") {
		return "<hr>"
	}
	if m := reHeading.FindStringSubmatch(block); len(m) == 3 && !strings.Contains(block, "\n") {
		level := len(m[1])
		return "<h" + strconv.Itoa(level) + ">" + inline(m[2]) + "</h" + strconv.Itoa(level) + ">"
	}
	if isTableBlock(block) {
		return renderTable(block)
	}
	if reUL.MatchString(block+"\n") || strings.HasPrefix(block, "- ") || strings.HasPrefix(block, "* ") {
		lines := strings.Split(block, "\n")
		var b strings.Builder
		b.WriteString("<ul>")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			if line == "" {
				continue
			}
			b.WriteString("<li>")
			b.WriteString(inline(line))
			b.WriteString("</li>")
		}
		b.WriteString("</ul>")
		return b.String()
	}
	if reOL.MatchString(block+"\n") || regexp.MustCompile(`^\d+\.\s+`).MatchString(block) {
		lines := strings.Split(block, "\n")
		var b strings.Builder
		b.WriteString("<ol>")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			line = regexp.MustCompile(`^\d+\.\s+`).ReplaceAllString(line, "")
			if line == "" {
				continue
			}
			b.WriteString("<li>")
			b.WriteString(inline(line))
			b.WriteString("</li>")
		}
		b.WriteString("</ol>")
		return b.String()
	}
	if strings.HasPrefix(block, "> ") || strings.HasPrefix(block, ">") {
		lines := strings.Split(block, "\n")
		var b strings.Builder
		b.WriteString("<blockquote><p>")
		parts := make([]string, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "> "), ">"))
			parts = append(parts, inline(line))
		}
		b.WriteString(strings.Join(parts, "<br>"))
		b.WriteString("</p></blockquote>")
		return b.String()
	}
	escapedLines := make([]string, 0)
	for _, line := range strings.Split(block, "\n") {
		escapedLines = append(escapedLines, inline(line))
	}
	return "<p>" + strings.Join(escapedLines, "<br>") + "</p>"
}

func isTableBlock(block string) bool {
	lines := strings.Split(block, "\n")
	if len(lines) < 2 {
		return false
	}
	for _, line := range lines {
		if !reTableRow.MatchString(strings.TrimSpace(line)) {
			return false
		}
	}
	return strings.Contains(lines[1], "---")
}

func renderTable(block string) string {
	lines := strings.Split(block, "\n")
	var b strings.Builder
	b.WriteString("<table>")
	bodyOpened := false
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if i == 1 && strings.Contains(line, "---") {
			continue
		}
		cells := splitTableRow(line)
		if i == 0 {
			b.WriteString("<thead><tr>")
			for _, cell := range cells {
				b.WriteString("<th>" + inline(strings.TrimSpace(cell)) + "</th>")
			}
			b.WriteString("</tr></thead>")
			continue
		}
		if !bodyOpened {
			b.WriteString("<tbody>")
			bodyOpened = true
		}
		b.WriteString("<tr>")
		for _, cell := range cells {
			b.WriteString("<td>" + inline(strings.TrimSpace(cell)) + "</td>")
		}
		b.WriteString("</tr>")
	}
	if bodyOpened {
		b.WriteString("</tbody>")
	}
	b.WriteString("</table>")
	return b.String()
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	return strings.Split(line, "|")
}

func inline(input string) string {
	slots := make([]string, 0)
	put := func(fragment string) string {
		slots = append(slots, fragment)
		return "\x00INL" + strconv.Itoa(len(slots)-1) + "\x00"
	}

	raw := reCode.ReplaceAllStringFunc(input, func(m string) string {
		match := reCode.FindStringSubmatch(m)
		if len(match) != 2 {
			return m
		}
		return put("<code>" + html.EscapeString(match[1]) + "</code>")
	})
	raw = reLink.ReplaceAllStringFunc(raw, func(m string) string {
		match := reLink.FindStringSubmatch(m)
		if len(match) != 3 {
			return m
		}
		label := applyEmphasis(match[1], put)
		return put(`<a href="` + html.EscapeString(match[2]) + `">` + label + `</a>`)
	})
	raw = applyEmphasis(raw, put)
	return restoreInlineSlots(escapeExceptSlots(raw), slots)
}

func applyEmphasis(raw string, put func(string) string) string {
	raw = reBold.ReplaceAllStringFunc(raw, func(m string) string {
		match := reBold.FindStringSubmatch(m)
		if len(match) < 3 {
			return m
		}
		text := match[1]
		if text == "" {
			text = match[2]
		}
		return put("<strong>" + escapeExceptSlots(text) + "</strong>")
	})
	raw = reItalicStar.ReplaceAllStringFunc(raw, func(m string) string {
		match := reItalicStar.FindStringSubmatch(m)
		if len(match) != 4 {
			return m
		}
		return match[1] + put("<em>"+escapeExceptSlots(match[2])+"</em>") + match[3]
	})
	raw = reItalicU.ReplaceAllStringFunc(raw, func(m string) string {
		match := reItalicU.FindStringSubmatch(m)
		if len(match) != 4 {
			return m
		}
		return match[1] + put("<em>"+escapeExceptSlots(match[2])+"</em>") + match[3]
	})
	return raw
}

func escapeExceptSlots(s string) string {
	var b strings.Builder
	last := 0
	for _, loc := range reInlineSlot.FindAllStringIndex(s, -1) {
		b.WriteString(html.EscapeString(s[last:loc[0]]))
		b.WriteString(s[loc[0]:loc[1]])
		last = loc[1]
	}
	b.WriteString(html.EscapeString(s[last:]))
	return b.String()
}

func restoreInlineSlots(s string, slots []string) string {
	for i := 0; i <= len(slots) && strings.Contains(s, "\x00INL"); i++ {
		next := reInlineSlot.ReplaceAllStringFunc(s, func(m string) string {
			match := reInlineSlot.FindStringSubmatch(m)
			if len(match) != 2 {
				return ""
			}
			n, err := strconv.Atoi(match[1])
			if err != nil || n < 0 || n >= len(slots) {
				return ""
			}
			return slots[n]
		})
		if next == s {
			break
		}
		s = next
	}
	return s
}
