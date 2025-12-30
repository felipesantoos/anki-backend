package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/note_type"
)

// ToNoteTypeResponse converts a NoteType domain entity to a NoteTypeResponse DTO
func ToNoteTypeResponse(nt *notetype.NoteType) *response.NoteTypeResponse {
	if nt == nil {
		return nil
	}
	return &response.NoteTypeResponse{
		ID:            nt.GetID(),
		UserID:        nt.GetUserID(),
		Name:          nt.GetName(),
		FieldsJSON:    nt.GetFieldsJSON(),
		CardTypesJSON: nt.GetCardTypesJSON(),
		TemplatesJSON: nt.GetTemplatesJSON(),
		CreatedAt:     nt.GetCreatedAt(),
		UpdatedAt:     nt.GetUpdatedAt(),
	}
}

// ToNoteTypeResponseList converts a list of NoteType domain entities to a list of NoteTypeResponse DTOs
func ToNoteTypeResponseList(noteTypes []*notetype.NoteType) []*response.NoteTypeResponse {
	res := make([]*response.NoteTypeResponse, len(noteTypes))
	for i, nt := range noteTypes {
		res[i] = ToNoteTypeResponse(nt)
	}
	return res
}

