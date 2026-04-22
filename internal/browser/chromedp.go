package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

type StyleSample struct {
	Selector            string `json:"selector"`
	Color               string `json:"color"`
	BackgroundColor     string `json:"backgroundColor"`
	FontFamily          string `json:"fontFamily"`
	FontSize            string `json:"fontSize"`
	FontWeight          string `json:"fontWeight"`
	LineHeight          string `json:"lineHeight"`
	LetterSpacing       string `json:"letterSpacing"`
	BorderRadius        string `json:"borderRadius"`
	BoxShadow           string `json:"boxShadow"`
	Border              string `json:"border"`
	Padding             string `json:"padding"`
	Margin              string `json:"margin"`
	Display             string `json:"display"`
	GridTemplateColumns string `json:"gridTemplateColumns"`
	FlexDirection       string `json:"flexDirection"`
	Gap                 string `json:"gap"`
	MaxWidth            string `json:"maxWidth"`
	Width               string `json:"width"`
	Height              string `json:"height"`
	Text                string `json:"text"`
}

type ViewportCapture struct {
	Viewport       Viewport
	ScreenshotPath string
	AboveFoldPath  string
	Notes          []string
	Samples        []StyleSample
}

type ValidationInfo struct {
	ScreenshotPath     string
	PageHeight         int64
	BodyTextPresent    bool
	HorizontalOverflow bool
	BlankPage          bool
}

func CapturePage(ctx context.Context, targetURL string, outputDir string, viewports []Viewport) ([]ViewportCapture, error) {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancelAlloc()
	results := make([]ViewportCapture, 0, len(viewports))
	for _, viewport := range viewports {
		browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
		capture, err := captureSingle(browserCtx, targetURL, outputDir, viewport)
		cancelBrowser()
		if err != nil {
			return nil, err
		}
		results = append(results, capture)
	}
	return results, nil
}

func captureSingle(ctx context.Context, targetURL string, outputDir string, viewport Viewport) (ViewportCapture, error) {
	var fullPNG []byte
	var foldPNG []byte
	var samplesJSON []byte
	var metrics struct {
		IsHamburger bool  `json:"isHamburger"`
		Columns     int64 `json:"columns"`
		Overflow    bool  `json:"overflow"`
	}
	taskCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	err := chromedp.Run(taskCtx,
		chromedp.EmulateViewport(viewport.Width, viewport.Height),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error { return chromedp.FullScreenshot(&fullPNG, 90).Do(ctx) }),
		chromedp.CaptureScreenshot(&foldPNG),
		chromedp.Evaluate(styleExtractionScript(), &samplesJSON),
		chromedp.Evaluate(`(() => {
			const nav = document.querySelector('nav');
			const button = nav ? nav.querySelector('button,[aria-label*="menu" i],[aria-expanded]') : document.querySelector('button,[aria-label*="menu" i]');
			const grid = Array.from(document.querySelectorAll('main, section, article, div')).find(el => getComputedStyle(el).gridTemplateColumns && getComputedStyle(el).gridTemplateColumns !== 'none');
			return {
				isHamburger: !!button,
				columns: grid ? getComputedStyle(grid).gridTemplateColumns.split(' ').length : 0,
				overflow: document.documentElement.scrollWidth > window.innerWidth,
			};
		})()`, &metrics),
	)
	if err != nil {
		return ViewportCapture{}, fmt.Errorf("capture %s viewport: %w", viewport.Name, err)
	}
	var samples []StyleSample
	if err := json.Unmarshal(samplesJSON, &samples); err != nil {
		return ViewportCapture{}, fmt.Errorf("decode style samples: %w", err)
	}
	fullPath := filepath.Join(outputDir, viewport.Name+".png")
	foldPath := filepath.Join(outputDir, viewport.Name+"-above-the-fold.png")
	if err := os.WriteFile(fullPath, fullPNG, 0o644); err != nil {
		return ViewportCapture{}, err
	}
	if err := os.WriteFile(foldPath, foldPNG, 0o644); err != nil {
		return ViewportCapture{}, err
	}
	notes := []string{}
	if metrics.IsHamburger {
		notes = append(notes, "navigation exposes a compact menu control")
	}
	if metrics.Columns > 0 {
		notes = append(notes, fmt.Sprintf("grid columns detected: %d", metrics.Columns))
	}
	if metrics.Overflow {
		notes = append(notes, "horizontal overflow detected")
	} else {
		notes = append(notes, "no horizontal overflow detected")
	}
	return ViewportCapture{Viewport: viewport, ScreenshotPath: fullPath, AboveFoldPath: foldPath, Notes: notes, Samples: samples}, nil
}

func CaptureValidation(ctx context.Context, targetURL string, screenshotPath string) (ValidationInfo, error) {
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancelAlloc()
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()
	var png []byte
	var state struct {
		PageHeight         int64 `json:"pageHeight"`
		BodyTextLength     int64 `json:"bodyTextLength"`
		HorizontalOverflow bool  `json:"horizontalOverflow"`
	}
	err := chromedp.Run(browserCtx,
		chromedp.EmulateViewport(1440, 1000),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(2*time.Second),
		chromedp.CaptureScreenshot(&png),
		chromedp.Evaluate(`(() => ({
			pageHeight: document.documentElement.scrollHeight,
			bodyTextLength: document.body ? document.body.innerText.trim().length : 0,
			horizontalOverflow: document.documentElement.scrollWidth > window.innerWidth,
		}))()`, &state),
	)
	if err != nil {
		return ValidationInfo{}, err
	}
	if err := os.WriteFile(screenshotPath, png, 0o644); err != nil {
		return ValidationInfo{}, err
	}
	return ValidationInfo{ScreenshotPath: screenshotPath, PageHeight: state.PageHeight, BodyTextPresent: state.BodyTextLength > 0, HorizontalOverflow: state.HorizontalOverflow, BlankPage: state.BodyTextLength == 0}, nil
}

func styleExtractionScript() string {
	return `(() => {
		const selectors = ['body','header','nav','main','section','article','footer','h1','h2','h3','p','a','button','input','[role="button"]','[class*="card"]','[class*="product"]','[class*="item"]','[class*="feature"]','[class*="hero"]'];
		const nodes = [];
		for (const selector of selectors) {
			const matched = Array.from(document.querySelectorAll(selector)).slice(0, 3);
			for (const node of matched) {
				const style = window.getComputedStyle(node);
				nodes.push({
					selector,
					color: style.color,
					backgroundColor: style.backgroundColor,
					fontFamily: style.fontFamily,
					fontSize: style.fontSize,
					fontWeight: style.fontWeight,
					lineHeight: style.lineHeight,
					letterSpacing: style.letterSpacing,
					borderRadius: style.borderRadius,
					boxShadow: style.boxShadow,
					border: style.border,
					padding: style.padding,
					margin: style.margin,
					display: style.display,
					gridTemplateColumns: style.gridTemplateColumns,
					flexDirection: style.flexDirection,
					gap: style.gap,
					maxWidth: style.maxWidth,
					width: style.width,
					height: style.height,
					text: (node.innerText || '').trim().slice(0, 80),
				});
			}
		}
		return nodes;
	})()`
}
