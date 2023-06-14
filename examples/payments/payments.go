//ftl:module payments
package payments

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/go-runtime/sdk/kvstore"
)

var payments = kvstore.Require[Account]()

type Account struct {
	ID      string
	Balance int
}

//ftl:verb
func Create(ctx context.Context, req Account) (Account, error) {
	return payments.Upsert(req.ID, func(v Account, created bool) (Account, error) {
		if created {
			return req, nil
		}
		return Account{}, errors.New("account already exists")
	})
}
