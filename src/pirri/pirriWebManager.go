package main

import (
	"net/http"
	"runtime"

	"github.com/newrelic/go-agent"
)

func startPirriWebApp() {

	if SETTINGS.NewRelic.Active {
		config := newrelic.NewConfig("PirriGo v"+VERSION, SETTINGS.NewRelic.Key)
		NRAPPMON, err := newrelic.NewApplication(config)

		if NRAPPMON == nil || err != nil {
			getLogger().LogEvent("NewRelic being used.")
		} else {
			for k, v := range protectedRoutes {
				// wrap each route and function in auth handler and new relic
				http.HandleFunc(newrelic.WrapHandleFunc(NRAPPMON, k, basicAuth(v)))

			}

			for k, v := range unprotectedRoutes {
				// wrap each route and function with new relic
				http.HandleFunc(newrelic.WrapHandleFunc(NRAPPMON, k, v))

			}
			// static content does not require authentication
			http.HandleFunc(newrelic.WrapHandleFunc(NRAPPMON, "/static/", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, r.URL.Path[1:])
			}))

			// routes to the login page if not authenticated, to the main /home otherwise
			http.HandleFunc(newrelic.WrapHandleFunc(NRAPPMON, "/", loginAuth))
		}
	} else {
		for k, v := range protectedRoutes {
			getLogger().LogEvent("Not using New Relic for: " + k)
			// wrap each route and function in auth handler
			http.HandleFunc(k, basicAuth(v))
		}
		for k, v := range unprotectedRoutes {
			getLogger().LogEvent("Not using New Relic for: " + k)
			http.HandleFunc(k, v)
		}
		// static content does not require authentication
		http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, r.URL.Path[1:])
		})

		// routes to the login page if not authenticated, to the main /home otherwise
		http.HandleFunc("/login", loginAuth)
	}

	// Host server
	panic(http.ListenAndServe(":"+SETTINGS.Web.Port, nil))
}

func logTraffic() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
