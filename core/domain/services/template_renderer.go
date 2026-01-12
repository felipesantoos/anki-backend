package services

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
)

// TemplateRenderer provides template rendering functionality for Anki cards
type TemplateRenderer struct{}

// NewTemplateRenderer creates a new TemplateRenderer instance
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

// RenderFront renders the front template of a card.
// It processes field replacements ({{Field}}) and conditional replacements ({{#Field}}...{{/Field}}).
// Returns an empty string if the rendered template is empty or contains only HTML tags and whitespace.
// This is used to determine if a card should be generated for a given note.
func (tr *TemplateRenderer) RenderFront(templatesJSON string, cardTypeIndex int, fields map[string]string) (string, error) {
	frontTemplate, err := tr.getFrontTemplate(templatesJSON, cardTypeIndex)
	if err != nil {
		return "", err
	}

	return tr.renderTemplate(frontTemplate, fields)
}

// RenderBack renders the back template of a card.
// It processes field replacements and conditional replacements.
func (tr *TemplateRenderer) RenderBack(templatesJSON string, cardTypeIndex int, fields map[string]string) (string, error) {
	backTemplate, err := tr.getBackTemplate(templatesJSON, cardTypeIndex)
	if err != nil {
		return "", err
	}

	return tr.renderTemplate(backTemplate, fields)
}

// renderTemplate processes a template string with field replacements and conditionals.
// It returns the rendered content after stripping all HTML tags and trimming leading/trailing whitespace.
// If the final text content is empty (e.g., it only contained tags or spaces), it returns an empty string.
// This "empty" check is critical for Anki's card generation rules.
func (tr *TemplateRenderer) renderTemplate(template string, fields map[string]string) (string, error) {
	if template == "" {
		return "", nil
	}

	result := template

	// Process conditional replacements: {{#Field}}...{{/Field}}
	result = tr.processConditionals(result, fields)

	// Process field replacements: {{FieldName}}
	result = tr.processFieldReplacements(result, fields)

	// Strip HTML tags and check if result is empty
	textContent := tr.stripHTML(result)
	// Replace non-breaking spaces with regular spaces before trimming
	textContent = strings.ReplaceAll(textContent, "\u00a0", " ")
	textContent = strings.TrimSpace(textContent)

	return textContent, nil
}

// processFieldReplacements replaces {{FieldName}} with actual field values
func (tr *TemplateRenderer) processFieldReplacements(template string, fields map[string]string) string {
	result := template

	// Replace each field placeholder
	for fieldName, fieldValue := range fields {
		// Escape field name for regex
		escapedName := regexp.QuoteMeta(fieldName)
		// Match {{FieldName}} or {{FieldName}} with optional whitespace
		pattern := regexp.MustCompile(`\{\{` + escapedName + `\}\}`)
		// Replace with field value
		replacement := fieldValue
		result = pattern.ReplaceAllString(result, replacement)
	}

	return result
}

// processConditionals processes conditional replacements: {{#Field}}...{{/Field}}
// If field exists and is not empty, include the content between tags
// If field is empty or doesn't exist, remove the content
func (tr *TemplateRenderer) processConditionals(template string, fields map[string]string) string {
	result := template

	// Pattern to match {{#FieldName}}...{{/FieldName}}
	// We need to match the opening and closing tags with the same field name
	// This is a simplified implementation - more complex conditionals can be added later
	// We use a two-step approach: first find all opening tags, then match their closing tags
	openingPattern := regexp.MustCompile(`\{\{#(\w+)\}\}`)

	for {
		// Find the first opening tag
		matches := openingPattern.FindStringSubmatchIndex(result)
		if matches == nil {
			break // No more conditional blocks
		}

		fieldName := result[matches[2]:matches[3]]
		startPos := matches[0]
		endPos := matches[1]

		// Find the corresponding closing tag {{/FieldName}}
		closingTag := `{{/` + fieldName + `}}`
		closingPos := strings.Index(result[endPos:], closingTag)
		if closingPos == -1 {
			// No closing tag found, skip this opening tag
			result = result[:startPos] + result[endPos:]
			continue
		}

		closingPos += endPos // Adjust to absolute position
		closingEndPos := closingPos + len(closingTag)

		// Extract content between tags
		content := result[endPos:closingPos]

		// Check if field exists and is not empty
		fieldValue, exists := fields[fieldName]
		if exists && strings.TrimSpace(fieldValue) != "" {
			// Field exists and is not empty, include the content
			result = result[:startPos] + content + result[closingEndPos:]
		} else {
			// Field is empty or doesn't exist, remove the entire conditional block
			result = result[:startPos] + result[closingEndPos:]
		}
	}

	return result
}

// stripHTML removes HTML tags from a string
func (tr *TemplateRenderer) stripHTML(htmlStr string) string {
	// Simple regex-based HTML tag removal
	// This is a basic implementation - for production, consider using a proper HTML parser
	htmlTagPattern := regexp.MustCompile(`<[^>]*>`)
	result := htmlTagPattern.ReplaceAllString(htmlStr, "")
	// Decode HTML entities
	result = html.UnescapeString(result)
	return result
}

// getFrontTemplate extracts the front template for a specific card type from templatesJSON
// templatesJSON is expected to be an array of template objects, one per card type
// Each template object should have "qfmt" (question format) and "afmt" (answer format) fields
func (tr *TemplateRenderer) getFrontTemplate(templatesJSON string, cardTypeIndex int) (string, error) {
	if templatesJSON == "" {
		return "", nil
	}

	// Parse templatesJSON as an array
	var templates []map[string]interface{}
	if err := tr.parseJSONArray(templatesJSON, &templates); err != nil {
		// If parsing as array fails, try parsing as object (single template)
		var templateObj map[string]interface{}
		if err2 := tr.parseJSONObject(templatesJSON, &templateObj); err2 == nil {
			// Single template object
			if qfmt, ok := templateObj["qfmt"].(string); ok {
				return qfmt, nil
			}
		}
		return "", err
	}

	// Check if cardTypeIndex is valid
	if cardTypeIndex < 0 || cardTypeIndex >= len(templates) {
		return "", nil
	}

	template := templates[cardTypeIndex]
	if qfmt, ok := template["qfmt"].(string); ok {
		return qfmt, nil
	}

	// Fallback: if qfmt doesn't exist, try "Front" or empty string
	return "", nil
}

// getBackTemplate extracts the back template for a specific card type from templatesJSON
func (tr *TemplateRenderer) getBackTemplate(templatesJSON string, cardTypeIndex int) (string, error) {
	if templatesJSON == "" {
		return "", nil
	}

	// Parse templatesJSON as an array
	var templates []map[string]interface{}
	if err := tr.parseJSONArray(templatesJSON, &templates); err != nil {
		// If parsing as array fails, try parsing as object (single template)
		var templateObj map[string]interface{}
		if err2 := tr.parseJSONObject(templatesJSON, &templateObj); err2 == nil {
			// Single template object
			if afmt, ok := templateObj["afmt"].(string); ok {
				return afmt, nil
			}
		}
		return "", err
	}

	// Check if cardTypeIndex is valid
	if cardTypeIndex < 0 || cardTypeIndex >= len(templates) {
		return "", nil
	}

	template := templates[cardTypeIndex]
	if afmt, ok := template["afmt"].(string); ok {
		return afmt, nil
	}

	// Fallback: if afmt doesn't exist, return empty string
	return "", nil
}

// parseJSONArray is a helper to parse JSON array
func (tr *TemplateRenderer) parseJSONArray(jsonStr string, result interface{}) error {
	return json.Unmarshal([]byte(jsonStr), result)
}

// parseJSONObject is a helper to parse JSON object
func (tr *TemplateRenderer) parseJSONObject(jsonStr string, result interface{}) error {
	return json.Unmarshal([]byte(jsonStr), result)
}

