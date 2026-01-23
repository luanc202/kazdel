package uniqueEntityId

import "github.com/google/uuid"

type ID = uuid.UUID

func NewID() ID {
	return ID(uuid.New())
}

func ParseID(id string) (ID, error) {
	parsedId, err := uuid.Parse(id)
	return parsedId, err
}
