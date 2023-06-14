// Package inventory provides a simple inventory system.
//
//ftl:module inventory
package inventory

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/go-runtime/sdk/kvstore"
)

var inventory = kvstore.Require[Item]()

type ItemRef string

type Item struct {
	ID          ItemRef
	Description string
	Count       int
}

type CreateRequest struct {
	ID          ItemRef
	Description string
}

//ftl:verb
func Create(ctx context.Context, req CreateRequest) (Item, error) {
	return inventory.Upsert(string(req.ID), func(item Item, created bool) (Item, error) {
		if !created {
			return Item{}, errors.Errorf("item %q already exists", req.ID)
		}
		return Item{ID: req.ID, Description: req.Description}, nil

	})
}

type AddRequest struct {
	ID    ItemRef
	Count int
}

//ftl:verb
func Add(ctx context.Context, req AddRequest) (Item, error) {
	return inventory.Upsert(string(req.ID), func(i Item, created bool) (Item, error) {
		if created {
			return Item{}, errors.Errorf("no such item")
		}
		i.Count += req.Count
		return i, nil
	})
}

type TakeRequest AddRequest

//ftl:verb
func Take(ctx context.Context, req TakeRequest) (Item, error) {
	return inventory.Upsert(string(req.ID), func(i Item, created bool) (Item, error) {
		if created {
			return Item{}, errors.Errorf("no such item")
		}
		i.Count -= req.Count
		return i, nil
	})
}
