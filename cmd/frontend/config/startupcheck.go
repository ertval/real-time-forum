// cmd/frontend/config/startupcheck.go
package config

import "forum/web/startupcheck"

func StartupCheckConfig() startupcheck.Config {
	return startupcheck.Config{
		WebRoot: ".",

		CriticalHTML: []string{
			"SPA/index.html",
			"web/errors/error.html",
		},

		JSDirs: []string{
			"SPA",
		},

		CriticalCSS: []string{
			"SPA/assets/css/main.css",
		},

		MinJSFiles:   1, // main.js
		ExactJSFiles: 0,
	}
}

// HELPER
func ValidateFrontendStartup() error {
	return startupcheck.ValidateFiles(StartupCheckConfig())
}
