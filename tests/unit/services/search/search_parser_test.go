package search

import (
	"testing"

	searchdomain "github.com/felipesantos/anki-backend/core/domain/services/search"
	"github.com/stretchr/testify/assert"
)

func TestParser_Parse(t *testing.T) {
	parser := searchdomain.NewParser()

	t.Run("Empty query", func(t *testing.T) {
		query, err := parser.Parse("")
		assert.NoError(t, err)
		assert.NotNil(t, query)
	})

	t.Run("Simple text search", func(t *testing.T) {
		query, err := parser.Parse("hello")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.TextSearches, 1)
		assert.Equal(t, "hello", query.TextSearches[0].Text)
	})

	t.Run("Deck filter", func(t *testing.T) {
		query, err := parser.Parse("deck:Default")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.DecksInclude, 1)
		assert.Equal(t, "Default", query.DecksInclude[0])
	})

	t.Run("Tag filter", func(t *testing.T) {
		query, err := parser.Parse("tag:vocabulary")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.TagsInclude, 1)
		assert.Equal(t, "vocabulary", query.TagsInclude[0])
	})

	t.Run("Negated tag filter", func(t *testing.T) {
		query, err := parser.Parse("-tag:marked")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.TagsExclude, 1)
		assert.Equal(t, "marked", query.TagsExclude[0])
	})

	t.Run("Field search", func(t *testing.T) {
		query, err := parser.Parse("front:hello")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Equal(t, "hello", query.FieldSearches["front"])
	})

	t.Run("State filter", func(t *testing.T) {
		query, err := parser.Parse("is:new")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Contains(t, query.States, "new")
	})

	t.Run("Flag filter", func(t *testing.T) {
		query, err := parser.Parse("flag:1")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Contains(t, query.Flags, 1)
	})

	t.Run("Property filter", func(t *testing.T) {
		query, err := parser.Parse("prop:ivl>=10")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.PropertyFilters, 1)
		assert.Equal(t, "ivl", query.PropertyFilters[0].Property)
		assert.Equal(t, ">=", query.PropertyFilters[0].Operator)
		assert.Equal(t, "10", query.PropertyFilters[0].Value)
	})

	t.Run("Exact phrase", func(t *testing.T) {
		query, err := parser.Parse(`"exact phrase"`)
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.TextSearches, 1)
		assert.True(t, query.TextSearches[0].IsExact)
		assert.Equal(t, "exact phrase", query.TextSearches[0].Text)
	})

	t.Run("Wildcard search", func(t *testing.T) {
		query, err := parser.Parse("hello*")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.TextSearches, 1)
		assert.True(t, query.TextSearches[0].IsWildcard)
	})

	t.Run("Regex search", func(t *testing.T) {
		query, err := parser.Parse("re:hello.*world")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.TextSearches, 1)
		assert.True(t, query.TextSearches[0].IsRegex)
		assert.Equal(t, "hello.*world", query.TextSearches[0].Text)
	})

	t.Run("Complex query", func(t *testing.T) {
		query, err := parser.Parse("deck:Default tag:vocabulary front:hello -tag:marked")
		assert.NoError(t, err)
		assert.NotNil(t, query)
		assert.Len(t, query.DecksInclude, 1)
		assert.Len(t, query.TagsInclude, 1)
		assert.Len(t, query.TagsExclude, 1)
		assert.Equal(t, "hello", query.FieldSearches["front"])
	})
}

