package audit

import "testing"

// ldWebPageOnly is a JSON-LD block with no entity signals; shared with other
// JSON-LD tests to keep the goconst threshold happy.
const ldWebPageOnly = `{"@type":"WebPage"}`

func checkState(t *testing.T, ld, id string) (pass, applicable bool) {
	t.Helper()
	check := findCheck(analyzeLD(t, ld, "").GEO.Checks, id)
	if check == nil {
		t.Fatalf("%s: check missing for %s", id, ld)
	}
	return check.Pass, check.Applicable == nil || *check.Applicable
}

func TestGEOSchemaKnowsAbout(t *testing.T) {
	cases := map[string]struct {
		pass, applicable bool
	}{
		`{"@type":"Organization","knowsAbout":["SEO","AEO","GEO"]}`:                                                       {true, true},
		`{"@type":"WebSite","knowsAbout":["SEO","AEO","GEO"]}`:                                                            {true, true},
		`{"@type":"Organization","knowsAbout":["SEO","AEO"]}`:                                                             {false, true},
		`{"@type":"Organization","knowsAbout":"SEO"}`:                                                                     {false, true},
		`{"@type":"LocalBusiness","knowsAbout":["SEO","AEO","GEO"]}`:                                                      {true, true},
		`{"@graph":[{"@type":"Organization","knowsAbout":["SEO","AEO","GEO"]}]}`:                                          {true, true},
		`{"@type":"Organization","knowsAbout":["SEO",{"@type":"Thing","name":"AEO"},"GEO"]}`:                              {true, true},
		`{"@type":"Organization","knowsAbout":["SEO",{"@type":"Thing"},"GEO"]}`:                                           {false, true},
		`{"@type":"Organization","knowsAbout":["SEO","AEO","GEO"],"department":{"@type":"Organization","knowsAbout":[]}}`: {true, true},
		`{"@type":"Article","knowsAbout":["SEO","AEO","GEO"]}`:                                                            {true, false},
		ldWebPageOnly: {true, false},
	}
	for ld, want := range cases {
		pass, applicable := checkState(t, ld, "geo_schema_knows_about")
		if pass != want.pass || applicable != want.applicable {
			t.Errorf("geo_schema_knows_about for %s = (pass=%v applicable=%v), want (pass=%v applicable=%v)", ld, pass, applicable, want.pass, want.applicable)
		}
	}
}

func TestGEOAuthorCredentialDepth(t *testing.T) {
	cases := map[string]struct {
		pass, applicable bool
	}{
		// Top-level author Person nodes.
		`{"@type":"Person","name":"A","jobTitle":"Head of Research"}`:                      {true, true},
		`{"@type":"Person","name":"A","alumniOf":{"@type":"College","name":"MIT"}}`:        {true, true},
		`{"@type":"Person","name":"A","hasCredential":"PhD"}`:                              {true, true},
		`{"@type":"Person","name":"A","sameAs":["https://orcid.org/0000-0002-1825-0097"]}`: {true, true},
		`{"@type":"Person","name":"A","sameAs":["https://uk.linkedin.com/in/a"]}`:          {true, true},
		`{"@type":"Person","name":"A"}`:                                                    {false, true},
		`{"@type":"Person","name":"A","sameAs":["https://example.com/a"]}`:                 {false, true},
		// Author Person nested in Article schema.
		`{"@type":"Article","author":{"@type":"Person","name":"A","jobTitle":"Editor"}}`:                            {true, true},
		`{"@type":"Article","author":[{"@type":"Person","name":"A"},{"@type":"Person","name":"B","jobTitle":"X"}]}`: {true, true},
		`{"@type":"Article","author":{"@type":"Person","name":"A"}}`:                                                {false, true},
		// String author has no Person schema to hold credentials.
		`{"@type":"Article","author":"Jane Doe"}`: {true, false},
		// Person only in a non-author position is not an author candidate.
		`{"@type":"Organization","founder":{"@type":"Person","name":"F","jobTitle":"CEO"}}`: {true, false},
		ldWebPageOnly: {true, false},
	}
	for ld, want := range cases {
		pass, applicable := checkState(t, ld, "geo_author_credential_depth")
		if pass != want.pass || applicable != want.applicable {
			t.Errorf("geo_author_credential_depth for %s = (pass=%v applicable=%v), want (pass=%v applicable=%v)", ld, pass, applicable, want.pass, want.applicable)
		}
	}
}

func TestKnowsAboutLabelCountsTopics(t *testing.T) {
	check := findCheck(analyzeLD(t, `{"@type":"Organization","knowsAbout":["SEO","AEO"]}`, "").GEO.Checks, "geo_schema_knows_about")
	if want := "knowsAbout lists 2 topic(s) on Organization or WebSite schema (3+ recommended)"; check.Label != want {
		t.Errorf("label = %q, want %q", check.Label, want)
	}
}
