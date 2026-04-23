package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
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

var blockedBrowserCIDRs = mustPrefixes([]string{
	"127.0.0.0/8",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"169.254.0.0/16",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
})

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
	if err := enableRequestGuard(taskCtx, false); err != nil {
		return ViewportCapture{}, fmt.Errorf("enable request guard: %w", err)
	}
	err := chromedp.Run(taskCtx,
		chromedp.EmulateViewport(viewport.Width, viewport.Height),
		chromedp.Navigate(targetURL),
		waitForPageStable(),
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
		waitForPageStable(),
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

func waitForPageStable() chromedp.Action {
	return chromedp.Tasks{
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			readyCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			ticker := time.NewTicker(150 * time.Millisecond)
			defer ticker.Stop()
			for {
				var ready bool
				if err := chromedp.Evaluate(`document.readyState === "complete" && !!document.body`, &ready).Do(readyCtx); err == nil && ready {
					return chromedp.Sleep(250 * time.Millisecond).Do(readyCtx)
				}
				select {
				case <-readyCtx.Done():
					return readyCtx.Err()
				case <-ticker.C:
				}
			}
		}),
	}
}

func enableRequestGuard(ctx context.Context, allowLocal bool) error {
	chromedp.ListenTarget(ctx, func(ev any) {
		paused, ok := ev.(*fetch.EventRequestPaused)
		if !ok {
			return
		}
		go func() {
			var action chromedp.Action = fetch.ContinueRequest(paused.RequestID)
			if !isBrowserRequestAllowed(ctx, paused.Request.URL, allowLocal) {
				action = fetch.FailRequest(paused.RequestID, network.ErrorReasonBlockedByClient)
			}
			_ = chromedp.Run(ctx, action)
		}()
	})
	return chromedp.Run(ctx, fetch.Enable().WithPatterns([]*fetch.RequestPattern{{URLPattern: "*"}}))
}

func isBrowserRequestAllowed(ctx context.Context, rawURL string, allowLocal bool) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	switch parsed.Scheme {
	case "", "about", "blob", "data":
		return true
	case "http", "https":
		host := strings.ToLower(parsed.Hostname())
		if allowLocal && (host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "127.0.0.1" || host == "::1") {
			return true
		}
		return validateBrowserHost(ctx, host) == nil
	default:
		return false
	}
}

func validateBrowserHost(ctx context.Context, host string) error {
	if host == "" {
		return fmt.Errorf("empty host")
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if isBlockedBrowserIP(ip) {
			return fmt.Errorf("blocked private ip")
		}
		return nil
	}
	resolver := net.Resolver{}
	addrs, err := resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return err
	}
	if len(addrs) == 0 {
		return fmt.Errorf("no ip addresses resolved")
	}
	for _, addr := range addrs {
		ip, ok := netip.AddrFromSlice(addr)
		if !ok {
			continue
		}
		if isBlockedBrowserIP(ip) {
			return fmt.Errorf("blocked private ip")
		}
	}
	return nil
}

func isBlockedBrowserIP(ip netip.Addr) bool {
	for _, prefix := range blockedBrowserCIDRs {
		if prefix.Contains(ip) {
			return true
		}
	}
	return ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsPrivate() || ip.IsMulticast() || ip.IsUnspecified()
}

func mustPrefixes(values []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			panic(err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}
