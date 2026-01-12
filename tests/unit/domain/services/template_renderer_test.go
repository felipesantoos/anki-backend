package services

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/services"
	"github.com/stretchr/testify/assert"
)

func TestTemplateRenderer_RenderFront(t *testing.T) {
	tr := services.NewTemplateRenderer()

	t.Run("Empty template returns empty", func(t *testing.T) {
		rendered, err := tr.RenderFront(`[{"qfmt": "", "afmt": ""}]`, 0, map[string]string{})
		assert.NoError(t, err)
		assert.Empty(t, rendered)
	})

	t.Run("Simple field replacement", func(t *testing.T) {
		fields := map[string]string{"Front": "Question"}
		rendered, err := tr.RenderFront(`[{"qfmt": "{{Front}}", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "Question", rendered)
	})

	t.Run("Strips HTML tags", func(t *testing.T) {
		fields := map[string]string{"Front": "<b>Question</b>"}
		rendered, err := tr.RenderFront(`[{"qfmt": "<div>{{Front}}</div>", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "Question", rendered)
	})

	t.Run("Trims whitespace", func(t *testing.T) {
		fields := map[string]string{"Front": "  Question  "}
		rendered, err := tr.RenderFront(`[{"qfmt": "  {{Front}}  ", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "Question", rendered)
	})

	t.Run("Empty after stripping HTML and trimming", func(t *testing.T) {
		fields := map[string]string{"Front": "   "}
		rendered, err := tr.RenderFront(`[{"qfmt": "<div><br> {{Front}} </div>", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Empty(t, rendered)
	})

	t.Run("Conditional replacement - field present", func(t *testing.T) {
		fields := map[string]string{"Front": "Q", "Extra": "Detail"}
		rendered, err := tr.RenderFront(`[{"qfmt": "{{Front}}{{#Extra}} ({{Extra}}){{/Extra}}", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "Q (Detail)", rendered)
	})

	t.Run("Conditional replacement - field empty", func(t *testing.T) {
		fields := map[string]string{"Front": "Q", "Extra": ""}
		rendered, err := tr.RenderFront(`[{"qfmt": "{{Front}}{{#Extra}} ({{Extra}}){{/Extra}}", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "Q", rendered)
	})

	t.Run("Conditional replacement - field missing", func(t *testing.T) {
		fields := map[string]string{"Front": "Q"}
		rendered, err := tr.RenderFront(`[{"qfmt": "{{Front}}{{#Extra}} ({{Extra}}){{/Extra}}", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "Q", rendered)
	})

	t.Run("Nested HTML tags and entities", func(t *testing.T) {
		fields := map[string]string{"Front": "A & B"}
		rendered, err := tr.RenderFront(`[{"qfmt": "<div class='test'><span>{{Front}}</span></div>", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "A & B", rendered)
	})

	t.Run("Multiple field types and complex HTML", func(t *testing.T) {
		fields := map[string]string{
			"Expression": "猫",
			"Reading":    "ねこ",
			"Meaning":    "Cat",
		}
		template := `<div class=\"jp\">{{Expression}}</div><hr>{{Reading}}<br>{{#Meaning}}<i>{{Meaning}}</i>{{/Meaning}}`
		rendered, err := tr.RenderFront(`[{"qfmt": "`+template+`", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		// Expected: 猫ねこCat (since all tags and newlines are stripped)
		assert.Equal(t, "猫ねこCat", rendered)
	})

	t.Run("Handles &nbsp; and multiple spaces", func(t *testing.T) {
		fields := map[string]string{"Front": "Question"}
		template := "  <div>  {{Front}}  &nbsp;  </div>  "
		rendered, err := tr.RenderFront(`[{"qfmt": "`+template+`", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "Question", rendered)
	})

	t.Run("Handles self-closing tags and tags with attributes", func(t *testing.T) {
		fields := map[string]string{"Front": "Q"}
		template := `<img src=\"test.jpg\" /> <div style=\"color: red\">{{Front}}</div> <br/>`
		rendered, err := tr.RenderFront(`[{"qfmt": "`+template+`", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Equal(t, "Q", rendered)
	})

	t.Run("Empty with only spaces and &nbsp;", func(t *testing.T) {
		fields := map[string]string{"Front": " &nbsp; "}
		template := `<div> {{Front}} </div>`
		rendered, err := tr.RenderFront(`[{"qfmt": "`+template+`", "afmt": ""}]`, 0, fields)
		assert.NoError(t, err)
		assert.Empty(t, rendered)
	})
}
