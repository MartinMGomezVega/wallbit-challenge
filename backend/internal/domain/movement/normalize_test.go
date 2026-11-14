package movement

import "testing"

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace and punctuation",
			input:    "   ... --- ,,,   ",
			expected: "",
		},
		{
			name:     "basic lowercase without changes",
			input:    "northwind labs",
			expected: "northwind labs",
		},
		{
			name:     "accents and uppercase",
			input:    "Martín Álvarez Ñandú",
			expected: "martin alvarez nandu",
		},
		{
			name:     "umlaut and diacritics",
			input:    "Müller & Söhne",
			expected: "muller sohne",
		},
		{
			name:     "corporate suffix LLC with double spaces",
			input:    "  Northwind   Labs   LLC  ",
			expected: "northwind labs",
		},
		{
			name:     "corporate suffix Inc with dot and comma",
			input:    "Acme Digital, Inc.",
			expected: "acme digital",
		},
		{
			name:     "corporate suffix S.A. with dots",
			input:    "Globant S.A.",
			expected: "globant",
		},
		{
			name:     "corporate suffix S.R.L.",
			input:    "MercadoLibre S.R.L.",
			expected: "mercadolibre",
		},
		{
			name:     "corporate suffix GmbH",
			input:    "Brightloop GmbH",
			expected: "brightloop",
		},
		{
			name:     "corporate suffix Ltd",
			input:    "Kiwi Studio Ltd",
			expected: "kiwi studio",
		},
		{
			name:     "corporate suffix Corp",
			input:    "Orbital SaaS Corp",
			expected: "orbital saas",
		},
		{
			name:     "corporate suffix Co",
			input:    "Alpha Beta Co.",
			expected: "alpha beta",
		},
		{
			name:     "compound suffixes Co Ltd",
			input:    "Omega Trading Co., Ltd.",
			expected: "omega trading",
		},
		{
			name:     "word containing suffix without word boundary is preserved",
			input:    "Costco Wholesale",
			expected: "costco wholesale",
		},
		{
			name:     "word ending in suffix letters is preserved",
			input:    "Corpo Salina",
			expected: "corpo salina",
		},
		{
			name:     "name is only the suffix",
			input:    "Inc",
			expected: "inc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeName(tc.input)
			if got != tc.expected {
				t.Errorf("NormalizeName(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}
