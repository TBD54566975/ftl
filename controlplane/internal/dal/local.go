package dal

import (
	"context"
	stdlibsha256 "crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/controlplane/internal/sql"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/schema"
)

var _ DAL = (*Local)(nil)

type localRunner struct {
	Runner
	lastUpdated        time.Time
	reservationTimeout time.Time
}

type localDeployment struct {
	minReplicas       int
	partialDeployment model.Deployment
	artefacts         func() []*model.Artefact
}

func NewLocal(blobStoreDir string) *Local {
	return &Local{
		blobStoreDir:              blobStoreDir,
		modules:                   map[string]*sql.Module{},
		deployments:               map[model.DeploymentKey]*localDeployment{},
		deploymentsByArtefactHash: map[sha256.SHA256]*localDeployment{},
		runners:                   map[model.RunnerKey]*localRunner{},
		runnersByEndpoint:         map[string]*localRunner{},
	}
}

// Local is an in-memory DAL for local development and testing.
type Local struct {
	lock                      sync.Mutex
	blobStoreDir              string
	modules                   map[string]*sql.Module
	deployments               map[model.DeploymentKey]*localDeployment
	deploymentsByArtefactHash map[sha256.SHA256]*localDeployment
	runners                   map[model.RunnerKey]*localRunner
	runnersByEndpoint         map[string]*localRunner
}

func (m *Local) GetRunnersForDeployment(ctx context.Context, deployment model.DeploymentKey) ([]Runner, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.deployments[deployment]; !ok {
		return nil, errors.Wrapf(ErrNotFound, "deployment %q not found", deployment)
	}
	runners := []Runner{}
	for _, runner := range m.runners {
		if depl, ok := runner.Deployment.Get(); ok && depl == deployment {
			if depl == deployment {
				runners = append(runners, runner.Runner)
			}
		}
	}
	return runners, nil
}

func (m *Local) UpsertModule(ctx context.Context, language, name string) (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.modules[name] = &sql.Module{
		Language: language,
		Name:     name,
	}
	return nil
}

func (m *Local) GetMissingArtefacts(ctx context.Context, digests []sha256.SHA256) ([]sha256.SHA256, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	missing := make([]sha256.SHA256, 0, len(digests))
	for _, digest := range digests {
		_, err := os.Stat(m.blobPath(digest))
		if errors.Is(err, os.ErrNotExist) {
			missing = append(missing, digest)
		} else if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return missing, nil
}

func (m *Local) CreateArtefact(ctx context.Context, content []byte) (digest sha256.SHA256, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	fw, err := os.CreateTemp(m.blobStoreDir, "blob-*")
	if err != nil {
		return sha256.SHA256{}, errors.WithStack(err)
	}
	defer fw.Close() //nolint:gosec
	defer os.Remove(fw.Name())

	hw := stdlibsha256.New()
	w := io.MultiWriter(fw, hw)
	_, err = w.Write(content)
	if err != nil {
		return sha256.SHA256{}, errors.WithStack(err)
	}

	err = fw.Close()
	if err != nil {
		return sha256.SHA256{}, errors.WithStack(err)
	}
	digest = sha256.FromBytes(hw.Sum(nil))
	err = os.Rename(fw.Name(), m.blobPath(digest))
	if err != nil {
		return sha256.SHA256{}, errors.WithStack(err)
	}
	return digest, nil
}

func (m *Local) blobPath(digest sha256.SHA256) string {
	return filepath.Join(m.blobStoreDir, digest.String())
}

func (m *Local) CreateDeployment(ctx context.Context, language string, schema *schema.Module, artefacts []DeploymentArtefact) (key model.DeploymentKey, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Check if deployment with identical artefacts already exists.
	h := stdlibsha256.New()
	enc := json.NewEncoder(h)
	_ = enc.Encode(language)
	_ = enc.Encode(schema)
	sort.Slice(artefacts, func(i, j int) bool {
		return artefacts[i].Path < artefacts[j].Path
	})
	for _, artefact := range artefacts {
		_ = enc.Encode(artefact)
	}
	byArtefactKey := sha256.FromBytes(h.Sum(nil))
	if existing, ok := m.deploymentsByArtefactHash[byArtefactKey]; ok {
		return existing.partialDeployment.Key, nil
	}

	// Doesn't exist, create it.
	key = model.NewDeploymentKey()
	m.deployments[key] = &localDeployment{
		partialDeployment: model.Deployment{
			Module:   schema.Name,
			Language: language,
			Key:      key,
			Schema:   schema,
		},
		artefacts: func() []*model.Artefact {
			return slices.Map(artefacts, func(in DeploymentArtefact) *model.Artefact {
				return &model.Artefact{
					Path:       in.Path,
					Executable: in.Executable,
					Digest:     in.Digest,
					Content:    &lazyFileReader{path: m.blobPath(in.Digest)},
				}
			})
		},
	}
	return key, nil
}

func (m *Local) GetDeployment(ctx context.Context, id model.DeploymentKey) (*model.Deployment, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	deployment, ok := m.deployments[id]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, "deployment")
	}
	hydrated := deployment.partialDeployment
	hydrated.Artefacts = deployment.artefacts()
	return &hydrated, nil
}

func (m *Local) UpsertRunner(ctx context.Context, runner Runner) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.runnersByEndpoint[runner.Endpoint]; ok {
		if _, ok := m.runners[runner.Key]; !ok {
			return errors.Wrap(ErrConflict, "runner")
		}
	}
	if dkey, ok := runner.Deployment.Get(); ok {
		if _, ok := m.deployments[dkey]; !ok {
			return errors.Wrap(ErrNotFound, "deployment")
		}
	}
	lr := &localRunner{
		Runner:      runner,
		lastUpdated: time.Now(),
	}
	m.runners[runner.Key] = lr
	m.runnersByEndpoint[runner.Endpoint] = lr
	return nil
}

func (m *Local) DeleteStaleRunners(ctx context.Context, age time.Duration) (int64, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	var count int64
	for _, runner := range m.runners {
		if time.Since(runner.lastUpdated) > age {
			count++
			delete(m.runners, runner.Key)
			delete(m.runnersByEndpoint, runner.Endpoint)
		}
	}
	return count, nil
}

func (m *Local) DeregisterRunner(ctx context.Context, key model.RunnerKey) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.runners[key]
	if !ok {
		return errors.Wrap(ErrNotFound, "runner")
	}
	delete(m.runners, key)
	return nil
}

func (m *Local) ReserveRunnerForDeployment(ctx context.Context, language string, deployment model.DeploymentKey, reservationTimeout time.Duration) (Reservation, error) {
	panic("not implemented")
	m.lock.Lock() //nolint:govet
	defer m.lock.Unlock()
	if _, ok := m.deployments[deployment]; !ok {
		return nil, errors.Wrap(ErrNotFound, "deployment")
	}
	for _, runner := range m.runners {
		if runner.Language == language && runner.State == RunnerStateIdle {
			runner.State = RunnerStateReserved
			runner.reservationTimeout = time.Now().Add(reservationTimeout)
			return &localClaim{runner: runner, lock: &m.lock}, nil
		}
	}
	return nil, errors.Wrap(ErrNotFound, "no idle runners found")
}

var _ Reservation = &localClaim{}

type localClaim struct {
	lock   *sync.Mutex
	runner *localRunner
}

func (l *localClaim) Runner() Runner                 { return l.runner.Runner }
func (l *localClaim) Commit(context.Context) error   { panic("not implemented") }
func (l *localClaim) Rollback(context.Context) error { panic("not implemented") }

func (m *Local) SetDeploymentReplicas(ctx context.Context, key model.DeploymentKey, minReplicas int) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	deployment, ok := m.deployments[key]
	if !ok {
		return errors.Wrap(ErrNotFound, "deployment")
	}
	deployment.minReplicas = minReplicas
	return nil
}

func (m *Local) GetDeploymentsNeedingReconciliation(ctx context.Context) ([]Reconciliation, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	out := []Reconciliation{}
	deploymentCounts := map[model.DeploymentKey]int{}
	for _, runner := range m.runners {
		if runner.State == RunnerStateAssigned {
			if dkey, ok := runner.Deployment.Get(); ok {
				deploymentCounts[dkey]++
			}
		}
	}
	for _, deployment := range m.deployments {
		assignedReplicas := deploymentCounts[deployment.partialDeployment.Key]
		if deployment.minReplicas != assignedReplicas {
			out = append(out, Reconciliation{
				Deployment:       deployment.partialDeployment.Key,
				Module:           deployment.partialDeployment.Module,
				Language:         deployment.partialDeployment.Language,
				AssignedReplicas: assignedReplicas,
				RequiredReplicas: deployment.minReplicas,
			})
		}
	}
	return out, nil
}

func (m *Local) GetIdleRunnersForLanguage(ctx context.Context, language string, limit int) ([]Runner, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	var runners []Runner
	for _, runner := range m.runners {
		if runner.Language == language && runner.State == RunnerStateIdle {
			runners = append(runners, runner.Runner)
			if len(runners) == limit {
				return runners, nil
			}
		}
	}
	return runners, nil
}

func (m *Local) GetRoutingTable(ctx context.Context, module string) ([]string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	var endpoints []string
	for _, runner := range m.runners {
		if runner.State == RunnerStateAssigned {
			if dkey, ok := runner.Deployment.Get(); ok {
				if m.deployments[dkey].partialDeployment.Module == module {
					endpoints = append(endpoints, runner.Endpoint)
				}
			}
		}
	}
	if len(endpoints) == 0 {
		return nil, errors.Wrap(ErrNotFound, "no runners found")
	}
	return endpoints, nil
}

func (m *Local) GetRunnerState(ctx context.Context, runnerKey model.RunnerKey) (RunnerState, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	runner, ok := m.runners[runnerKey]
	if !ok {
		return RunnerStateIdle, errors.Wrap(ErrNotFound, "runner")
	}
	return runner.State, nil
}

func (m *Local) ExpireRunnerClaims(ctx context.Context) (int64, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	var count int64
	now := time.Now()
	for _, runner := range m.runners {
		if runner.State == RunnerStateReserved && runner.reservationTimeout.Before(now) {
			runner.State = RunnerStateIdle
			runner.reservationTimeout = time.Time{}
			count++
		}
	}
	return count, nil
}

func (m *Local) InsertDeploymentLogEntry(ctx context.Context, deployment model.DeploymentKey, logEntry log.Entry) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	panic("implement me")
}

func (m *Local) InsertMetricEntry(ctx context.Context, metric Metric) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	panic("implement me")
}

type lazyFileReader struct {
	path string
	r    *os.File
}

func (l *lazyFileReader) Read(p []byte) (n int, err error) {
	if l.r == nil {
		l.r, err = os.Open(l.path)
		if err != nil {
			return 0, errors.WithStack(err)
		}
	}
	return l.r.Read(p) //nolint:wrapcheck
}

func (l *lazyFileReader) Close() error {
	if l.r != nil {
		return l.r.Close() //nolint:wrapcheck
	}
	return nil
}
