package handlers

// ErrInvalidID represents an error for invalid entity IDs.
type ErrInvalidID struct {
	Description string `json:"description"`
}

// ErrInvalidBody represents an error for invalid request bodies.
type ErrInvalidBody struct {
	Description string `json:"description"`
}

func (e *ErrInvalidID) Error() string {
	return e.Description
}

func (e *ErrInvalidBody) Error() string {
	return e.Description
}
