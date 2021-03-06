// Copyright Jeevanandam M. (https://github.com/jeevatkm, jeeva@myjeeva.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"aahframe.work"
	"aahframe.work/essentials"
)

// CheckConfig method subscribes to aah `OnInit` event to check config
// and puts default values as needed.
//
// Reads thumbai config values and sets appropriate on aah config instance.
func CheckConfig(e *aah.Event) {
	app := aah.App()
	cfg := app.Config()
	appProfile := cfg.StringDefault("thumbai.env.active", "prod")
	cfg.SetString("env.active", appProfile)

	checkRequiredValues([]string{
		"thumbai.admin.host",
		"thumbai.server",
		"thumbai.log",
		"thumbai.security.session.sign_key",
		"thumbai.security.session.enc_key",
		"thumbai.security.anti_csrf.sign_key",
		"thumbai.security.anti_csrf.enc_key",
	})

	if !cfg.IsExists("thumbai.admin.host") {
		app.Log().Fatalf("'thumbai.admin.host' value is not configured")
	}

	adminHost := cfg.StringDefault("thumbai.admin.host", "")
	if i := strings.IndexByte(adminHost, ':'); i > 0 {
		cfg.SetString("env."+appProfile+".routes.domains.thumbai.port", adminHost[i+1:])
		adminHost = adminHost[:i]
	}
	cfg.SetString("env."+appProfile+".routes.domains.thumbai.host", adminHost)

	if !cfg.IsExists("thumbai.admin.data_store.directory") {
		cfg.SetString("thumbai.admin.data_store.directory", filepath.Join(app.BaseDir(), "data"))
	}

	if ess.IsStrEmpty(cfg.StringDefault("thumbai.admin.contact_email", "")) {
		app.Log().Warn("'thumbai.admin.contact_email' value is not yet configured. Highly recommended to configure it.")
	}

	cfg.SetBool("server.access_log.enable", false)

	readSectionAndSet("thumbai.server", "env."+appProfile+".server")
	readSectionAndSet("thumbai.log", "env."+appProfile+".log")

	readAndSet("thumbai.security.session.sign_key")
	readAndSet("thumbai.security.session.enc_key")
	readAndSet("thumbai.security.anti_csrf.sign_key")
	readAndSet("thumbai.security.anti_csrf.enc_key")
}

func readSectionAndSet(srcSecKey, dstSecKey string) {
	if tocfg, found := aah.App().Config().GetSubConfig(srcSecKey); found {
		if err := aah.App().Config().Merge2Section(dstSecKey, tocfg); err != nil {
			aah.App().Log().Error(err)
		}
	}
}

func readAndSet(cfgKey string) {
	cfgValue := aah.App().Config().StringDefault(cfgKey, "")
	if len(cfgValue) == 0 {
		aah.App().Log().Fatalf("'%s' config value is not provided", cfgKey)
	}
	aah.App().Config().SetString(strings.TrimPrefix(cfgKey, "thumbai."), cfgValue)
}

func checkRequiredValues(cfgKeys []string) {
	var errs []string
	app := aah.App()
	cfg := app.Config()
	for _, cfgKey := range cfgKeys {
		if !cfg.IsExists(cfgKey) {
			errs = append(errs, fmt.Sprintf("'%s' value is missing", cfgKey))
		}
	}
	if len(errs) > 0 {
		app.Log().Fatalf("Required configuration vaules are missing: \n\t%s", strings.Join(errs, "\n\t"))
	}
}
