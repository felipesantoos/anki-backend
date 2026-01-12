package services

// ITemplateRenderer defines the interface for template rendering functionality
type ITemplateRenderer interface {
	RenderFront(templatesJSON string, cardTypeIndex int, fields map[string]string) (string, error)
	RenderBack(templatesJSON string, cardTypeIndex int, fields map[string]string) (string, error)
}
