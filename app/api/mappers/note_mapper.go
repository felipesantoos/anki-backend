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

// ToFindDuplicatesResponse converts a DuplicateResult domain entity to a FindDuplicatesResponse DTO
func ToFindDuplicatesResponse(result *note.DuplicateResult) *response.FindDuplicatesResponse {
	if result == nil {
		return &response.FindDuplicatesResponse{
			Duplicates: []response.DuplicateGroup{},
			Total:      0,
		}
	}

	groups := make([]response.DuplicateGroup, len(result.Duplicates))
	for i, group := range result.Duplicates {
		notes := make([]response.DuplicateNoteInfo, len(group.Notes))
		for j, n := range group.Notes {
			notes[j] = response.DuplicateNoteInfo{
				ID:        n.ID,
				GUID:      n.GUID,
				DeckID:    n.DeckID,
				CreatedAt: n.CreatedAt,
			}
		}
		groups[i] = response.DuplicateGroup{
			FieldValue: group.FieldValue,
			Notes:      notes,
		}
	}

	return &response.FindDuplicatesResponse{
		Duplicates: groups,
		Total:      result.Total,
	}
}

