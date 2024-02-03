//ftl:module ad
package ad

import (
	"context"
	_ "embed"
	"math/rand"

	"ftl/builtin"

	"github.com/TBD54566975/ftl/examples/online-boutique/common"
	"golang.org/x/exp/maps"
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

//ftl:verb
//ftl:ingress GET /ad
func Get(ctx context.Context, req builtin.HttpRequest[AdRequest]) (builtin.HttpResponse[AdResponse], error) {
	var ads []Ad
	if len(req.Body.ContextKeys) > 0 {
		ads = contextualAds(req.Body.ContextKeys)
	} else {
		ads = randomAds()
	}

	return builtin.HttpResponse[AdResponse]{
		Body: AdResponse{Name: "ad", Ads: ads},
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
