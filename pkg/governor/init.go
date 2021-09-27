package governor

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"github.com/xqk/ox/pkg/util/ostring"

	jsoniter "github.com/json-iterator/go"
	"github.com/xqk/ox/pkg"
	"github.com/xqk/ox/pkg/conf"
)

func init() {
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Print("hook config, init runtime(governor)")

	})

	registerHandlers()
}

func registerHandlers() {
	HandleFunc("/configs", func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		if r.URL.Query().Get("pretty") == "true" {
			encoder.SetIndent("", "    ")
		}
		encoder.Encode(conf.Traverse("."))
	})

	HandleFunc("/debug/config", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(ostring.PrettyJSONBytes(conf.Traverse(".")))
	})

	HandleFunc("/debug/env", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = jsoniter.NewEncoder(w).Encode(os.Environ())
	})

	HandleFunc("/build/info", func(w http.ResponseWriter, r *http.Request) {
		serverStats := map[string]string{
			"name":       pkg.Name(),
			"appID":      pkg.AppID(),
			"appMode":    pkg.AppMode(),
			"appVersion": pkg.AppVersion(),
			"oxVersion":  pkg.OxVersion(),
			"buildUser":  pkg.BuildUser(),
			"buildHost":  pkg.BuildHost(),
			"buildTime":  pkg.BuildTime(),
			"startTime":  pkg.StartTime(),
			"hostName":   pkg.HostName(),
			"goVersion":  pkg.GoVersion(),
		}
		_ = jsoniter.NewEncoder(w).Encode(serverStats)
	})
}
