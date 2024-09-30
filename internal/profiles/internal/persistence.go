// Package internal manages the persistent profile configuration of the FTL CLI.
//
// Layout will be something like:
//
//	.ftl-project/
//		project.json
//		profiles/
//			<profile>/
//				profile.json
//				[secrets.json]
//				[config.json]
//
// See the [design document] for more information.
//
// [design document]: https://hackmd.io/@ftl/Sy2GtZKnR
package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/sha256"
)

type ProfileType string

const (
	ProfileTypeLocal  ProfileType = "local"
	ProfileTypeRemote ProfileType = "remote"
)

type Profile struct {
	Name            string                    `json:"name"`
	Endpoint        string                    `json:"endpoint"`
	Type            ProfileType               `json:"type"`
	SecretsProvider configuration.ProviderKey `json:"secrets-provider"`
	ConfigProvider  configuration.ProviderKey `json:"config-provider"`
}

func (p *Profile) EndpointURL() (*url.URL, error) {
	u, err := url.Parse(p.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("profile endpoint: %w", err)
	}
	return u, nil
}

type Project struct {
	Realm          string   `json:"realm"`
	FTLMinVersion  string   `json:"ftl-min-version,omitempty"`
	ModuleRoots    []string `json:"module-roots,omitempty"`
	NoGit          bool     `json:"no-git,omitempty"`
	DefaultProfile string   `json:"default-profile,omitempty"`

	Root string `json:"-"`
}

// ActiveProfile returns the name of the active profile.
//
// If no profile is active, it returns the default.
func (p Project) ActiveProfile() (string, error) {
	cacheDir, err := p.ensureUserProjectDir()
	if err != nil {
		return "", err
	}
	profile, err := os.ReadFile(filepath.Join(cacheDir, "active-profile"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("read active profile: %w", err)
	}
	return strings.TrimSpace(string(profile)), nil
}

func (p Project) SetActiveProfile(profile string) error {
	cacheDir, err := p.ensureUserProjectDir()
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(cacheDir, "active-profile"), []byte(profile), 0600)
	if err != nil {
		return fmt.Errorf("write active profile: %w", err)
	}
	return nil
}

func (p Project) ensureUserProjectDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache dir: %w", err)
	}

	cacheDir = filepath.Join(cacheDir, "ftl-projects", sha256.Sum([]byte(p.Root)).String())
	if err = os.MkdirAll(cacheDir, 0700); err != nil {
		return "", fmt.Errorf("mkdir cache dir: %w", err)
	}
	return cacheDir, nil
}

// ListProfiles returns the names of all profiles in the project.
func (p Project) ListProfiles() ([]Profile, error) {
	profileDir := filepath.Join(p.Root, ".ftl-project", "profiles")
	profiles, err := filepath.Glob(filepath.Join(profileDir, "*", "profile.json"))
	if err != nil {
		return nil, fmt.Errorf("profiles: %s: %w", profileDir, err)
	}
	out := make([]Profile, 0, len(profiles))
	for _, profile := range profiles {
		name := filepath.Base(filepath.Dir(profile))
		profile, err := p.LoadProfile(name)
		if err != nil {
			return nil, fmt.Errorf("%s: load profile: %w", name, err)
		}
		out = append(out, profile)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (p Project) LoadProfile(name string) (Profile, error) {
	profilePath := filepath.Join(p.Root, ".ftl-project", "profiles", name, "profile.json")
	r, err := os.Open(profilePath)
	if err != nil {
		return Profile{}, fmt.Errorf("open %s: %w", profilePath, err)
	}
	defer r.Close() //nolint:errcheck

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	profile := Profile{}
	if err = dec.Decode(&profile); err != nil {
		return Profile{}, fmt.Errorf("decoding %s: %w", profilePath, err)
	}
	return profile, nil
}

// SaveProfile saves a profile to the project.
func (p Project) SaveProfile(profile Profile) error {
	profilePath := filepath.Join(p.Root, ".ftl-project", "profiles", profile.Name, "profile.json")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(profilePath), err)
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return fmt.Errorf("create %s: %w", profilePath, err)
	}
	defer w.Close() //nolint:errcheck

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(profile); err != nil {
		return fmt.Errorf("encoding %s: %w", profilePath, err)
	}
	return nil
}

func (p Project) Save() error {
	profilePath := filepath.Join(p.Root, ".ftl-project", "project.json")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(profilePath), err)
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return fmt.Errorf("create %s: %w", profilePath, err)
	}
	defer w.Close() //nolint:errcheck

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(p); err != nil {
		return fmt.Errorf("encoding %s: %w", profilePath, err)
	}
	return nil
}

// LocalSecretsPath returns the path to the secrets file for the given local profile.
func (p Project) LocalSecretsPath(profile string) string {
	return filepath.Join(p.Root, ".ftl-project", "profiles", profile, "secrets.json")
}

// LocalConfigPath returns the path to the config file for the given local profile.
func (p Project) LocalConfigPath(profile string) string {
	return filepath.Join(p.Root, ".ftl-project", "profiles", profile, "config.json")
}

func Init(project Project) error {
	if project.Root == "" {
		return errors.New("project root is empty")
	}
	if project.DefaultProfile == "" {
		project.DefaultProfile = "local"
	}
	profilePath := filepath.Join(project.Root, ".ftl-project", "project.json")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0700); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(profilePath), err)
	}

	w, err := os.Create(profilePath)
	if err != nil {
		return fmt.Errorf("create %s: %w", profilePath, err)
	}
	defer w.Close() //nolint:errcheck

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(project); err != nil {
		return fmt.Errorf("encoding %s: %w", profilePath, err)
	}

	if err = project.SaveProfile(Profile{
		Name:            project.DefaultProfile,
		Endpoint:        "http://localhost:8892",
		Type:            ProfileTypeLocal,
		SecretsProvider: "inline",
		ConfigProvider:  "inline",
	}); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}

	return nil
}

// Load the project configuration from the given root directory.
func Load(root string) (Project, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Project{}, fmt.Errorf("failed to get absolute path: %w", err)
	}
	profilePath := filepath.Join(root, ".ftl-project", "project.json")
	r, err := os.Open(profilePath)
	if errors.Is(err, os.ErrNotExist) {
		return Project{
			Root: root,
		}, nil
	} else if err != nil {
		return Project{}, fmt.Errorf("open %s: %w", profilePath, err)
	}
	defer r.Close() //nolint:errcheck

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	project := Project{}
	if err = dec.Decode(&project); err != nil {
		return Project{}, fmt.Errorf("decoding %s: %w", profilePath, err)
	}
	project.Root = root
	return project, nil
}
