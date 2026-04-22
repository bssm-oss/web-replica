package analyzer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bssm-oss/web-replica/internal/spec"
)

func ExtractStructure(doc *goquery.Document) spec.PageStructure {
	structure := spec.PageStructure{}
	doc.Find("header, nav, main, footer, [role]").Each(func(_ int, s *goquery.Selection) {
		if tag := goquery.NodeName(s); tag != "" {
			role, _ := s.Attr("role")
			structure.Landmarks = append(structure.Landmarks, spec.Landmark{Tag: tag, Role: role, Text: spec.ShortText(s.Text(), 60)})
		}
	})
	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		level := 0
		if name := goquery.NodeName(s); len(name) == 2 && name[0] == 'h' {
			level = int(name[1] - '0')
		}
		text := spec.ShortText(s.Text(), 100)
		if text != "" {
			structure.Headings = append(structure.Headings, spec.Heading{Level: level, Text: text})
		}
	})
	doc.Find("nav a").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := spec.ShortText(s.Text(), 50)
		if text != "" {
			href, _ := s.Attr("href")
			structure.Navigation = append(structure.Navigation, spec.NavItem{Text: text, Href: href})
		}
		return len(structure.Navigation) < 12
	})
	doc.Find("section, article, header, footer, main, div").Each(func(_ int, s *goquery.Selection) {
		section := classifySection(s)
		if section.Kind != "" {
			structure.Sections = append(structure.Sections, section)
		}
	})
	doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		form := spec.FormSummary{}
		form.Action, _ = s.Attr("action")
		form.Method, _ = s.Attr("method")
		s.Find("input, select, textarea, button").Each(func(_ int, field *goquery.Selection) {
			label := spec.ShortText(field.AttrOr("aria-label", ""), 50)
			if label == "" {
				label = spec.ShortText(field.AttrOr("placeholder", ""), 50)
			}
			form.Fields = append(form.Fields, spec.FormField{Type: field.AttrOr("type", goquery.NodeName(field)), Name: field.AttrOr("name", ""), Label: label})
		})
		structure.Forms = append(structure.Forms, form)
	})
	doc.Find("a[href]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := spec.ShortText(s.Text(), 70)
		href, _ := s.Attr("href")
		structure.Links = append(structure.Links, spec.LinkSummary{Text: text, Href: href})
		return len(structure.Links) < 20
	})
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		structure.Images = append(structure.Images, spec.ImageSummary{Src: s.AttrOr("src", ""), Alt: spec.ShortText(s.AttrOr("alt", ""), 80), Width: s.AttrOr("width", ""), Height: s.AttrOr("height", "")})
	})
	return structure
}

func classifySection(s *goquery.Selection) spec.Section {
	tag := goquery.NodeName(s)
	text := spec.ShortText(s.Text(), 140)
	class := strings.ToLower(s.AttrOr("class", ""))
	id := strings.ToLower(s.AttrOr("id", ""))
	kind := ""
	switch {
	case strings.Contains(class, "hero") || strings.Contains(id, "hero"):
		kind = "hero"
	case tag == "header":
		kind = "header"
	case tag == "footer":
		kind = "footer"
	case strings.Contains(class, "card") || strings.Contains(class, "feature") || strings.Contains(class, "product") || strings.Contains(class, "item"):
		kind = "card-list"
	case tag == "section" || tag == "article" || tag == "main":
		kind = "section"
	default:
		return spec.Section{}
	}
	heading := ""
	if value := s.Find("h1, h2, h3").First().Text(); value != "" {
		heading = spec.ShortText(value, 100)
	}
	labels := []string{}
	s.Find("li, .card, article").EachWithBreak(func(_ int, child *goquery.Selection) bool {
		label := spec.ShortText(child.Text(), 60)
		if label != "" {
			labels = append(labels, label)
		}
		return len(labels) < 4
	})
	return spec.Section{Kind: kind, Heading: heading, Labels: labels, Summary: text, Repeated: len(labels) > 1}
}
