package vector_db

import (
	"testing"

	"github.com/digital-memory/api-service/internal/models"
)

func TestSortSearchMatchesOrdersByDescendingScore(t *testing.T) {
	results := []models.SearchMatch{
		{ID: "c", Score: 0.42},
		{ID: "a", Score: 0.98},
		{ID: "b", Score: 0.42},
		{ID: "d", Score: 0.77},
	}

	sortSearchMatches(results)

	expected := []models.SearchMatch{
		{ID: "a", Score: 0.98},
		{ID: "d", Score: 0.77},
		{ID: "b", Score: 0.42},
		{ID: "c", Score: 0.42},
	}

	for i := range expected {
		if results[i] != expected[i] {
			t.Fatalf("result %d = %+v, want %+v", i, results[i], expected[i])
		}
	}
}
