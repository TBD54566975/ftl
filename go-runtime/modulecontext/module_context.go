package modulecontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
)

type MockVerb func(ctx context.Context, req any) (resp any, err error)

type Database struct {
	DSN    string
	DBType DBType

	db *sql.DB
}

// NewDatabase creates a Database that can be added to ModuleContext
func NewDatabase(dbType DBType, dsn string) (Database, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return Database{}, err
	}
	return Database{
		DSN:    dsn,
		DBType: dbType,
		db:     db,
	}, nil
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
	module    string
	configs   map[string][]byte
	secrets   map[string][]byte
	databases map[string]Database

	isTesting               bool
	mockVerbs               map[schema.RefKey]MockVerb
	allowDirectVerbBehavior bool
}

type contextKeyModuleContext struct{}

// New creates a new blank ModuleContext for the given module.
func New(module string) ModuleContext {
	return ModuleContext{
		module:    module,
		configs:   map[string][]byte{},
		secrets:   map[string][]byte{},
		databases: map[string]Database{},
		mockVerbs: map[schema.RefKey]MockVerb{},
	}
}

// Update copies a ModuleContext and adds configs, secrets and databases.
func (m ModuleContext) Update(configs map[string][]byte, secrets map[string][]byte, databases map[string]Database) ModuleContext {
	for name, data := range configs {
		m.configs[name] = data
	}
	for name, data := range secrets {
		m.secrets[name] = data
	}
	for name, db := range databases {
		m.databases[name] = db
	}
	return m
}

// UpdateForTesting copies a ModuleContext and marks it as part of a test environment and adds mock verbs and flags for other test features.
func (m ModuleContext) UpdateForTesting(mockVerbs map[schema.RefKey]MockVerb, allowDirectVerbBehavior bool) ModuleContext {
	m.isTesting = true
	for name, verb := range mockVerbs {
		m.mockVerbs[name] = verb
	}
	m.allowDirectVerbBehavior = allowDirectVerbBehavior
	return m
}

// FromContext returns the ModuleContext attached to a context.
func FromContext(ctx context.Context) ModuleContext {
	m, ok := ctx.Value(contextKeyModuleContext{}).(ModuleContext)
	if !ok {
		panic("no ModuleContext in context")
	}
	return m
}

// ApplyToContext returns a Go context.Context with ModuleContext added.
func (m ModuleContext) ApplyToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyModuleContext{}, m)
}

// GetConfig reads a configuration value for the module.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m ModuleContext) GetConfig(name string, value any) error {
	data, ok := m.configs[name]
	if !ok {
		return fmt.Errorf("no config value for %q", name)
	}
	return json.Unmarshal(data, value)
}

// GetSecret reads a secret value for the module.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m ModuleContext) GetSecret(name string, value any) error {
	data, ok := m.secrets[name]
	if !ok {
		return fmt.Errorf("no secret value for %q", name)
	}
	return json.Unmarshal(data, value)
}

// GetDatabase gets a database connection
//
// Returns an error if no database with that name is found or it is not the expected type
func (m ModuleContext) GetDatabase(name string, dbType DBType) (*sql.DB, error) {
	db, ok := m.databases[name]
	if !ok {
		return nil, fmt.Errorf("missing DSN for database %s", name)
	}
	if db.DBType != dbType {
		return nil, fmt.Errorf("database %s does not match expected type of %s", name, dbType)
	}
	return db.db, nil
}

// BehaviorForVerb returns what to do to execute a verb
//
// This allows module context to dictate behavior based on testing options
func (m ModuleContext) BehaviorForVerb(ref schema.Ref) (VerbBehavior, error) {
	if mock, ok := m.mockVerbs[ref.ToRefKey()]; ok {
		return MockBehavior{Mock: mock}, nil
	} else if m.allowDirectVerbBehavior && ref.Module == m.module {
		return DirectBehavior{}, nil
	} else if m.isTesting {
		if ref.Module == m.module {
			return StandardBehavior{}, fmt.Errorf("no mock found: provide a mock with ftltest.WhenVerb(%s, ...) or enable all calls within the module with ftltest.WithCallsAllowedWithinModule()", strings.ToUpper(ref.Name[:1])+ref.Name[1:])
		}
		return StandardBehavior{}, fmt.Errorf("no mock found: provide a mock with ftltest.WhenVerb(%s.%s, ...)", ref.Module, strings.ToUpper(ref.Name[:1])+ref.Name[1:])
	}
	return StandardBehavior{}, nil
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
