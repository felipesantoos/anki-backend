package search

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Parser parses Anki search syntax into a SearchQuery
type Parser struct{}

// NewParser creates a new search parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses an Anki search query string into a SearchQuery
func (p *Parser) Parse(query string) (*SearchQuery, error) {
	if query == "" {
		return NewSearchQuery(), nil
	}

	sq := NewSearchQuery()

	// Check for grouping (parentheses)
	if strings.Contains(query, "(") || strings.Contains(query, ")") {
		sq.HasGrouping = true
	}

	// Check for OR operator
	if strings.Contains(strings.ToLower(query), " or ") {
		sq.HasOR = true
	}

	// Tokenize the query - split by spaces but preserve quoted strings
	tokens := p.tokenize(query)

	// Process each token
	for _, token := range tokens {
		if err := p.processToken(token, sq); err != nil {
			return nil, fmt.Errorf("failed to parse token '%s': %w", token, err)
		}
	}

	return sq, nil
}

// tokenize splits the query into tokens, preserving quoted strings
func (p *Parser) tokenize(query string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for i, r := range query {
		if r == '"' {
			if inQuotes {
				// End of quoted string
				if current.Len() > 0 {
					tokens = append(tokens, `"`+current.String()+`"`)
					current.Reset()
				}
				inQuotes = false
			} else {
				// Start of quoted string
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				inQuotes = true
			}
		} else if r == ' ' && !inQuotes {
			// Space outside quotes - end of token
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}

		// Handle last token
		if i == len(query)-1 && current.Len() > 0 {
			tokens = append(tokens, current.String())
		}
	}

	// Filter out empty tokens and operators
	var filtered []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token != "" && strings.ToUpper(token) != "AND" {
			filtered = append(filtered, token)
		}
	}

	return filtered
}

// processToken processes a single token and updates the SearchQuery
func (p *Parser) processToken(token string, sq *SearchQuery) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}

	// Handle negation
	isNegated := strings.HasPrefix(token, "-")
	if isNegated {
		token = token[1:]
	}

	// Handle quoted strings (exact phrases)
	if strings.HasPrefix(token, `"`) && strings.HasSuffix(token, `"`) {
		text := strings.Trim(token, `"`)
		sq.TextSearches = append(sq.TextSearches, TextSearch{
			Text:      text,
			IsExact:   true,
			IsNegated: isNegated,
		})
		return nil
	}

	// Handle special prefixes: re:, nc:, w: (must be checked before field searches)
	if strings.HasPrefix(token, "re:") {
		pattern := strings.TrimPrefix(token, "re:")
		sq.TextSearches = append(sq.TextSearches, TextSearch{
			Text:      pattern,
			IsRegex:   true,
			IsNegated: isNegated,
		})
		return nil
	}

	if strings.HasPrefix(token, "nc:") {
		text := strings.TrimPrefix(token, "nc:")
		sq.TextSearches = append(sq.TextSearches, TextSearch{
			Text:          text,
			IsNoCombining: true,
			IsNegated:     isNegated,
		})
		return nil
	}

	if strings.HasPrefix(token, "w:") {
		text := strings.TrimPrefix(token, "w:")
		sq.TextSearches = append(sq.TextSearches, TextSearch{
			Text:          text,
			IsWordBoundary: true,
			IsNegated:     isNegated,
		})
		return nil
	}

	// Handle field searches: field:value or field:name:value
	if strings.Contains(token, ":") {
		return p.processFieldToken(token, isNegated, sq)
	}

	// Handle wildcards
	if strings.Contains(token, "*") || strings.Contains(token, "_") {
		sq.TextSearches = append(sq.TextSearches, TextSearch{
			Text:        token,
			IsWildcard:  true,
			IsNegated:   isNegated,
		})
		return nil
	}

	// Regular text search
	sq.TextSearches = append(sq.TextSearches, TextSearch{
		Text:      token,
		IsNegated: isNegated,
	})

	return nil
}

// processFieldToken processes a field search token (e.g., deck:name, tag:vocab, front:text)
func (p *Parser) processFieldToken(token string, isNegated bool, sq *SearchQuery) error {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid field token format: %s", token)
	}

	field := strings.ToLower(parts[0])
	value := parts[1]

	// Remove quotes from value if present
	value = strings.Trim(value, `"`)

	switch field {
	case "deck":
		if isNegated {
			sq.DecksExclude = append(sq.DecksExclude, value)
		} else {
			sq.DecksInclude = append(sq.DecksInclude, value)
		}

	case "tag":
		if isNegated {
			sq.TagsExclude = append(sq.TagsExclude, value)
		} else {
			sq.TagsInclude = append(sq.TagsInclude, value)
		}

	case "is":
		// State filter: is:new, is:due, is:review, etc.
		state := strings.ToLower(value)
		validStates := map[string]bool{
			"new":       true,
			"due":       true,
			"review":    true,
			"learn":     true,
			"suspended": true,
			"buried":    true,
			"marked":    true,
		}
		if validStates[state] {
			sq.States = append(sq.States, state)
		}

	case "flag":
		// Flag filter: flag:1, flag:2, etc.
		flagNum, err := strconv.Atoi(value)
		if err != nil || flagNum < 0 || flagNum > 7 {
			return fmt.Errorf("invalid flag number: %s (must be 0-7)", value)
		}
		sq.Flags = append(sq.Flags, flagNum)

	case "prop":
		// Property filter: prop:ivl>=10, prop:due=-1, etc.
		propFilter, err := p.parsePropertyFilter(value)
		if err != nil {
			return fmt.Errorf("invalid property filter: %w", err)
		}
		sq.PropertyFilters = append(sq.PropertyFilters, propFilter)

	case "front", "back":
		// Field search: front:text, back:text
		// Check if value starts with "re:" for regex search
		if strings.HasPrefix(value, "re:") {
			pattern := strings.TrimPrefix(value, "re:")
			sq.TextSearches = append(sq.TextSearches, TextSearch{
				Text:      pattern,
				IsRegex:   true,
				Field:     field,
				IsNegated: isNegated,
			})
			return nil
		}
		sq.FieldSearches[field] = value

	default:
		// Generic field search: field:name:text
		// Check if it's field:name:value format
		if strings.Contains(value, ":") {
			fieldParts := strings.SplitN(value, ":", 2)
			fieldName := fieldParts[0]
			fieldValue := fieldParts[1]
			// Check if fieldValue starts with "re:" for regex search
			if strings.HasPrefix(fieldValue, "re:") {
				pattern := strings.TrimPrefix(fieldValue, "re:")
				sq.TextSearches = append(sq.TextSearches, TextSearch{
					Text:      pattern,
					IsRegex:   true,
					Field:     fieldName,
					IsNegated: isNegated,
				})
				return nil
			}
			sq.FieldSearches[fieldName] = fieldValue
		} else {
			// Check if value starts with "re:" for regex search
			if strings.HasPrefix(value, "re:") {
				pattern := strings.TrimPrefix(value, "re:")
				sq.TextSearches = append(sq.TextSearches, TextSearch{
					Text:      pattern,
					IsRegex:   true,
					Field:     field,
					IsNegated: isNegated,
				})
				return nil
			}
			// Treat as field name
			sq.FieldSearches[field] = value
		}
	}

	return nil
}

// parsePropertyFilter parses a property filter (e.g., ivl>=10, due=-1)
func (p *Parser) parsePropertyFilter(value string) (PropertyFilter, error) {
	// Match patterns like: ivl>=10, due=-1, lapses>3, reps<10
	re := regexp.MustCompile(`^(\w+)(>=|<=|>|<|=)(-?\d+)$`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 4 {
		return PropertyFilter{}, fmt.Errorf("invalid property filter format: %s", value)
	}

	property := matches[1]
	operator := matches[2]
	val := matches[3]

	validProperties := map[string]bool{
		"ivl":    true,
		"due":    true,
		"lapses": true,
		"reps":   true,
	}

	if !validProperties[property] {
		return PropertyFilter{}, fmt.Errorf("invalid property: %s (valid: ivl, due, lapses, reps)", property)
	}

	return PropertyFilter{
		Property: property,
		Operator: operator,
		Value:    val,
	}, nil
}

