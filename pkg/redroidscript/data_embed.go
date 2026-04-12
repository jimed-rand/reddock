package redroidscript

import (
	"encoding/json"
	"embed"
	"fmt"
)

//go:embed data/*.json
var dataFS embed.FS

var liteGappsLinks map[string]map[string][]string
var mindTheGappsLinks map[string]map[string][]string

func init() {
	raw, err := dataFS.ReadFile("data/litegapps.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(raw, &liteGappsLinks); err != nil {
		panic(err)
	}
	// Upstream typo key "13.0.0__64only" — accept reddock's "13.0.0_64only".
	if v, ok := liteGappsLinks["13.0.0__64only"]; ok {
		liteGappsLinks["13.0.0_64only"] = v
	}

	raw2, err := dataFS.ReadFile("data/mindthegapps.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(raw2, &mindTheGappsLinks); err != nil {
		panic(err)
	}
	if len(liteGappsLinks) == 0 || len(mindTheGappsLinks) == 0 {
		panic(fmt.Errorf("redroidscript: empty link tables"))
	}
}
