package view

import (
	"fmt"
	"strings"
	"unicode"
)

type ifTokKind int

const (
	ifTokEOF ifTokKind = iota
	ifTokAnd
	ifTokOr
	ifTokEq
	ifTokNe
	ifTokGe
	ifTokLe
	ifTokGt
	ifTokLt
	ifTokNot
	ifTokPlus
	ifTokMinus
	ifTokStar
	ifTokSlash
	ifTokPercent
	ifTokLParen
	ifTokRParen
	ifTokLBracket
	ifTokRBracket
	ifTokComma
	ifTokQuestion
	ifTokColon
	ifTokNullCoal
	ifTokPower
	ifTokConcat
	ifTokAndWord
	ifTokOrWord
	ifTokXor
	ifTokVar
	ifTokParent
	ifTokRange
	ifTokNumber
	ifTokString
	ifTokIdent
	ifTokTrue
	ifTokFalse
	ifTokNull
)

type ifTok struct {
	kind  ifTokKind
	raw   string
	path  string
	alias string
	field string
}

type ifLexer struct {
	s   string
	i   int
	err error
}

func (l *ifLexer) skipSpace() {
	for l.i < len(l.s) && unicode.IsSpace(rune(l.s[l.i])) {
		l.i++
	}
}

func (l *ifLexer) ident() (string, bool) {
	if l.i >= len(l.s) {
		return "", false
	}
	c := l.s[l.i]
	if !isIdentStart(c) {
		return "", false
	}
	start := l.i
	l.i++
	for l.i < len(l.s) && isIdentPart(l.s[l.i]) {
		l.i++
	}
	return l.s[start:l.i], true
}

func isIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || (c >= '0' && c <= '9')
}

func isIdentPart(c byte) bool {
	return isIdentStart(c)
}

func isNameStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func (l *ifLexer) dotted(first string) string {
	var b strings.Builder
	b.WriteString(first)
	for l.i < len(l.s) && l.s[l.i] == '.' {
		l.i++
		part, ok := l.ident()
		if !ok {
			l.err = fmt.Errorf("expected identifier after '.'")
			return b.String()
		}
		b.WriteByte('.')
		b.WriteString(part)
	}
	return b.String()
}

func (l *ifLexer) next() ifTok {
	if l.err != nil {
		return ifTok{kind: ifTokEOF}
	}
	l.skipSpace()
	if l.i >= len(l.s) {
		return ifTok{kind: ifTokEOF}
	}
	s := l.s[l.i:]
	switch {
	case strings.HasPrefix(s, "&&"):
		l.i += 2
		return ifTok{kind: ifTokAnd}
	case strings.HasPrefix(s, "||"):
		l.i += 2
		return ifTok{kind: ifTokOr}
	case strings.HasPrefix(s, "??"):
		l.i += 2
		return ifTok{kind: ifTokNullCoal}
	case strings.HasPrefix(s, "**"):
		l.i += 2
		return ifTok{kind: ifTokPower}
	case strings.HasPrefix(s, "==="):
		l.i += 3
		return ifTok{kind: ifTokEq}
	case strings.HasPrefix(s, "!=="):
		l.i += 3
		return ifTok{kind: ifTokNe}
	case strings.HasPrefix(s, "=="):
		l.i += 2
		return ifTok{kind: ifTokEq}
	case strings.HasPrefix(s, "!="):
		l.i += 2
		return ifTok{kind: ifTokNe}
	case strings.HasPrefix(s, ">="):
		l.i += 2
		return ifTok{kind: ifTokGe}
	case strings.HasPrefix(s, "<="):
		l.i += 2
		return ifTok{kind: ifTokLe}
	case strings.HasPrefix(s, "__ZPARENT__."):
		l.i += len("__ZPARENT__.")
		part, ok := l.ident()
		if !ok {
			l.err = fmt.Errorf("expected path after __ZPARENT__.")
			return ifTok{kind: ifTokEOF}
		}
		return ifTok{kind: ifTokParent, path: l.dotted(part)}
	case strings.HasPrefix(s, "__ZRV_"):
		l.i += len("__ZRV_")
		start := l.i
		for l.i < len(l.s) {
			if l.i+1 < len(l.s) && l.s[l.i] == '_' && l.s[l.i+1] == '_' {
				break
			}
			if !isIdentPart(l.s[l.i]) {
				l.err = fmt.Errorf("malformed range variable")
				return ifTok{kind: ifTokEOF}
			}
			l.i++
		}
		if start == l.i || l.i+1 >= len(l.s) || l.s[l.i] != '_' || l.s[l.i+1] != '_' {
			l.err = fmt.Errorf("malformed range variable")
			return ifTok{kind: ifTokEOF}
		}
		alias := l.s[start:l.i]
		l.i += 2
		tok := ifTok{kind: ifTokRange, alias: alias}
		if l.i < len(l.s) && l.s[l.i] == '.' {
			l.i++
			field, ok := l.ident()
			if !ok {
				l.err = fmt.Errorf("expected field after range variable")
				return ifTok{kind: ifTokEOF}
			}
			tok.field = l.dotted(field)
		}
		return tok
	}
	c := l.s[l.i]
	switch c {
	case '!':
		l.i++
		return ifTok{kind: ifTokNot}
	case '?':
		l.i++
		return ifTok{kind: ifTokQuestion}
	case ':':
		l.i++
		return ifTok{kind: ifTokColon}
	case ',':
		l.i++
		return ifTok{kind: ifTokComma}
	case '>':
		l.i++
		return ifTok{kind: ifTokGt}
	case '<':
		l.i++
		return ifTok{kind: ifTokLt}
	case '+':
		l.i++
		return ifTok{kind: ifTokPlus}
	case '-':
		l.i++
		return ifTok{kind: ifTokMinus}
	case '*':
		l.i++
		return ifTok{kind: ifTokStar}
	case '/':
		l.i++
		return ifTok{kind: ifTokSlash}
	case '%':
		l.i++
		return ifTok{kind: ifTokPercent}
	case '.':
		if l.i+1 < len(l.s) && l.s[l.i+1] >= '0' && l.s[l.i+1] <= '9' {
			start := l.i
			l.i++
			for l.i < len(l.s) && l.s[l.i] >= '0' && l.s[l.i] <= '9' {
				l.i++
			}
			return ifTok{kind: ifTokNumber, raw: l.s[start:l.i]}
		}
		l.i++
		return ifTok{kind: ifTokConcat}
	case '(':
		l.i++
		return ifTok{kind: ifTokLParen}
	case ')':
		l.i++
		return ifTok{kind: ifTokRParen}
	case '[':
		l.i++
		return ifTok{kind: ifTokLBracket}
	case ']':
		l.i++
		return ifTok{kind: ifTokRBracket}
	case '$':
		l.i++
		part, ok := l.ident()
		if !ok {
			l.err = fmt.Errorf("expected identifier after $")
			return ifTok{kind: ifTokEOF}
		}
		return ifTok{kind: ifTokVar, path: l.dotted(part)}
	case '\'', '"':
		quote := c
		l.i++
		start := l.i
		for l.i < len(l.s) && l.s[l.i] != quote {
			l.i++
		}
		if l.i >= len(l.s) {
			l.err = fmt.Errorf("unterminated string")
			return ifTok{kind: ifTokEOF}
		}
		lit := l.s[start:l.i]
		l.i++
		return ifTok{kind: ifTokString, raw: lit}
	}
	if c >= '0' && c <= '9' {
		start := l.i
		for l.i < len(l.s) && l.s[l.i] >= '0' && l.s[l.i] <= '9' {
			l.i++
		}
		if l.i < len(l.s) && l.s[l.i] == '.' {
			l.i++
			if l.i >= len(l.s) || l.s[l.i] < '0' || l.s[l.i] > '9' {
				l.err = fmt.Errorf("expected digits after decimal")
				return ifTok{kind: ifTokEOF}
			}
			for l.i < len(l.s) && l.s[l.i] >= '0' && l.s[l.i] <= '9' {
				l.i++
			}
		}
		return ifTok{kind: ifTokNumber, raw: l.s[start:l.i]}
	}
	if isNameStart(c) {
		name, _ := l.ident()
		switch strings.ToLower(name) {
		case "and":
			return ifTok{kind: ifTokAndWord}
		case "or":
			return ifTok{kind: ifTokOrWord}
		case "xor":
			return ifTok{kind: ifTokXor}
		case "not":
			return ifTok{kind: ifTokNot}
		case "true":
			return ifTok{kind: ifTokTrue}
		case "false":
			return ifTok{kind: ifTokFalse}
		case "null", "nil":
			return ifTok{kind: ifTokNull}
		default:
			if ifFuncName(name) {
				return ifTok{kind: ifTokIdent, raw: strings.ToLower(name)}
			}
			l.err = fmt.Errorf("unexpected identifier %q in expression", name)
			return ifTok{kind: ifTokEOF}
		}
	}
	l.err = fmt.Errorf("unexpected %q in @if expression", s[0:1])
	return ifTok{kind: ifTokEOF}
}

type ifParser struct {
	lx  ifLexer
	cur ifTok
}

func (p *ifParser) eat() {
	p.cur = p.lx.next()
}

func compileIfInner(expr string) (string, error) {
	p := &ifParser{lx: ifLexer{s: strings.TrimSpace(expr)}}
	p.eat()
	if p.lx.err != nil {
		return "", p.lx.err
	}
	if p.cur.kind == ifTokEOF {
		return "", fmt.Errorf("empty expression")
	}
	out, err := p.parseExpr()
	if err != nil {
		return "", err
	}
	if p.lx.err != nil {
		return "", p.lx.err
	}
	if p.cur.kind != ifTokEOF {
		return "", fmt.Errorf("unexpected token after expression")
	}
	return out, nil
}

func (p *ifParser) parseExpr() (string, error) {
	return p.parseOrWord()
}

func (p *ifParser) parseOrWord() (string, error) {
	left, err := p.parseXor()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokOrWord {
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseXor()
		if err != nil {
			return "", err
		}
		left = fmt.Sprintf("(or %s %s)", left, right)
	}
	return left, nil
}

func (p *ifParser) parseXor() (string, error) {
	left, err := p.parseAndWord()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokXor {
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseAndWord()
		if err != nil {
			return "", err
		}
		left = fmt.Sprintf("(ifXor %s %s)", left, right)
	}
	return left, nil
}

func (p *ifParser) parseAndWord() (string, error) {
	left, err := p.parseTernary()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokAndWord {
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseTernary()
		if err != nil {
			return "", err
		}
		left = fmt.Sprintf("(and %s %s)", left, right)
	}
	return left, nil
}

func (p *ifParser) parseTernary() (string, error) {
	cond, err := p.parseNullCoal()
	if err != nil {
		return "", err
	}
	if p.cur.kind != ifTokQuestion {
		return cond, nil
	}
	p.eat()
	if p.lx.err != nil {
		return "", p.lx.err
	}
	if p.cur.kind == ifTokColon {
		p.eat()
		alt, err := p.parseTernary()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(ifElvis %s %s)", cond, alt), nil
	}
	yes, err := p.parseTernary()
	if err != nil {
		return "", err
	}
	if p.cur.kind != ifTokColon {
		return "", fmt.Errorf("expected ':' in ternary expression")
	}
	p.eat()
	no, err := p.parseTernary()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(ifTernary %s %s %s)", cond, yes, no), nil
}

func (p *ifParser) parseNullCoal() (string, error) {
	left, err := p.parseOr()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokNullCoal {
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseOr()
		if err != nil {
			return "", err
		}
		left = fmt.Sprintf("(ifCoalesce %s %s)", left, right)
	}
	return left, nil
}

func (p *ifParser) parseOr() (string, error) {
	left, err := p.parseAnd()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokOr {
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseAnd()
		if err != nil {
			return "", err
		}
		left = fmt.Sprintf("(or %s %s)", left, right)
	}
	return left, nil
}

func (p *ifParser) parseAnd() (string, error) {
	left, err := p.parseCmp()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokAnd {
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseCmp()
		if err != nil {
			return "", err
		}
		left = fmt.Sprintf("(and %s %s)", left, right)
	}
	return left, nil
}

func (p *ifParser) parseCmp() (string, error) {
	left, err := p.parseConcat()
	if err != nil {
		return "", err
	}
	op := p.cur.kind
	switch op {
	case ifTokEq, ifTokNe, ifTokGe, ifTokLe, ifTokGt, ifTokLt:
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseConcat()
		if err != nil {
			return "", err
		}
		return compileIfCompare(op, left, right), nil
	default:
		return left, nil
	}
}

func (p *ifParser) parseConcat() (string, error) {
	left, err := p.parseAdd()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokConcat {
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseAdd()
		if err != nil {
			return "", err
		}
		left = fmt.Sprintf("(strConcat %s %s)", left, right)
	}
	return left, nil
}

func (p *ifParser) parseAdd() (string, error) {
	left, err := p.parseMul()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokPlus || p.cur.kind == ifTokMinus {
		op := p.cur.kind
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parseMul()
		if err != nil {
			return "", err
		}
		if op == ifTokPlus {
			left = fmt.Sprintf("(numAdd %s %s)", left, right)
		} else {
			left = fmt.Sprintf("(numSub %s %s)", left, right)
		}
	}
	return left, nil
}

func (p *ifParser) parseMul() (string, error) {
	left, err := p.parsePow()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokStar || p.cur.kind == ifTokSlash || p.cur.kind == ifTokPercent {
		op := p.cur.kind
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		right, err := p.parsePow()
		if err != nil {
			return "", err
		}
		switch op {
		case ifTokStar:
			left = fmt.Sprintf("(numMul %s %s)", left, right)
		case ifTokSlash:
			left = fmt.Sprintf("(numDiv %s %s)", left, right)
		default:
			left = fmt.Sprintf("(numMod %s %s)", left, right)
		}
	}
	return left, nil
}

func (p *ifParser) parsePow() (string, error) {
	left, err := p.parseUnary()
	if err != nil {
		return "", err
	}
	if p.cur.kind != ifTokPower {
		return left, nil
	}
	p.eat()
	if p.lx.err != nil {
		return "", p.lx.err
	}
	right, err := p.parsePow()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(numPow %s %s)", left, right), nil
}

func (p *ifParser) parseUnary() (string, error) {
	if p.lx.err != nil {
		return "", p.lx.err
	}
	switch p.cur.kind {
	case ifTokNot:
		p.eat()
		inner, err := p.parseUnary()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(not %s)", inner), nil
	case ifTokMinus:
		p.eat()
		inner, err := p.parseUnary()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(numNeg %s)", inner), nil
	case ifTokPlus:
		p.eat()
		return p.parseUnary()
	default:
		return p.parsePostfix()
	}
}

func (p *ifParser) parsePostfix() (string, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return "", err
	}
	for p.cur.kind == ifTokLBracket {
		p.eat()
		if p.lx.err != nil {
			return "", p.lx.err
		}
		idx, err := p.parseExpr()
		if err != nil {
			return "", err
		}
		if p.cur.kind != ifTokRBracket {
			return "", fmt.Errorf("expected ']'")
		}
		p.eat()
		left = fmt.Sprintf("(dataIndex %s %s)", left, idx)
	}
	return left, nil
}

func (p *ifParser) parsePrimary() (string, error) {
	if p.lx.err != nil {
		return "", p.lx.err
	}
	switch p.cur.kind {
	case ifTokLParen:
		p.eat()
		inner, err := p.parseExpr()
		if err != nil {
			return "", err
		}
		if p.cur.kind != ifTokRParen {
			return "", fmt.Errorf("expected ')'")
		}
		p.eat()
		return inner, nil
	case ifTokIdent:
		return p.parseCall()
	case ifTokTrue:
		p.eat()
		return "true", nil
	case ifTokFalse:
		p.eat()
		return "false", nil
	case ifTokNull:
		p.eat()
		return "(ifNull)", nil
	case ifTokVar, ifTokParent, ifTokRange, ifTokNumber, ifTokString:
		out, err := compileIfValue(p.cur)
		if err != nil {
			return "", err
		}
		p.eat()
		return out, nil
	case ifTokEOF:
		return "", fmt.Errorf("unexpected end of @if expression")
	default:
		return "", fmt.Errorf("unexpected token in @if expression")
	}
}

func (p *ifParser) parseCall() (string, error) {
	name := p.cur.raw
	p.eat()
	if p.cur.kind != ifTokLParen {
		return "", fmt.Errorf("expected '(' after %s", name)
	}
	p.eat()
	if p.lx.err != nil {
		return "", p.lx.err
	}
	var args []string
	if p.cur.kind != ifTokRParen {
		for {
			arg, err := p.parseExpr()
			if err != nil {
				return "", err
			}
			args = append(args, arg)
			if p.cur.kind != ifTokComma {
				break
			}
			p.eat()
			if p.lx.err != nil {
				return "", p.lx.err
			}
		}
	}
	if p.cur.kind != ifTokRParen {
		return "", fmt.Errorf("expected ')' after %s(...)", name)
	}
	p.eat()
	return compileIfCall(name, args)
}

func compileIfValue(tok ifTok) (string, error) {
	switch tok.kind {
	case ifTokVar:
		return fmt.Sprintf("(dataGet . %s)", tplLit(tok.path)), nil
	case ifTokParent:
		return fmt.Sprintf("(dataGet $ %s)", tplLit(tok.path)), nil
	case ifTokRange:
		if tok.field == "" {
			return "$" + tok.alias, nil
		}
		return fmt.Sprintf("(dataGet $%s %s)", tok.alias, tplLit(tok.field)), nil
	case ifTokNumber:
		return tok.raw, nil
	case ifTokString:
		return tplLit(tok.raw), nil
	default:
		return "", fmt.Errorf("invalid value in @if expression")
	}
}

func compileIfCompare(op ifTokKind, left, right string) string {
	switch op {
	case ifTokEq:
		return fmt.Sprintf("(eq (printf `%%v` %s) (printf `%%v` %s))", left, right)
	case ifTokNe:
		return fmt.Sprintf("(ne (printf `%%v` %s) (printf `%%v` %s))", left, right)
	case ifTokGe:
		return fmt.Sprintf("(cmpGe %s %s)", left, right)
	case ifTokLe:
		return fmt.Sprintf("(cmpLe %s %s)", left, right)
	case ifTokGt:
		return fmt.Sprintf("(cmpGt %s %s)", left, right)
	case ifTokLt:
		return fmt.Sprintf("(cmpLt %s %s)", left, right)
	default:
		return left
	}
}

func matchingParen(s string, open int) (int, bool) {
	if open < 0 || open >= len(s) || s[open] != '(' {
		return 0, false
	}
	depth := 0
	var quote byte
	esc := false
	for i := open; i < len(s); i++ {
		c := s[i]
		if quote != 0 {
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func skipDirectiveSpace(s string) string {
	return strings.TrimLeft(s, " \t\n\r")
}

func readIfCall(s string, at int) (kind, expr string, end int, ok bool) {
	if at < 0 || at >= len(s) || s[at] != '@' {
		return "", "", 0, false
	}
	rest := s[at+1:]
	lower := strings.ToLower(rest)
	switch {
	case strings.HasPrefix(lower, "elseif"):
		kind = "elseif"
		rest = rest[len("elseif"):]
	case strings.HasPrefix(lower, "if"):
		kind = "if"
		rest = rest[len("if"):]
	default:
		return "", "", 0, false
	}
	rest = skipDirectiveSpace(rest)
	if !strings.HasPrefix(rest, "(") {
		return "", "", 0, false
	}
	open := len(s) - len(rest)
	close, found := matchingParen(s, open)
	if !found {
		return "", "", 0, false
	}
	return kind, s[open+1 : close], close + 1, true
}

func findNextIfCall(s string, from int) int {
	i := from
	for i < len(s) {
		at := strings.IndexByte(s[i:], '@')
		if at < 0 {
			return -1
		}
		at += i
		_, _, _, ok := readIfCall(s, at)
		if ok {
			return at
		}
		i = at + 1
	}
	return -1
}

func compileEchoExpressions(out string) string {
	var b strings.Builder
	i := 0
	for i < len(out) {
		start, raw, ok := findEchoOpen(out, i)
		if !ok {
			b.WriteString(out[i:])
			break
		}
		b.WriteString(out[i:start])
		inner, end, found := readEcho(out, start, raw)
		if !found {
			b.WriteByte(out[start])
			i = start + 1
			continue
		}
		compiled, err := compileIfInner(strings.TrimSpace(inner))
		if err != nil {
			b.WriteString(out[start:end])
			i = end
			continue
		}
		if raw {
			b.WriteString("{{ safeStr ")
			b.WriteString(compiled)
			b.WriteString(" }}")
		} else {
			b.WriteString("{{ ")
			b.WriteString(compiled)
			b.WriteString(" }}")
		}
		i = end
	}
	return b.String()
}

func findEchoOpen(s string, from int) (start int, raw bool, ok bool) {
	i := from
	for i < len(s) {
		j := strings.IndexByte(s[i:], '{')
		if j < 0 {
			return 0, false, false
		}
		j += i
		if strings.HasPrefix(s[j:], "{!!") {
			return j, true, true
		}
		if strings.HasPrefix(s[j:], "{{") {
			return j, false, true
		}
		i = j + 1
	}
	return 0, false, false
}

func readEcho(s string, start int, raw bool) (inner string, end int, ok bool) {
	open, close := "{{", "}}"
	if raw {
		open, close = "{!!", "!!}"
	}
	if !strings.HasPrefix(s[start:], open) {
		return "", 0, false
	}
	bodyStart := start + len(open)
	closeAt := indexEchoClose(s, bodyStart, close)
	if closeAt < 0 {
		return "", 0, false
	}
	return s[bodyStart:closeAt], closeAt + len(close), true
}

func indexEchoClose(s string, from int, close string) int {
	var quote byte
	esc := false
	for i := from; i < len(s); i++ {
		c := s[i]
		if quote != 0 {
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if strings.HasPrefix(s[i:], close) {
			return i
		}
	}
	return -1
}
