package analyzer

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bssm-oss/web-replica/internal/spec"
)

type HTMLAnalysis struct {
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	CanonicalURL    string             `json:"canonical_url"`
	Language        string             `json:"language"`
	ViewportMeta    string             `json:"viewport_meta"`
	ContentSummary  string             `json:"content_summary"`
	Structure       spec.PageStructure `json:"structure"`
	ButtonTexts     []string           `json:"button_texts"`
	LandmarkTexts   []string           `json:"landmark_texts"`
	CandidateAssets []spec.AssetEntry  `json:"candidate_assets"`
	TextFragments   []string           `json:"text_fragments"`
}

func AnalyzeHTML(input []byte, sourceURL string, allowOwnedAssets bool) (HTMLAnalysis, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(input))
	if err != nil {
		return HTMLAnalysis{}, fmt.Errorf("parse html: %w", err)
	}
	result := HTMLAnalysis{}
	result.Title = spec.ShortText(doc.Find("title").First().Text(), 120)
	result.Description, _ = doc.Find(`meta[name="description"]`).Attr("content")
	result.Description = spec.ShortText(result.Description, 180)
	result.CanonicalURL, _ = doc.Find(`link[rel="canonical"]`).Attr("href")
	result.Language, _ = doc.Find("html").Attr("lang")
	result.ViewportMeta, _ = doc.Find(`meta[name="viewport"]`).Attr("content")
	result.Structure = ExtractStructure(doc)
	result.CandidateAssets = CollectAssetCandidates(doc, sourceURL, allowOwnedAssets)
	result.TextFragments = collectTextFragments(doc)
	result.ContentSummary = summarizePagePurpose(result)
	result.ButtonTexts = collectButtonTexts(doc)
	result.LandmarkTexts = collectLandmarkTexts(doc)
	return result, nil
}

func collectTextFragments(doc *goquery.Document) []string {
	items := make([]string, 0, 16)
	doc.Find("main p, article p, section p").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := spec.ShortText(s.Text(), 180)
		if text != "" {
			items = append(items, text)
		}
		return len(items) < 6
	})
	if len(items) == 0 {
		items = append(items, spec.ShortText(doc.Find("body").Text(), 180))
	}
	return items
}

func collectButtonTexts(doc *goquery.Document) []string {
	items := []string{}
	doc.Find("button, input[type=submit], a[role=button]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := spec.CleanText(strings.TrimSpace(s.Text()))
		if text == "" {
			text, _ = s.Attr("value")
		}
		text = spec.ShortText(text, 50)
		if text != "" {
			items = append(items, text)
		}
		return len(items) < 10
	})
	return items
}

func collectLandmarkTexts(doc *goquery.Document) []string {
	items := []string{}
	doc.Find("header, nav, main, footer, [role=banner], [role=navigation], [role=main], [role=contentinfo]").Each(func(_ int, s *goquery.Selection) {
		text := spec.ShortText(s.Text(), 80)
		if text != "" {
			items = append(items, text)
		}
	})
	return items
}

func summarizePagePurpose(result HTMLAnalysis) string {
	headingCount := len(result.Structure.Headings)
	navCount := len(result.Structure.Navigation)
	sectionCount := len(result.Structure.Sections)
	formCount := len(result.Structure.Forms)
	imageCount := len(result.Structure.Images)
	buttonCount := len(result.ButtonTexts)
	if headingCount == 0 && navCount == 0 && sectionCount == 0 && formCount == 0 && imageCount == 0 && buttonCount == 0 {
		return "Public webpage analyzed for layout, content structure, and responsive frontend reimplementation guidance."
	}
	parts := []string{"Public webpage analyzed for safe inspired reimplementation"}
	if headingCount > 0 {
		parts = append(parts, fmt.Sprintf("%d heading group(s) detected", headingCount))
	}
	if navCount > 0 {
		parts = append(parts, fmt.Sprintf("%d navigation item(s)", navCount))
	}
	if sectionCount > 0 {
		parts = append(parts, fmt.Sprintf("%d major section candidate(s)", sectionCount))
	}
	if formCount > 0 {
		parts = append(parts, fmt.Sprintf("%d form(s)", formCount))
	}
	if buttonCount > 0 {
		parts = append(parts, fmt.Sprintf("%d call-to-action element(s)", buttonCount))
	}
	if imageCount > 0 {
		parts = append(parts, fmt.Sprintf("%d image asset(s)", imageCount))
	}
	return strings.Join(parts, "; ") + "."
}
