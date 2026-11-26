package vm

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidInput = errors.New("invalid VM input")
	ErrNotFound     = errors.New("virtual machine not found")
	ErrConflict     = errors.New("virtual machine must be stopped before deletion")
)

// UUIDs in resource URLs must use the standard hyphenated hex format.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}package vm

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

)

// CreateInput contains the values a client is allowed to supply.
type CreateInput struct {
	Name     string
	CPUCores int32
	MemoryMB int32
	Region   string
}

// Repository describes the storage operations the service needs.
type Repository interface {
	List(context.Context) ([]VirtualMachine, error)
	Create(context.Context, CreateInput) (VirtualMachine, error)
	Get(context.Context, string) (VirtualMachine, error)
	Delete(context.Context, string) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]VirtualMachine, error) {
	return s.repository.List(ctx)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (VirtualMachine, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Region = strings.TrimSpace(input.Region)

	switch {
	case input.Name == "" || utf8.RuneCountInString(input.Name) > 100 || strings.ContainsRune(input.Name, 0):
		return VirtualMachine{}, fmt.Errorf("%w: name must contain 1 to 100 characters and no NUL", ErrInvalidInput)
	case input.CPUCores < 1 || input.CPUCores > 64:
		return VirtualMachine{}, fmt.Errorf("%w: cpu_cores must be between 1 and 64", ErrInvalidInput)
	case input.MemoryMB < 512 || input.MemoryMB > 262144 || input.MemoryMB%512 != 0:
		return VirtualMachine{}, fmt.Errorf("%w: memory_mb must be between 512 and 262144 in steps of 512", ErrInvalidInput)
	case input.Region == "" || utf8.RuneCountInString(input.Region) > 50 || strings.ContainsRune(input.Region, 0):
		return VirtualMachine{}, fmt.Errorf("%w: region must contain 1 to 50 characters and no NUL", ErrInvalidInput)
	}

	return s.repository.Create(ctx, input)
}

func (s *Service) Get(ctx context.Context, id string) (VirtualMachine, error) {
	if !uuidPattern.MatchString(id) {
		return VirtualMachine{}, fmt.Errorf("%w: id must be a UUID", ErrInvalidInput)
	}
	return s.repository.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if !uuidPattern.MatchString(id) {
		return fmt.Errorf("%w: id must be a UUID", ErrInvalidInput)
	}
	return s.repository.Delete(ctx, id)
}
