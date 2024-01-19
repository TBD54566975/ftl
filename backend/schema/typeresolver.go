package schema

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Resolver may be implemented be a node in the AST to resolve references within itself.
type Resolver interface {
	// Resolve a reference to a symbol declaration or nil.
	Resolve(ref Ref) *ModuleDecl
}

// Scoped is implemented by nodes that wish to introduce a new scope.
type Scoped interface {
	Scope() Scope
}

// Scope maps relative names to fully qualified types.
type Scope map[string]ModuleDecl

// ModuleDecl is a declaration associated with a module.
type ModuleDecl struct {
	Module *Module // May be nil.
	Decl   Decl
}

func (s Scope) String() string {
	out := &strings.Builder{}
	for name, decl := range s {
		fmt.Fprintf(out, "%s: %T\n", name, decl.Decl)
	}
	return out.String()
}

// Scopes to search during type resolution.
type Scopes []Scope

// NewScopes constructs a new type resolution stack with builtins pre-populated.
func NewScopes() Scopes {
	builtins := Builtins()
	// Empty scope tail for builtins.
	scopes := Scopes{primitivesScope, Scope{}}
	if err := scopes.Add(nil, builtins.Name, builtins); err != nil {
		panic(err)
	}
	scopes = scopes.PushScope(builtins.Scope())
	// Push an empty scope for other modules to be added to.
	scopes = scopes.Push()
	return scopes
}

var _ Resolver = Scopes{}

func (s Scopes) String() string {
	out := &strings.Builder{}
	for i, scope := range s {
		if i != 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "Scope %d:\n", i)
		for name, decl := range scope {
			fmt.Fprintf(out, "  %s: %T\n", name, decl.Decl)
		}
	}
	return out.String()
}

func (s Scopes) PushScope(scope Scope) Scopes {
	out := make(Scopes, 0, len(s)+1)
	out = append(out, s...)
	out = append(out, scope)
	return out

}

// Push a new Scope onto the stack.
//
// This contains references to previous Scopes so that updates are preserved.
func (s Scopes) Push() Scopes {
	return s.PushScope(Scope{})
}

// Add a declaration to the current scope.
func (s *Scopes) Add(owner *Module, name string, decl Decl) error {
	end := len(*s) - 1
	if name == "destroy" {
		debug.PrintStack()
	}
	if prev, ok := (*s)[end][name]; ok {
		return fmt.Errorf("%s: duplicate declaration of %q at %s", decl.Position(), name, prev.Decl.Position())
	}
	(*s)[end][name] = ModuleDecl{owner, decl}
	return nil
}

// Resolve a reference to a symbol declaration or nil.
func (s Scopes) Resolve(ref Ref) *ModuleDecl {
	if ref.Module == "" {
		for i := len(s) - 1; i >= 0; i-- {
			scope := s[i]
			if decl, ok := scope[ref.Name]; ok {
				return &decl
			}
		}
		return nil
	}
	// If a module is provided, try to resolve it, then resolve the reference through the module.
	for i := len(s) - 1; i >= 0; i-- {
		scope := s[i]
		if mdecl, ok := scope[ref.Module]; ok {
			if resolver, ok := mdecl.Decl.(Resolver); ok {
				if decl := resolver.Resolve(ref); decl != nil {
					// Holy nested if statement Batman.
					return decl
				}
			}
		}
	}
	return nil
}
