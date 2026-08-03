package retrieval

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type webFetchStylePattern struct {
	Property string
	Pattern  *regexp.Regexp
}

var (
	webFetchHiddenStylePatterns = []webFetchStylePattern{
		{Property: "display", Pattern: regexp.MustCompile(`^\s*none\s*$`)},
		{Property: "visibility", Pattern: regexp.MustCompile(`^\s*hidden\s*$`)},
		{Property: "opacity", Pattern: regexp.MustCompile(`^\s*0(?:\.0+)?\s*$`)},
		{Property: "font-size", Pattern: regexp.MustCompile(`^\s*0(?:px|em|rem|pt|%)?\s*$`)},
		{Property: "text-indent", Pattern: regexp.MustCompile(`^\s*-\d{4,}px\s*$`)},
		{Property: "color", Pattern: regexp.MustCompile(`^\s*transparent\s*$`)},
		{Property: "color", Pattern: regexp.MustCompile(`^\s*rgba\s*\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*,\s*0(?:\.0+)?\s*\)\s*$`)},
		{Property: "color", Pattern: regexp.MustCompile(`^\s*hsla\s*\(\s*[\d.]+\s*,\s*[\d.]+%?\s*,\s*[\d.]+%?\s*,\s*0(?:\.0+)?\s*\)\s*$`)},
	}
	webFetchHiddenClassNames = map[string]bool{
		"sr-only":            true,
		"visually-hidden":    true,
		"d-none":             true,
		"hidden":             true,
		"invisible":          true,
		"screen-reader-only": true,
		"offscreen":          true,
	}
	webFetchBoilerplateNames = map[string]bool{
		"nav":        true,
		"navbar":     true,
		"navigation": true,
		"menu":       true,
		"comments":   true,
		"comment":    true,
		"sidebar":    true,
		"breadcrumb": true,
		"pagination": true,
		"related":    true,
		"share":      true,
		"social":     true,
		"newsletter": true,
		"subscribe":  true,
		"footer":     true,
	}
	webFetchStyleNoneRe               = regexp.MustCompile(`^\s*none\s*$`)
	webFetchStyleClipPathInsetRe      = regexp.MustCompile(`inset\s*\(\s*(?:0*\.\d+|[1-9]\d*(?:\.\d+)?)%`)
	webFetchStyleTransformScaleZeroRe = regexp.MustCompile(`scale\s*\(\s*0(?:\.0+)?\s*\)`)
	webFetchStyleTranslateXRe         = regexp.MustCompile(`translateX\s*\(\s*-\d{4,}px\s*\)`)
	webFetchStyleTranslateYRe         = regexp.MustCompile(`translateY\s*\(\s*-\d{4,}px\s*\)`)
	webFetchStyleZeroSizeRe           = regexp.MustCompile(`^\s*0(?:px)?\s*$`)
	webFetchStyleOverflowHiddenRe     = regexp.MustCompile(`^\s*hidden\s*$`)
	webFetchStyleOffscreenOffsetRe    = regexp.MustCompile(`^\s*-\d{4,}px\s*$`)
	webFetchInvisibleUnicodeRe        = regexp.MustCompile(`[\x{200B}-\x{200F}\x{202A}-\x{202E}\x{2060}-\x{2064}\x{206A}-\x{206F}\x{FEFF}]`)
)

func stripInvisibleUnicode(value string) string {
	if value == "" {
		return ""
	}
	return webFetchInvisibleUnicodeRe.ReplaceAllString(value, "")
}

func sanitizeReadableSelection(sel *goquery.Selection) {
	if sel == nil || sel.Length() == 0 {
		return
	}
	dropSelectors := []string{
		"script",
		"style",
		"noscript",
		"svg",
		"canvas",
		"iframe",
		"form",
		"nav",
		"footer",
		"aside",
		"header",
		"button",
		"input",
		"textarea",
		"select",
		"template",
		`[aria-hidden="true"]`,
		`[role="navigation"]`,
		"[hidden]",
		".advertisement",
		".ads",
		".ad",
		".cookie",
		".share",
		".social",
		".related",
	}
	for _, selector := range dropSelectors {
		sel.Find(selector).Remove()
	}
	nodes := sel.Find("*")
	toRemove := make([]*goquery.Selection, 0, nodes.Length()/8+1)
	nodes.Each(func(_ int, node *goquery.Selection) {
		if shouldDropHiddenReadableNode(node) {
			toRemove = append(toRemove, node)
		}
	})
	for i := len(toRemove) - 1; i >= 0; i-- {
		toRemove[i].Remove()
	}
}

func shouldDropHiddenReadableNode(node *goquery.Selection) bool {
	if node == nil || node.Length() == 0 {
		return false
	}
	tagName := strings.ToLower(strings.TrimSpace(goquery.NodeName(node)))
	switch tagName {
	case "meta", "template", "svg", "canvas", "iframe", "object", "embed":
		return true
	}
	if tagName == "input" && strings.EqualFold(strings.TrimSpace(node.AttrOr("type", "")), "hidden") {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(node.AttrOr("aria-hidden", "")), "true") {
		return true
	}
	if _, exists := node.Attr("hidden"); exists {
		return true
	}
	if hasWebFetchHiddenClass(node.AttrOr("class", "")) {
		return true
	}
	if hasWebFetchBoilerplateName(node.AttrOr("class", "")) || hasWebFetchBoilerplateName(node.AttrOr("id", "")) {
		return true
	}
	style := node.AttrOr("style", "")
	return style != "" && isWebFetchHiddenStyle(style)
}

func hasWebFetchHiddenClass(className string) bool {
	for _, part := range strings.Fields(strings.ToLower(strings.TrimSpace(className))) {
		if webFetchHiddenClassNames[part] {
			return true
		}
	}
	return false
}

func hasWebFetchBoilerplateName(value string) bool {
	for _, part := range strings.FieldsFunc(strings.ToLower(strings.TrimSpace(value)), func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', '\r', '-', '_':
			return true
		default:
			return false
		}
	}) {
		if webFetchBoilerplateNames[part] {
			return true
		}
	}
	return false
}

func isWebFetchHiddenStyle(style string) bool {
	for _, item := range webFetchHiddenStylePatterns {
		value, ok := readCSSPropertyValue(style, item.Property)
		if ok && item.Pattern.MatchString(value) {
			return true
		}
	}
	if value, ok := readCSSPropertyValue(style, "clip-path"); ok && !webFetchStyleNoneRe.MatchString(value) {
		if webFetchStyleClipPathInsetRe.MatchString(value) {
			return true
		}
	}
	if value, ok := readCSSPropertyValue(style, "transform"); ok {
		switch {
		case webFetchStyleTransformScaleZeroRe.MatchString(value):
			return true
		case webFetchStyleTranslateXRe.MatchString(value):
			return true
		case webFetchStyleTranslateYRe.MatchString(value):
			return true
		}
	}
	width, hasWidth := readCSSPropertyValue(style, "width")
	height, hasHeight := readCSSPropertyValue(style, "height")
	overflow, hasOverflow := readCSSPropertyValue(style, "overflow")
	if hasWidth && hasHeight && hasOverflow &&
		webFetchStyleZeroSizeRe.MatchString(width) &&
		webFetchStyleZeroSizeRe.MatchString(height) &&
		webFetchStyleOverflowHiddenRe.MatchString(overflow) {
		return true
	}
	if value, ok := readCSSPropertyValue(style, "left"); ok && webFetchStyleOffscreenOffsetRe.MatchString(value) {
		return true
	}
	if value, ok := readCSSPropertyValue(style, "top"); ok && webFetchStyleOffscreenOffsetRe.MatchString(value) {
		return true
	}
	return false
}

func readCSSPropertyValue(style string, property string) (string, bool) {
	for _, chunk := range strings.Split(style, ";") {
		name, value, ok := strings.Cut(chunk, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), property) {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				return "", false
			}
			return trimmed, true
		}
	}
	return "", false
}
