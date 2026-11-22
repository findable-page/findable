package audit

import "testing"

func TestGeoSchemaKnowsAbout(t *testing.T) {
	cases := map[string]struct {
		ld      string
		want    bool
		skipped bool
	}{
		"org with three topics":     {`{"@type":"Organization","name":"Acme","knowsAbout":["SEO","AEO","Structured Data"]}`, true, false},
		"localbusiness topics":      {`{"@type":"LocalBusiness","knowsAbout":["Bread","Baking","Sourdough"]}`, true, false},
		"topics inside @graph":      {`{"@graph":[{"@type":"WebSite","knowsAbout":["A","B","C"]}]}`, true, false},
		"website with three topics": {`{"@type":"WebSite","name":"Acme","knowsAbout":["SEO","AEO","GEO"]}`, true, false},
		"only two topics":           {`{"@type":"Organization","knowsAbout":["SEO","AEO"]}`, false, false},
		"empty entries ignored":     {`{"@type":"Organization","knowsAbout":["SEO","","AEO","  "]}`, false, false},
		"no knowsAbout":             {`{"@type":"Organization","name":"Acme"}`, false, false},
		"unrelated type":            {`{"@type":"Product","knowsAbout":["A","B","C"]}`, false, true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			check := findCheck(analyzeLD(t, tc.ld, "").GEO.Checks, "geo_schema_knows_about")
			if tc.skipped {
				if check == nil || check.Applicable == nil || *check.Applicable {
					t.Fatalf("expected skip, got %+v", check)
				}
				return
			}
			if check == nil || check.Pass != tc.want {
				t.Errorf("geo_schema_knows_about = %+v, want pass=%v", check, tc.want)
			}
		})
	}
}

func TestGeoAuthorCredentialDepth(t *testing.T) {
	jobTitleLD := `{"@type":"Article","author":{"@type":"Person","name":"Jane","jobTitle":"Head of Search"}}`
	linkedin := `{"@type":"Article","author":{"@type":"Person","name":"Jane","sameAs":["https://www.linkedin.com/in/jane"]}}`
	plain := `{"@type":"Article","author":{"@type":"Person","name":"Jane"}}`
	stringAuthor := `{"@type":"Article","author":"Jane"}`
	cases := map[string]struct {
		ld      string
		want    bool
		skipped bool
	}{
		"jobTitle counts":            {jobTitleLD, true, false},
		"professional sameAs counts": {linkedin, true, false},
		"no credential props":        {plain, false, false},
		"string author has none":     {stringAuthor, false, false},
		"author inside @graph":       {`{"@graph":[{"@type":"Article","author":{"@type":"Person","alumniOf":{"@type":"College"}}}]}`, true, false},
		"no article schema":          {`{"@type":"Organization"}`, false, true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			check := findCheck(analyzeLD(t, tc.ld, "").GEO.Checks, "geo_author_credential_depth")
			if tc.skipped {
				if check == nil || check.Applicable == nil || *check.Applicable {
					t.Fatalf("expected skip, got %+v", check)
				}
				return
			}
			if check == nil || check.Pass != tc.want {
				t.Errorf("geo_author_credential_depth = %+v, want pass=%v", check, tc.want)
			}
		})
	}
}

func TestGeoAuthorCredentialSameAsHostMatch(t *testing.T) {
	cases := map[string]bool{
		`{"@type":"Article","author":{"@type":"Person","sameAs":["https://www.linkedin.com/in/jane"]}}`:      true,
		`{"@type":"Article","author":{"@type":"Person","sameAs":["https://orcid.org/0000-0002-1825-0097"]}}`: true,
		`{"@type":"Article","author":{"@type":"Person","sameAs":["https://evil.com/linkedin.com"]}}`:         false,
		`{"@type":"Article","author":{"@type":"Person","sameAs":["https://example.com/jane"]}}`:              false,
	}
	for ld, want := range cases {
		check := findCheck(analyzeLD(t, ld, "").GEO.Checks, "geo_author_credential_depth")
		if check == nil || check.Pass != want {
			t.Errorf("credential sameAs host = %+v, want pass=%v for %s", check, want, ld)
		}
	}
}

func TestGeoAuthorCredentialAnyAuthorCounts(t *testing.T) {
	ld := `{"@type":"Article","author":[{"@type":"Person","name":"Jane"},{"@type":"Person","name":"Bob","hasCredential":"PhD"}]}`
	check := findCheck(analyzeLD(t, ld, "").GEO.Checks, "geo_author_credential_depth")
	if check == nil || !check.Pass {
		t.Errorf("one credentialed author of two should pass, got %+v", check)
	}
}
