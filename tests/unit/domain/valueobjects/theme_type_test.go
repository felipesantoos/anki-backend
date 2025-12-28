package valueobjects

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestThemeType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		theme valueobjects.ThemeType
		want  bool
	}{
		{
			name:  "valid light",
			theme: valueobjects.ThemeTypeLight,
			want:  true,
		},
		{
			name:  "valid dark",
			theme: valueobjects.ThemeTypeDark,
			want:  true,
		},
		{
			name:  "valid auto",
			theme: valueobjects.ThemeTypeAuto,
			want:  true,
		},
		{
			name:  "invalid theme",
			theme: valueobjects.ThemeType("invalid"),
			want:  false,
		},
		{
			name:  "empty theme",
			theme: valueobjects.ThemeType(""),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.theme.IsValid()
			if got != tt.want {
				t.Errorf("ThemeType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThemeType_String(t *testing.T) {
	if valueobjects.ThemeTypeLight.String() != "light" {
		t.Errorf("ThemeTypeLight.String() = %v, want 'light'", valueobjects.ThemeTypeLight.String())
	}
	if valueobjects.ThemeTypeDark.String() != "dark" {
		t.Errorf("ThemeTypeDark.String() = %v, want 'dark'", valueobjects.ThemeTypeDark.String())
	}
	if valueobjects.ThemeTypeAuto.String() != "auto" {
		t.Errorf("ThemeTypeAuto.String() = %v, want 'auto'", valueobjects.ThemeTypeAuto.String())
	}
}

