package dto

// DTO represents objects that will be used to transfer data between layers
type DTO interface {
	Validate() error
}

// ValidateDTO validates a DTO according to the rules defined in the DTO itself
func ValidateDTO(dto DTO) error {
	return dto.Validate()
}
