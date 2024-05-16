package named

import (
	"context"
	"ftl/namedext"
)

// named module is for testing names types: typealiases and enums

// ID testing if typealias before struct works
//
//ftl:typealias export
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

//ftl:data export
type User struct {
	Id           ID
	Name         Name
	State        UserState
	Source       UserSource
	Comment      namedext.Comment
	EmailConsent namedext.EmailConsent
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
