//ftl:module ad
package ad

import (
	"context"
	_ "embed"
	"math/rand"

	"ftl/builtin"

	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/examples/online-boutique/backend/common"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

const maxAdsToServe = 2

var (
	//go:embed database.json
	databaseJSON []byte
	database     = common.LoadDatabase[map[string]Ad](databaseJSON)
)

type AdRequest struct {
	ContextKeys []string
}

type Ad struct {
	RedirectURL string
	Text        string
}

type AdResponse struct {
	Name string
	Ads  []Ad
}

//ftl:ingress GET /ad
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, AdRequest]) (builtin.HttpResponse[AdResponse, ftl.Unit], error) {
	var ads []Ad
	if len(req.Query.ContextKeys) > 0 {
		ads = contextualAds(req.Query.ContextKeys)
	} else {
		ads = randomAds()
	}

	return builtin.HttpResponse[AdResponse, ftl.Unit]{
		Body: ftl.Some(AdResponse{Name: "ad", Ads: ads}),
	}, nil
}

func contextualAds(contextKeys []string) (ads []Ad) {
	for _, key := range contextKeys {
		if ad, ok := database[key]; ok {
			ads = append(ads, ad)
		}
	}
	return ads
}

func randomAds() (ads []Ad) {
	allAds := maps.Values(database)
	for i := 0; i < maxAdsToServe; i++ {
		ads = append(ads, allAds[rand.Intn(len(allAds))])
	}
	return ads
}
