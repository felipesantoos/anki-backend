package mappers

import (
	"github.com/felipesantos/anki-backend/app/api/dtos/response"
	"github.com/felipesantos/anki-backend/core/domain/entities/note"
)

// ToNoteResponse converts a Note domain entity to a NoteResponse DTO
func ToNoteResponse(n *note.Note) *response.NoteResponse {
	if n == nil {
		return nil
	}
	return &response.NoteResponse{
		ID:         n.GetID(),
		GUID:       n.GetGUID().Value(),
		NoteTypeID: n.GetNoteTypeID(),
		FieldsJSON: n.GetFieldsJSON(),
		Tags:       n.GetTags(),
		CreatedAt:  n.GetCreatedAt(),
		UpdatedAt:  n.GetUpdatedAt(),
	}
}

// ToNoteResponseList converts a list of Note domain entities to a list of NoteResponse DTOs
func ToNoteResponseList(notes []*note.Note) []*response.NoteResponse {
	res := make([]*response.NoteResponse, len(notes))
	for i, n := range notes {
		res[i] = ToNoteResponse(n)
	}
	return res
}

