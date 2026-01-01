package search

// PropertyFilter represents a property filter (e.g., prop:ivl>=10)
type PropertyFilter struct {
	Property string // "ivl", "due", "lapses", "reps"
	Operator string // ">=", "<=", ">", "<", "="
	Value    string // "10", "-1", "3"
}

// TextSearch represents a text search with optional modifiers
type TextSearch struct {
	Text         string // The search text
	IsExact      bool   // "exact phrase"
	IsWildcard   bool   // Contains * or _
	IsRegex      bool   // re:pattern
	IsNoCombining bool  // nc:text (ignore accents)
	IsWordBoundary bool // w:text (word boundary)
	IsNegated    bool   // -text (negation)
	Field        string // Optional field name (front:, back:, field:name:)
}

// SearchQuery represents a parsed Anki search query
type SearchQuery struct {
	// Field searches: map[fieldName]searchText
	// Examples: "front:hello" -> map["front"]="hello"
	//           "field:name:text" -> map["name"]="text"
	FieldSearches map[string]string

	// Tag filters
	TagsInclude []string // tag:vocab
	TagsExclude []string // -tag:marked

	// Deck filters
	DecksInclude []string // deck:Default
	DecksExclude []string // -deck:Filtered

	// Card state filters
	States []string // is:new, is:due, is:review, is:learn, is:suspended, is:buried, is:marked

	// Flag filters
	Flags []int // flag:1, flag:2, etc. (0-7)

	// Property filters
	PropertyFilters []PropertyFilter // prop:ivl>=10, prop:due=-1, etc.

	// Text searches (general text search in all fields)
	TextSearches []TextSearch

	// Operators and grouping (for complex queries)
	// This is a simplified representation - full AST would be more complex
	HasOR bool // Contains OR operator
	HasGrouping bool // Contains parentheses
}

// NewSearchQuery creates a new empty SearchQuery
func NewSearchQuery() *SearchQuery {
	return &SearchQuery{
		FieldSearches:  make(map[string]string),
		TagsInclude:    []string{},
		TagsExclude:    []string{},
		DecksInclude:   []string{},
		DecksExclude:   []string{},
		States:         []string{},
		Flags:          []int{},
		PropertyFilters: []PropertyFilter{},
		TextSearches:   []TextSearch{},
		HasOR:          false,
		HasGrouping:    false,
	}
}

