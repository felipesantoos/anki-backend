package valueobjects

// ThemeType represents the UI theme preference
type ThemeType string

const (
	// ThemeTypeLight represents a light theme
	ThemeTypeLight ThemeType = "light"
	// ThemeTypeDark represents a dark theme
	ThemeTypeDark ThemeType = "dark"
	// ThemeTypeAuto represents automatic theme based on system preference
	ThemeTypeAuto ThemeType = "auto"
)

// IsValid checks if the theme type is valid
func (t ThemeType) IsValid() bool {
	return t == ThemeTypeLight || t == ThemeTypeDark || t == ThemeTypeAuto
}

// String returns the string representation of the theme type
func (t ThemeType) String() string {
	return string(t)
}

