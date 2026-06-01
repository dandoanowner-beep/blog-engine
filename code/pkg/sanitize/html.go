package sanitize

import "github.com/microcosm-cc/bluemonday"

type HTMLSanitizer struct {
	policy *bluemonday.Policy
}

func NewHTMLSanitizer() *HTMLSanitizer {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "strong", "em", "u", "s", "h1", "h2", "h3", "h4",
		"ul", "ol", "li", "code", "pre", "blockquote", "hr")
	p.AllowAttrs("class").OnElements("code", "pre")
	p.AllowStandardURLs()
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "alt").OnElements("img")
	p.RequireNoFollowOnLinks(false)
	return &HTMLSanitizer{policy: p}
}

func (s *HTMLSanitizer) Sanitize(html string) string {
	return s.policy.Sanitize(html)
}
