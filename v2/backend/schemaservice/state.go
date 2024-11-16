package schemaservice

import (
	"fmt"
	"slices"
	"sync"

	"github.com/TBD54566975/ftl/internal/reflect"
	"github.com/TBD54566975/ftl/internal/schema"
)

var ErrNotFound = fmt.Errorf("module not found")

type State struct {
	lock   sync.Mutex
	schema *schema.Schema
}

func NewState() *State {
	return &State{
		schema: &schema.Schema{},
	}
}

// Schema returns a copy of the current schema.
func (s *State) Schema() *schema.Schema {
	s.lock.Lock()
	defer s.lock.Unlock()
	return reflect.DeepCopy(s.schema)
}

// Module returns a copy of a module in the schema, if present.
func (s *State) Module(name string) (*schema.Module, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	module, ok := s.schema.Module(name).Get()
	if !ok {
		return nil, false
	}
	return reflect.DeepCopy(module), true
}

// UpsertModule adds or updates a module in the schema, validating the new schema.
func (s *State) UpsertModule(module *schema.Module) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	sch := reflect.DeepCopy(s.schema)
	sch.Modules = slices.DeleteFunc(sch.Modules, func(m *schema.Module) bool { return m.Name == module.Name })
	sch.Modules = append(sch.Modules, module)
	sch, err := schema.ValidateSchema(sch)
	if err != nil {
		return fmt.Errorf("new module is invalid in current schema: %w", err)
	}

	s.schema = sch
	return nil
}

// DeleteModule removes a module from the schema, validating the new schema and returning the module if successful.
//
// If the module is not found, returns an error with ErrNotFound.
func (s *State) DeleteModule(name string) (*schema.Module, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	sch := reflect.DeepCopy(s.schema)
	var module *schema.Module
	for i, m := range sch.Modules {
		if m.Name == name {
			module = m
			sch.Modules = append(sch.Modules[:i], sch.Modules[i+1:]...)
			break
		}
	}
	if module == nil {
		return nil, fmt.Errorf("%s: %w", name, ErrNotFound)
	}
	sch, err := schema.ValidateSchema(sch)
	if err != nil {
		return nil, fmt.Errorf("new module is invalid in current schema: %w", err)
	}

	s.schema = sch
	return module, nil
}
