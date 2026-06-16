package refactoringguru

// Pattern is one entry in the GoF design-patterns catalog.
type Pattern struct {
	Rank     int    `json:"rank"     csv:"rank"     tsv:"rank"`
	Slug     string `json:"slug"     csv:"slug"     tsv:"slug"`
	Category string `json:"category" csv:"category" tsv:"category"`
	Title    string `json:"title"    csv:"title"    tsv:"title"`
	URL      string `json:"url"      csv:"url"      tsv:"url"`
}

// Info holds aggregate statistics for the refactoring.guru catalog.
type Info struct {
	TotalPatterns int    `json:"total_patterns"`
	Creational    int    `json:"creational"`
	Structural    int    `json:"structural"`
	Behavioral    int    `json:"behavioral"`
	SiteURL       string `json:"site_url"`
	CatalogURL    string `json:"catalog_url"`
}
