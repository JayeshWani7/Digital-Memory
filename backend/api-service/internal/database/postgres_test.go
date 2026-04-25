package database

import (
	"testing"

	"go.uber.org/zap"

	"github.com/digital-memory/api-service/internal/models"
)

func TestOrderQueryResultsPreservesMatchRanking(t *testing.T) {
	matches := []models.SearchMatch{
		{ID: "doc-3", Score: 0.99},
		{ID: "doc-1", Score: 0.91},
		{ID: "doc-2", Score: 0.65},
	}

	resultByID := map[string]models.QueryResult{
		"doc-1": {ID: "doc-1", SimilarityScore: 0.91, Summary: "first"},
		"doc-2": {ID: "doc-2", SimilarityScore: 0.65, Summary: "second"},
		"doc-3": {ID: "doc-3", SimilarityScore: 0.99, Summary: "third"},
	}

	results := orderQueryResults(matches, resultByID, zap.NewNop())

	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}

	expectedIDs := []string{"doc-3", "doc-1", "doc-2"}
	for i, expectedID := range expectedIDs {
		if results[i].ID != expectedID {
			t.Fatalf("results[%d].ID = %q, want %q", i, results[i].ID, expectedID)
		}
	}
}

func TestOrderQueryResultsSkipsMissingHydratedRows(t *testing.T) {
	matches := []models.SearchMatch{
		{ID: "doc-1", Score: 0.91},
		{ID: "doc-missing", Score: 0.88},
		{ID: "doc-2", Score: 0.65},
	}

	resultByID := map[string]models.QueryResult{
		"doc-1": {ID: "doc-1", SimilarityScore: 0.91},
		"doc-2": {ID: "doc-2", SimilarityScore: 0.65},
	}

	results := orderQueryResults(matches, resultByID, zap.NewNop())

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	if results[0].ID != "doc-1" || results[1].ID != "doc-2" {
		t.Fatalf("unexpected ordered IDs: %+v", []string{results[0].ID, results[1].ID})
	}
}
