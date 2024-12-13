package deploymentcontext

import (
	"fmt"
	"strconv"
	"strings"

	deploymentpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
)

// Database represents a database connection based on a DSN
// It holds a private field for the database which is accessible through moduleCtx.GetDatabase(name)
type Database struct {
	DSN      string
	DBType   DBType
	isTestDB bool
}

// NewDatabase creates a Database that can be added to DeploymentContext
func NewDatabase(dbType DBType, dsn string) (Database, error) {
	return Database{
		DSN:    dsn,
		DBType: dbType,
	}, nil
}

// NewTestDatabase creates a Database that can be added to DeploymentContext
//
// Test databases can be used within module tests
func NewTestDatabase(dbType DBType, dsn string) (Database, error) {
	db, err := NewDatabase(dbType, dsn)
	if err != nil {
		return Database{}, err
	}
	db.isTestDB = true
	return db, nil
}

type DBType deploymentpb.GetDeploymentContextResponse_DbType

const (
	DBTypeUnspecified DBType = DBType(deploymentpb.GetDeploymentContextResponse_DB_TYPE_UNSPECIFIED)
	DBTypePostgres    DBType = DBType(deploymentpb.GetDeploymentContextResponse_DB_TYPE_POSTGRES)
	DBTypeMySQL       DBType = DBType(deploymentpb.GetDeploymentContextResponse_DB_TYPE_MYSQL)
)

func DBTypeFromString(dt string) (DBType, error) {
	dt = strings.ToLower(dt)
	if dt == "postgres" {
		return DBTypePostgres, nil
	} else if dt == "mysql" {
		return DBTypeMySQL, nil
	}
	return DBTypeUnspecified, fmt.Errorf("unknown DB type: %s", dt)
}

func (x DBType) String() string {
	switch x {
	case DBTypePostgres:
		return "postgres"
	case DBTypeMySQL:
		return "mysql"
	case DBTypeUnspecified:
		return "unspecified"
	default:
		panic(fmt.Sprintf("unknown DB type: %s", strconv.Itoa(int(x))))
	}
}
