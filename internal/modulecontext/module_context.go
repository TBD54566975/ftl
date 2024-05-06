package modulecontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/types/optional"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/reflect"
)

// Database represents a database connection based on a DSN
//
// It holds a private field for the database which is accessible through moduleCtx.GetDatabase(name)
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

// Verb is a function that takes a request and returns a response but is not constrained by request/response type like ftl.Verb
//
// It is used for definitions of mock verbs as well as real implementations of verbs to directly execute
type Verb func(ctx context.Context, req any) (resp any, err error)

// ModuleContext holds the context needed for a module, including configs, secrets and DSNs
//
// ModuleContext is immutable
type ModuleContext struct {
	module    string
	configs   map[string][]byte
	secrets   map[string][]byte
	databases map[string]Database

	isTesting               bool
	mockVerbs               map[schema.RefKey]Verb
	allowDirectVerbBehavior bool
}

// Builder is used to build a ModuleContext
type Builder ModuleContext

type contextKeyModuleContext struct{}

// NewBuilder creates a new blank Builder for the given module.
func NewBuilder(module string) *Builder {
	return &Builder{
		module:    module,
		configs:   map[string][]byte{},
		secrets:   map[string][]byte{},
		databases: map[string]Database{},
		mockVerbs: map[schema.RefKey]Verb{},
	}
}

// AddConfigs adds configuration values (as bytes) to the builder
func (b *Builder) AddConfigs(configs map[string][]byte) *Builder {
	for name, data := range configs {
		b.configs[name] = data
	}
	return b
}

// AddSecrets adds configuration values (as bytes) to the builder
func (b *Builder) AddSecrets(secrets map[string][]byte) *Builder {
	for name, data := range secrets {
		b.secrets[name] = data
	}
	return b
}

// AddDatabases adds databases to the builder
func (b *Builder) AddDatabases(databases map[string]Database) *Builder {
	for name, db := range databases {
		b.databases[name] = db
	}
	return b
}

// UpdateForTesting marks the builder as part of a test environment and adds mock verbs and flags for other test features.
func (b *Builder) UpdateForTesting(mockVerbs map[schema.RefKey]Verb, allowDirectVerbBehavior bool) *Builder {
	b.isTesting = true
	for name, verb := range mockVerbs {
		b.mockVerbs[name] = verb
	}
	b.allowDirectVerbBehavior = allowDirectVerbBehavior
	return b
}

func (b *Builder) Build() ModuleContext {
	return ModuleContext(reflect.DeepCopy(*b))
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
// Returning optional.Nil indicates the verb should be executed normally via the controller
func (m ModuleContext) BehaviorForVerb(ref schema.Ref) (optional.Option[VerbBehavior], error) {
	if mock, ok := m.mockVerbs[ref.ToRefKey()]; ok {
		return optional.Some(VerbBehavior(MockBehavior{Mock: mock})), nil
	} else if m.allowDirectVerbBehavior && ref.Module == m.module {
		return optional.Some(VerbBehavior(DirectBehavior{})), nil
	} else if m.isTesting {
		if ref.Module == m.module {
			return optional.None[VerbBehavior](), fmt.Errorf("no mock found: provide a mock with ftltest.WhenVerb(%s, ...) or enable all calls within the module with ftltest.WithCallsAllowedWithinModule()", strings.ToUpper(ref.Name[:1])+ref.Name[1:])
		}
		return optional.None[VerbBehavior](), fmt.Errorf("no mock found: provide a mock with ftltest.WhenVerb(%s.%s, ...)", ref.Module, strings.ToUpper(ref.Name[:1])+ref.Name[1:])
	}
	return optional.None[VerbBehavior](), nil
}

// VerbBehavior indicates how to execute a verb
type VerbBehavior interface {
	Call(ctx context.Context, verb Verb, request any) (any, error)
}

// DirectBehavior indicates that the verb should be executed by calling the function directly (for testing)
type DirectBehavior struct{}

func (DirectBehavior) Call(ctx context.Context, verb Verb, req any) (any, error) {
	return verb(ctx, req)
}

var _ VerbBehavior = DirectBehavior{}

// MockBehavior indicates the verb has a mock implementation
type MockBehavior struct {
	Mock Verb
}

func (b MockBehavior) Call(ctx context.Context, verb Verb, req any) (any, error) {
	return b.Mock(ctx, req)
}
