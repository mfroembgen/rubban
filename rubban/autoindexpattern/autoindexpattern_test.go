package autoindexpattern

import (
	"context"
	"testing"

	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

type mockAPI struct {
	indices       []kibana.Index
	indexPatterns []kibana.IndexPattern
}

func (m *mockAPI) Info(ctx context.Context) (kibana.Info, error) {
	panic("implement me")
}

func (m *mockAPI) Indices(ctx context.Context, filter string) ([]kibana.Index, error) {
	return m.indices, nil
}

func (m *mockAPI) IndexPatterns(ctx context.Context, filter string, fields []string) ([]kibana.IndexPattern, error) {
	return m.indexPatterns, nil
}

func (m *mockAPI) BulkCreateIndexPattern(ctx context.Context, indexPatterns []kibana.IndexPattern) error {
	panic("implement me")
}

func newMockAPI(indices []kibana.Index, indexPatterns []kibana.IndexPattern) kibana.API {
	return &mockAPI{indices: indices, indexPatterns: indexPatterns}
}

// TestAutoindexPatternMatchers tests how the matchers work.
func TestAutoindexPatternMatchers(t *testing.T) {
	for _, tcase := range []struct {
		generalPattern        string
		indices               []kibana.Index
		indexpatterns         []kibana.IndexPattern
		expectedIndexPatterns []string
		tcaseName             string
	}{
		{
			generalPattern:        "logs-mcoins-analytics-?-*",
			indices:               []kibana.Index{{Name: "logs-mcoins-analytics-writer-2020.02.14"}, {Name: "logs-mcoins-analytics-prediction-persister-2020.02.14"}, {Name: "logs-mcoins-analytics-denormalizer-subscriber-2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{},
			expectedIndexPatterns: []string{"logs-mcoins-analytics-writer-*", "logs-mcoins-analytics-prediction-persister-*", "logs-mcoins-analytics-denormalizer-subscriber-*"},
			tcaseName:             `analytics`,
		},
		{
			generalPattern:        "logs-mcoins-marketing-?-*",
			indices:               []kibana.Index{{Name: "logs-mcoins-marketing-terminal-views-subscriber-2020.06.30"}},
			indexpatterns:         []kibana.IndexPattern{},
			expectedIndexPatterns: []string{"logs-mcoins-marketing-terminal-views-subscriber-*"},
			tcaseName:             `marketing`,
		},
	} {
		autoIdxPttrn := NewAutoIndexPattern(config.AutoIndexPattern{
			Enabled: true,
			GeneralPatterns: []config.GeneralPattern{{
				Pattern:       tcase.generalPattern,
				TimeFieldName: "@timestamp",
			}},
			Schedule: "* * * * *",
		}, newMockAPI(tcase.indices, tcase.indexpatterns), log.Default())

		///
		result := autoIdxPttrn.getIndexPattern(context.Background(), autoIdxPttrn.GeneralPatterns[0])

		t.Run(tcase.tcaseName, func(t *testing.T) {
			if len(tcase.expectedIndexPatterns) == 0 && len(result) != 0 {
				t.Fatalf("expected zero index patterns but got %d (%v)", len(result), result)
			} else {
				if len(tcase.expectedIndexPatterns) != len(result) {
					t.Fatalf("expected %d index patterns but got %d (%v)", len(tcase.expectedIndexPatterns), len(result), result)
				}
				for _, e := range tcase.expectedIndexPatterns {
					_, ok := result[e]
					if !ok {
						t.Fatalf("failed to find index pattern %s (%v)", e, result)
					}
				}
			}
		})

	}
}
