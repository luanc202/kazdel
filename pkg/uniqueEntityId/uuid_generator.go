package uniqueEntityId

import "github.com/google/uuid"

type ID = uuid.UUID

func NewID() ID {
	id, _ := uuid.NewV7()
	return ID(id)
}

func ParseID(id string) (ID, error) {
	parsedId, err := uuid.Parse(id)
	return parsedId, err
}
