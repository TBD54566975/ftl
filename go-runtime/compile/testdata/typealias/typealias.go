package typealias

import "context"

// ID testing if typealias before struct works
//
//ftl:typealias
type ID string

// UserState, testing that defining an enum before struct works
//
//ftl:enum
type UserState string

const (
	Registered UserState = "registered"
	Active     UserState = "active"
	Inactive   UserState = "inactive"
)

//ftl:data
type User struct {
	Id     ID
	Name   Name
	State  UserState
	Source UserSource
}

// Name testing if typealias after struct works
//
//ftl:typealias
type Name string

// UserSource, testing that defining an enum after struct works
//
//ftl:enum
type UserSource string

const (
	Magazine UserSource = "magazine"
	Friend   UserSource = "friend"
	Ad       UserSource = "ad"
)

//ftl:typealias
type InternalUser User

//ftl:typealias
type DoubleAliasedUser InternalUser

//ftl:verb
func PingUser(ctx context.Context, req User) error {
	return nil
}

//ftl:verb
func PingInternalUser(ctx context.Context, req InternalUser) error {
	return nil
}

//TODO: type alias which is used as an enum. eg seasons

//TODO: type alias of type alias

//TODO: make sure typealias logic isnt breaking verb names (different namespaces)

//TODO: refer to an external type alias, what happens?

//TODO: something where we do not export the type alias but then it is used in an exported way, does that follow through to the original type?
