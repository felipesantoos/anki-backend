package entities
import (
	"github.com/felipesantos/anki-backend/core/domain/entities/media"
)

import (
	"testing"
	"time"
)

func TestMedia_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		media    *media.Media
		expected bool
	}{
		{
			name: "active media",
			media: func() *media.Media {
				m := &media.Media{}
				m.SetDeletedAt(nil)
				return m
			}(),
			expected: true,
		},
		{
			name: "deleted media",
			media: func() *media.Media {
				m := &media.Media{}
				m.SetDeletedAt(timePtr(time.Now()))
				return m
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.media.IsActive()
			if got != tt.expected {
				t.Errorf("Media.IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMedia_GetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "image file",
			filename: "photo.jpg",
			expected: ".jpg",
		},
		{
			name:     "audio file",
			filename: "sound.mp3",
			expected: ".mp3",
		},
		{
			name:     "video file",
			filename: "video.mp4",
			expected: ".mp4",
		},
		{
			name:     "uppercase extension",
			filename: "IMAGE.PNG",
			expected: ".png",
		},
		{
			name:     "no extension",
			filename: "file",
			expected: "",
		},
		{
			name:     "multiple dots",
			filename: "file.name.ext",
			expected: ".ext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &media.Media{}
			m.SetFilename(tt.filename)
			got := m.GetFileExtension()
			if got != tt.expected {
				t.Errorf("Media.GetFileExtension() = %v, want %v", got, tt.expected)
			}
		})
	}
}


