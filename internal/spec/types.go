package spec

type DesignSpec struct {
	SchemaVersion   string         `json:"schema_version"`
	SourceURL       string         `json:"source_url"`
	NormalizedURL   string         `json:"normalized_url"`
	Mode            string         `json:"mode"`
	CreatedAt       string         `json:"created_at"`
	Page            Page           `json:"page"`
	DesignTokens    DesignTokens   `json:"design_tokens"`
	Responsive      ResponsiveSpec `json:"responsive"`
	Assets          AssetPolicy    `json:"assets"`
	GenerationRules []string       `json:"generation_rules"`
	RawHTMLPath     string         `json:"raw_html_path,omitempty"`
}

type Page struct {
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Language       string        `json:"language"`
	ContentSummary string        `json:"content_summary"`
	Structure      PageStructure `json:"structure"`
}

type PageStructure struct {
	Landmarks  []Landmark     `json:"landmarks"`
	Headings   []Heading      `json:"headings"`
	Navigation []NavItem      `json:"navigation"`
	Sections   []Section      `json:"sections"`
	Forms      []FormSummary  `json:"forms"`
	Links      []LinkSummary  `json:"links"`
	Images     []ImageSummary `json:"images"`
}

type Landmark struct {
	Tag  string `json:"tag"`
	Role string `json:"role,omitempty"`
	Text string `json:"text,omitempty"`
}

type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

type NavItem struct {
	Text string `json:"text"`
	Href string `json:"href,omitempty"`
}

type Section struct {
	Kind     string   `json:"kind"`
	Heading  string   `json:"heading,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Summary  string   `json:"summary,omitempty"`
	Repeated bool     `json:"repeated,omitempty"`
}

type FormSummary struct {
	Action string      `json:"action,omitempty"`
	Method string      `json:"method,omitempty"`
	Fields []FormField `json:"fields,omitempty"`
}

type FormField struct {
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Label string `json:"label,omitempty"`
}

type LinkSummary struct {
	Text string `json:"text"`
	Href string `json:"href,omitempty"`
}

type ImageSummary struct {
	Src    string `json:"src,omitempty"`
	Alt    string `json:"alt,omitempty"`
	Width  string `json:"width,omitempty"`
	Height string `json:"height,omitempty"`
}

type DesignTokens struct {
	Colors     ColorTokens      `json:"colors"`
	Typography TypographyTokens `json:"typography"`
	Spacing    []string         `json:"spacing"`
	Radii      []string         `json:"radii"`
	Shadows    []string         `json:"shadows"`
	Layout     LayoutTokens     `json:"layout"`
}

type ColorTokens struct {
	Background []string `json:"background"`
	Text       []string `json:"text"`
	Accent     []string `json:"accent"`
	Border     []string `json:"border"`
}

type TypographyTokens struct {
	FontFamilies []string `json:"font_families"`
	FontSizes    []string `json:"font_sizes"`
	FontWeights  []string `json:"font_weights"`
	LineHeights  []string `json:"line_heights"`
}

type LayoutTokens struct {
	ContainerWidths []string `json:"container_widths"`
	GridPatterns    []string `json:"grid_patterns"`
	FlexPatterns    []string `json:"flex_patterns"`
}

type ResponsiveSpec struct {
	Desktop ViewportAnalysis `json:"desktop"`
	Tablet  ViewportAnalysis `json:"tablet"`
	Mobile  ViewportAnalysis `json:"mobile"`
}

type ViewportAnalysis struct {
	Screenshot string   `json:"screenshot"`
	Notes      []string `json:"notes"`
}

type AssetPolicy struct {
	Policy             string       `json:"policy"`
	AllowedOwnedAssets bool         `json:"allowed_owned_assets"`
	Images             []AssetEntry `json:"images"`
	Fonts              []AssetEntry `json:"fonts"`
}

type AssetEntry struct {
	URL       string `json:"url"`
	LocalPath string `json:"local_path,omitempty"`
	MimeType  string `json:"mime_type,omitempty"`
	Allowed   bool   `json:"allowed"`
	Reason    string `json:"reason,omitempty"`
}
