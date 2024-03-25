package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
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
	Module optional.Option[*Module]
	Symbol Symbol
}

func (s Scope) String() string {
	out := &strings.Builder{}
	for name, decl := range s {
		fmt.Fprintf(out, "%s: %T\n", name, decl.Symbol)
	}
	return out.String()
}

// ResolveTypeAs resolves a [Type] to a concrete symbol and declaration.
func ResolveTypeAs[S Symbol](scopes Scopes, t Type) (symbol S, decl *ModuleDecl) {
	decl = scopes.ResolveType(t)
	if decl == nil {
		return
	}
	var ok bool
	symbol, ok = decl.Symbol.(S)
	if !ok {
		return
	}
	return symbol, decl
}

// ResolveAs resolves a [Ref] to a concrete symbol and declaration.
func ResolveAs[S Symbol](scopes Scopes, ref Ref) (symbol S, decl *ModuleDecl) {
	decl = scopes.Resolve(ref)
	if decl == nil {
		return
	}
	var ok bool
	symbol, ok = decl.Symbol.(S)
	if !ok {
		return
	}
	return symbol, decl
}

// Scopes to search during type resolution.
type Scopes []Scope

// NewScopes constructs a new type resolution stack with builtins pre-populated.
func NewScopes() Scopes {
	builtins := Builtins()
	// Empty scope tail for builtins.
	scopes := Scopes{primitivesScope, Scope{}}
	if err := scopes.Add(optional.None[*Module](), builtins.Name, builtins); err != nil {
		panic(err)
	}
	for _, decl := range builtins.Decls {
		if err := scopes.Add(optional.Some(builtins), decl.GetName(), decl); err != nil {
			panic(err)
		}
	}
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
			fmt.Fprintf(out, "  %s: %T\n", name, decl.Symbol)
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
func (s *Scopes) Add(owner optional.Option[*Module], name string, symbol Symbol) error {
	end := len(*s) - 1
	if prev, ok := (*s)[end][name]; ok {
		return fmt.Errorf("%s: duplicate declaration of %q at %s", symbol.Position(), name, prev.Symbol.Position())
	}
	(*s)[end][name] = ModuleDecl{owner, symbol}
	return nil
}

func (s Scopes) ResolveType(t Type) *ModuleDecl {
	switch t := t.(type) {
	case Named:
		return s.Resolve(Ref{Name: t.GetName()})

	case *Ref:
		return s.Resolve(*t)

	case Symbol:
		return &ModuleDecl{optional.None[*Module](), t}

	default:
		return nil
	}
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
			if resolver, ok := mdecl.Symbol.(Resolver); ok {
				if decl := resolver.Resolve(ref); decl != nil {
					// Holy nested if statement Batman.
					return decl
				}
			} else {
				if module, ok := mdecl.Module.Get(); ok {
					for _, d := range module.Decls {
						if d.GetName() == ref.Name {
							return &ModuleDecl{mdecl.Module, d}
						}
					}
				}
			}
		}
	}
	return nil
}
