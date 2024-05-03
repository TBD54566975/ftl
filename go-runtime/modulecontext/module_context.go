package modulecontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver
)

type Ref struct {
	Module string
	Name   string
}

type MockVerb func(ctx context.Context, req any) (resp any, err error)

type dbEntry struct {
	dsn    string
	dbType DBType
	db     *sql.DB
}

type DBType ftlv1.ModuleContextResponse_DBType

const (
	DBTypePostgres = DBType(ftlv1.ModuleContextResponse_POSTGRES)
)

func (x DBType) String() string {
	switch x {
	case DBTypePostgres:
		return "Postgres"
	default:
		panic(fmt.Sprintf("unknown DB type: %s", strconv.Itoa(int(x))))
	}
}

// ModuleContext holds the context needed for a module, including configs, secrets and DSNs
type ModuleContext struct {
	isTesting bool
	configs   map[string][]byte
	secrets   map[string][]byte
	databases map[string]dbEntry
	mockVerbs map[Ref]MockVerb
}

type contextKeyModuleContext struct{}

func New() *ModuleContext {
	return &ModuleContext{
		configs:   map[string][]byte{},
		secrets:   map[string][]byte{},
		databases: map[string]dbEntry{},
		mockVerbs: map[Ref]MockVerb{},
	}
}

func NewForTesting() *ModuleContext {
	moduleCtx := New()
	moduleCtx.isTesting = true
	return moduleCtx
}

func FromContext(ctx context.Context) *ModuleContext {
	m, ok := ctx.Value(contextKeyModuleContext{}).(*ModuleContext)
	if !ok {
		panic("no ModuleContext in context")
	}
	return m
}

// ApplyToContext returns a Go context.Context with ModuleContext added.
func (m *ModuleContext) ApplyToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyModuleContext{}, m)
}

// GetConfig reads a configuration value for the module.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m *ModuleContext) GetConfig(name string, value any) error {
	data, ok := m.configs[name]
	if !ok {
		return fmt.Errorf("no config value for %q", name)
	}
	return json.Unmarshal(data, value)
}

// SetConfig sets a configuration value for the module.
func (m *ModuleContext) SetConfig(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.SetConfigData(name, data)
	return nil
}

// SetConfigData sets a configuration value with raw bytes
func (m *ModuleContext) SetConfigData(name string, data []byte) {
	m.configs[name] = data
}

// GetSecret reads a secret value for the module.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m *ModuleContext) GetSecret(name string, value any) error {
	data, ok := m.secrets[name]
	if !ok {
		return fmt.Errorf("no secret value for %q", name)
	}
	return json.Unmarshal(data, value)
}

// SetSecret sets a secret value for the module.
func (m *ModuleContext) SetSecret(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.SetSecretData(name, data)
	return nil
}

// SetSecretData sets a secret value with raw bytes
func (m *ModuleContext) SetSecretData(name string, data []byte) {
	m.secrets[name] = data
}

// AddDatabase adds a database connection
func (m *ModuleContext) AddDatabase(name string, dbType DBType, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	m.databases[name] = dbEntry{
		dsn:    dsn,
		db:     db,
		dbType: dbType,
	}
	return nil
}

// GetDatabase gets a database connection
//
// Returns an error if no database with that name is found or it is not the expected type
func (m *ModuleContext) GetDatabase(name string, dbType DBType) (*sql.DB, error) {
	entry, ok := m.databases[name]
	if !ok {
		return nil, fmt.Errorf("missing DSN for database %s", name)
	}
	if entry.dbType != dbType {
		return nil, fmt.Errorf("database %s does not match expected type of %s", name, dbType)
	}
	return entry.db, nil
}

// BehaviorForVerb returns what to do to execute a verb
//
// This allows module context to dictate behavior based on testing options
func (m *ModuleContext) BehaviorForVerb(ref Ref) (VerbBehavior, error) {
	if mock, ok := m.mockVerbs[ref]; ok {
		return MockBehavior{Mock: mock}, nil
	}
	// TODO: add logic here for when to do direct behavior
	if m.isTesting {
		return StandardBehavior{}, fmt.Errorf("no mock found")
	}
	return StandardBehavior{}, nil
}

func (m *ModuleContext) SetMockVerb(ref Ref, mock MockVerb) {
	m.mockVerbs[ref] = mock
}

// VerbBehavior indicates how to execute a verb
//
//sumtype:decl
type VerbBehavior interface {
	verbBehavior()
}

// StandardBehavior indicates that the verb should be executed via the controller
type StandardBehavior struct{}

func (StandardBehavior) verbBehavior() {}

var _ VerbBehavior = StandardBehavior{}

// DirectBehavior indicates that the verb should be executed by calling the function directly (for testing)
type DirectBehavior struct{}

func (DirectBehavior) verbBehavior() {}

var _ VerbBehavior = DirectBehavior{}

// MockBehavior indicates the verb has a mock implementation
type MockBehavior struct {
	Mock MockVerb
}

func (MockBehavior) verbBehavior() {}

var _ VerbBehavior = MockBehavior{}
