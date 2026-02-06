package config

import "forum/web/startupcheck"

func StartupCheckConfig() startupcheck.Config {
	return startupcheck.Config{
		WebRoot: "web",

		CriticalHTML: []string{
			"templates/home.html",
			"templates/login.html",
			"templates/register.html",
			"templates/create-post.html",
			"templates/my-posts.html",
			"templates/my-liked-posts.html",
			"templates/view-post.html",
			"static/partials/header.html",
			"errors/error.html",
		},

		JSDirs: []string{
			"static/js",
		},

		MinJSFiles:   0,  // disabled
		ExactJSFiles: 21, // enforced
	}
}


//HELPER
func ValidateFrontendStartup() error {
	return startupcheck.ValidateFiles(StartupCheckConfig())
}