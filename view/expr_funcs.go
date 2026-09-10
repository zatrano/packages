package view

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ifFuncName(name string) bool {
	_, ok := ifCallCompilers[strings.ToLower(name)]
	return ok
}

type ifCallSpec struct {
	min, max int
	compile  func([]string) (string, error)
}

func needArgs(name string, args []string, min, max int) error {
	n := len(args)
	if n < min || (max >= 0 && n > max) {
		if max == min {
			return fmt.Errorf("%s() expects %d argument(s)", name, min)
		}
		return fmt.Errorf("%s() expects %d to %d arguments", name, min, max)
	}
	return nil
}

func call1(fn string) func([]string) (string, error) {
	return func(args []string) (string, error) {
		if err := needArgs(fn, args, 1, 1); err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s)", fn, args[0]), nil
	}
}

func call2(fn string) func([]string) (string, error) {
	return func(args []string) (string, error) {
		if err := needArgs(fn, args, 2, 2); err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", fn, args[0], args[1]), nil
	}
}

var ifCallCompilers = map[string]ifCallSpec{
	"empty":            {1, 1, call1("empty")},
	"count":            {1, 1, call1("ifCount")},
	"sizeof":           {1, 1, call1("ifCount")},
	"is_null":          {1, 1, call1("isNull")},
	"isnull":           {1, 1, call1("isNull")},
	"is_numeric":       {1, 1, call1("isNumeric")},
	"is_string":        {1, 1, call1("isString")},
	"is_array":         {1, 1, call1("isArray")},
	"is_bool":          {1, 1, call1("isBool")},
	"is_int":           {1, 1, call1("isInt")},
	"is_integer":       {1, 1, call1("isInt")},
	"is_float":         {1, 1, call1("isFloat")},
	"is_object":        {1, 1, call1("isObject")},
	"is_countable":     {1, 1, call1("isCountable")},
	"is_scalar":        {1, 1, call1("isScalar")},
	"filled":           {1, 1, call1("ifFilled")},
	"blank":            {1, 1, call1("ifBlank")},
	"in_array":         {2, 3, call2("inArray")},
	"array_key_exists": {2, 2, call2("arrayKeyExists")},
	"key_exists":       {2, 2, call2("arrayKeyExists")},
	"str_contains":     {2, 2, call2("strContains")},
	"str_starts_with":  {2, 2, call2("strStartsWith")},
	"str_ends_with":    {2, 2, call2("strEndsWith")},
	"strlen":           {1, 1, call1("ifStrlen")},
	"mb_strlen":        {1, 1, call1("ifMbStrlen")},
	"strtolower":       {1, 1, call1("ifLower")},
	"mb_strtolower":    {1, 1, call1("ifLower")},
	"strtoupper":       {1, 1, call1("ifUpper")},
	"mb_strtoupper":    {1, 1, call1("ifUpper")},
	"trim":             {1, 2, call1("ifTrim")},
	"ltrim":            {1, 1, call1("ifLtrim")},
	"rtrim":            {1, 1, call1("ifRtrim")},
	"ucfirst":          {1, 1, call1("ifUcfirst")},
	"lcfirst":          {1, 1, call1("ifLcfirst")},
	"ucwords":          {1, 1, call1("ifUcwords")},
	"abs":              {1, 1, call1("ifAbs")},
	"round":            {1, 2, call1("ifRound")},
	"floor":            {1, 1, call1("ifFloor")},
	"ceil":             {1, 1, call1("ifCeil")},
	"intval":           {1, 1, call1("ifIntval")},
	"floatval":         {1, 1, call1("ifFloatval")},
	"strval":           {1, 1, call1("toString")},
	"boolval":          {1, 1, call1("ifBoolval")},
	"isset": {1, -1, func(args []string) (string, error) {
		if err := needArgs("isset", args, 1, -1); err != nil {
			return "", err
		}
		out := fmt.Sprintf("(ifIsset %s)", args[0])
		for _, a := range args[1:] {
			out = fmt.Sprintf("(and %s (ifIsset %s))", out, a)
		}
		return out, nil
	}},
	"data_get": {2, 3, func(args []string) (string, error) {
		if err := needArgs("data_get", args, 2, 3); err != nil {
			return "", err
		}
		got := fmt.Sprintf("(dataGet %s %s)", args[0], args[1])
		if len(args) == 3 {
			return fmt.Sprintf("(ifCoalesce %s %s)", got, args[2]), nil
		}
		return got, nil
	}},
	"min": {2, -1, func(args []string) (string, error) {
		if err := needArgs("min", args, 2, -1); err != nil {
			return "", err
		}
		return "(ifMin " + strings.Join(args, " ") + ")", nil
	}},
	"max": {2, -1, func(args []string) (string, error) {
		if err := needArgs("max", args, 2, -1); err != nil {
			return "", err
		}
		return "(ifMax " + strings.Join(args, " ") + ")", nil
	}},
	"implode": {1, 2, compileImplode},
	"join":    {1, 2, compileImplode},
	"explode": {2, 3, func(args []string) (string, error) {
		if err := needArgs("explode", args, 2, 2); err != nil {
			return "", err
		}
		return fmt.Sprintf("(ifExplode %s %s)", args[0], args[1]), nil
	}},
	"str_replace": {3, 3, func(args []string) (string, error) {
		if err := needArgs("str_replace", args, 3, 3); err != nil {
			return "", err
		}
		return fmt.Sprintf("(ifStrReplace %s %s %s)", args[0], args[1], args[2]), nil
	}},
	"substr": {2, 3, func(args []string) (string, error) {
		if err := needArgs("substr", args, 2, 3); err != nil {
			return "", err
		}
		if len(args) == 2 {
			return fmt.Sprintf("(ifSubstr %s %s -1)", args[0], args[1]), nil
		}
		return fmt.Sprintf("(ifSubstr %s %s %s)", args[0], args[1], args[2]), nil
	}},
	"strpos": {2, 3, func(args []string) (string, error) {
		if err := needArgs("strpos", args, 2, 2); err != nil {
			return "", err
		}
		return fmt.Sprintf("(ifStrpos %s %s)", args[0], args[1]), nil
	}},
}

func compileImplode(args []string) (string, error) {
	if err := needArgs("implode", args, 1, 2); err != nil {
		return "", err
	}
	if len(args) == 1 {
		return fmt.Sprintf("(ifImplode `` %s)", args[0]), nil
	}
	return fmt.Sprintf("(ifImplode %s %s)", args[0], args[1]), nil
}

func compileIfCall(name string, args []string) (string, error) {
	spec, ok := ifCallCompilers[name]
	if !ok {
		return "", fmt.Errorf("unsupported function %s in expression", name)
	}
	if name == "in_array" && len(args) > 2 {
		args = args[:2]
	}
	if spec.compile != nil {
		return spec.compile(args)
	}
	return "", fmt.Errorf("unsupported function %s in expression", name)
}

func pathToString(path any) string {
	switch x := path.(type) {
	case nil:
		return ""
	case string:
		return x
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	default:
		return fmt.Sprint(x)
	}
}

func ifTernary(cond, yes, no any) any {
	if isEmptyValue(cond) {
		return no
	}
	return yes
}

func ifElvis(a, b any) any {
	if isEmptyValue(a) {
		return b
	}
	return a
}

func ifCoalesce(a, b any) any {
	if ifIsset(a) {
		return a
	}
	return b
}

func ifXor(a, b any) bool {
	return isEmptyValue(a) != isEmptyValue(b)
}

func strConcat(a, b any) string {
	return toString(a) + toString(b)
}

func numPow(a, b any) any {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if !aok || !bok {
		return 0
	}
	out := math.Pow(af, bf)
	if i := int64(out); out == float64(i) {
		return i
	}
	return out
}

func isNull(v any) bool { return !ifIsset(v) }

func isNumeric(v any) bool {
	_, ok := toFloat64(v)
	return ok
}

func isString(v any) bool {
	_, ok := v.(string)
	return ok
}

func isArray(v any) bool {
	if v == nil {
		return false
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return true
	}
	return false
}

func isBool(v any) bool {
	_, ok := v.(bool)
	return ok
}

func isInt(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func isFloat(v any) bool {
	switch v.(type) {
	case float32, float64:
		return true
	}
	return false
}

func isObject(v any) bool {
	if v == nil {
		return false
	}
	k := reflect.ValueOf(v).Kind()
	return k == reflect.Struct || k == reflect.Map
}

func isCountable(v any) bool {
	if v == nil {
		return false
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return true
	}
	return false
}

func isScalar(v any) bool {
	return isBool(v) || isInt(v) || isFloat(v) || isString(v)
}

func ifBlank(v any) bool {
	if isEmptyValue(v) {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}

func ifFilled(v any) bool { return !ifBlank(v) }

func inArray(needle, haystack any) bool {
	if haystack == nil {
		return false
	}
	rv := reflect.ValueOf(haystack)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}
	want := fmt.Sprint(needle)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if fmt.Sprint(rv.Index(i).Interface()) == want {
				return true
			}
		}
	case reflect.Map:
		for _, k := range rv.MapKeys() {
			if fmt.Sprint(rv.MapIndex(k).Interface()) == want {
				return true
			}
		}
	}
	return false
}

func arrayKeyExists(key, arr any) bool {
	if arr == nil {
		return false
	}
	rv := reflect.ValueOf(arr)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Map:
		return mapIndexFlexible(rv, key).IsValid()
	case reflect.Slice, reflect.Array:
		f, ok := toFloat64(key)
		if !ok {
			return false
		}
		i := int(f)
		return i >= 0 && i < rv.Len()
	}
	return false
}

func strContains(haystack, needle any) bool {
	return strings.Contains(toString(haystack), toString(needle))
}

func strStartsWith(haystack, needle any) bool {
	return strings.HasPrefix(toString(haystack), toString(needle))
}

func strEndsWith(haystack, needle any) bool {
	return strings.HasSuffix(toString(haystack), toString(needle))
}

func ifStrlen(v any) int   { return len(toString(v)) }
func ifMbStrlen(v any) int { return utf8.RuneCountInString(toString(v)) }
func ifLower(v any) string { return strings.ToLower(toString(v)) }
func ifUpper(v any) string { return strings.ToUpper(toString(v)) }
func ifTrim(v any) string  { return strings.TrimSpace(toString(v)) }
func ifLtrim(v any) string { return strings.TrimLeftFunc(toString(v), unicode.IsSpace) }
func ifRtrim(v any) string { return strings.TrimRightFunc(toString(v), unicode.IsSpace) }

func ifUcfirst(v any) string {
	s := toString(v)
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return strings.ToUpper(string(r)) + s[size:]
}

func ifLcfirst(v any) string {
	s := toString(v)
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return strings.ToLower(string(r)) + s[size:]
}

func ifUcwords(v any) string {
	s := strings.ToLower(toString(v))
	prev := ' '
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(prev) || prev == '-' {
			b.WriteRune(unicode.ToTitle(r))
		} else {
			b.WriteRune(r)
		}
		prev = r
	}
	return b.String()
}

func ifAbs(v any) any {
	f, ok := toFloat64(v)
	if !ok {
		return 0
	}
	if f < 0 {
		f = -f
	}
	if i := int64(f); f == float64(i) {
		return i
	}
	return f
}

func ifRound(v any) any {
	f, ok := toFloat64(v)
	if !ok {
		return 0
	}
	out := math.Round(f)
	if i := int64(out); out == float64(i) {
		return i
	}
	return out
}

func ifFloor(v any) any {
	f, ok := toFloat64(v)
	if !ok {
		return 0
	}
	return int64(math.Floor(f))
}

func ifCeil(v any) any {
	f, ok := toFloat64(v)
	if !ok {
		return 0
	}
	return int64(math.Ceil(f))
}

func ifIntval(v any) int64 {
	f, ok := toFloat64(v)
	if !ok {
		return 0
	}
	return int64(f)
}

func ifFloatval(v any) float64 {
	f, ok := toFloat64(v)
	if !ok {
		return 0
	}
	return f
}

func ifBoolval(v any) bool { return !isEmptyValue(v) }

func ifMin(vals ...any) any {
	if len(vals) == 0 {
		return 0
	}
	best, ok := toFloat64(vals[0])
	if !ok {
		return 0
	}
	for _, v := range vals[1:] {
		f, fok := toFloat64(v)
		if fok && f < best {
			best = f
		}
	}
	if i := int64(best); best == float64(i) {
		return i
	}
	return best
}

func ifMax(vals ...any) any {
	if len(vals) == 0 {
		return 0
	}
	best, ok := toFloat64(vals[0])
	if !ok {
		return 0
	}
	for _, v := range vals[1:] {
		f, fok := toFloat64(v)
		if fok && f > best {
			best = f
		}
	}
	if i := int64(best); best == float64(i) {
		return i
	}
	return best
}

func ifImplode(glue, arr any) string {
	g := toString(glue)
	if arr == nil {
		return ""
	}
	rv := reflect.ValueOf(arr)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return toString(arr)
	}
	parts := make([]string, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		parts = append(parts, toString(rv.Index(i).Interface()))
	}
	return strings.Join(parts, g)
}

func ifExplode(sep, s any) []string {
	if toString(sep) == "" {
		return []string{toString(s)}
	}
	return strings.Split(toString(s), toString(sep))
}

func ifStrReplace(search, repl, subject any) string {
	return strings.ReplaceAll(toString(subject), toString(search), toString(repl))
}

func ifSubstr(s any, start any, length any) string {
	str := toString(s)
	runes := []rune(str)
	st, ok := toFloat64(start)
	if !ok {
		return ""
	}
	i := int(st)
	if i < 0 {
		i = len(runes) + i
	}
	if i < 0 {
		i = 0
	}
	if i > len(runes) {
		return ""
	}
	lf, lok := toFloat64(length)
	if !lok || int(lf) < 0 {
		return string(runes[i:])
	}
	end := i + int(lf)
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[i:end])
}

func ifStrpos(haystack, needle any) any {
	h, n := toString(haystack), toString(needle)
	i := strings.Index(h, n)
	if i < 0 {
		return false
	}
	return i
}
