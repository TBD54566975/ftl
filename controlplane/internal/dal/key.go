package dal

import (
	"github.com/oklog/ulid/v2"

	"github.com/TBD54566975/ftl/controlplane/internal/sqltypes"
)

type DeploymentKey ulid.ULID

func (d DeploymentKey) String() string      { return "urn:ftl:deployment:" + ulid.ULID(d).String() }
func (d DeploymentKey) ULID() ulid.ULID     { return ulid.ULID(d) }
func (d DeploymentKey) DBKey() sqltypes.Key { return sqltypes.Key(d) }

type RunnerKey ulid.ULID

func (r RunnerKey) String() string      { return "urn:ftl:runner:" + ulid.ULID(r).String() }
func (r RunnerKey) ULID() ulid.ULID     { return ulid.ULID(r) }
func (r RunnerKey) DBKey() sqltypes.Key { return sqltypes.Key(r) }
