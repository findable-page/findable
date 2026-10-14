package audit

import "testing"

func scoredPage(url string, seo, aeo, geo float64) Page {
	return Page{URL: url, Scan: Scan{SEO: Category{Score: &seo}, AEO: Category{Score: &aeo}, GEO: Category{Score: &geo}}}
}

func TestPagesBelowKeepsOnlyPagesUnderTheMinimum(t *testing.T) {
	pages := []Page{
		scoredPage("https://a.test/high", 90, 90, 90),
		scoredPage("https://a.test/edge", 70, 70, 70),
		scoredPage("https://a.test/low", 60, 40, 40),
	}
	got := PagesBelow(pages, 70)
	if len(got) != 1 || got[0].URL != "https://a.test/low" {
		t.Errorf("PagesBelow(70) = %+v, want only /low (overall 50)", got)
	}
}

func TestPagesBelowZeroNeverFails(t *testing.T) {
	if got := PagesBelow([]Page{scoredPage("https://a.test/", 0, 0, 0)}, 0); len(got) != 0 {
		t.Errorf("PagesBelow(0) = %+v, want none", got)
	}
}
