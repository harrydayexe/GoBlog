// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import "github.com/harrydayexe/GoWebUtilities/middleware"

// ServerConfig holds the resolved configuration of the HTTP server.
//
// Each field is populated by applying [ServerOption] values in
// [github.com/harrydayexe/GoBlog/v2/pkg/server.New]; the struct is embedded in
// the server so the resolved values are readable through it (for example
// srv.Port or srv.Logger.Logger).
//
// GeneratorOpts and RendererOpts collect the options forwarded to the internal
// generator and template renderer. They are populated by passing
// [GeneratorOption.AsServerOption] and [RendererOption.AsServerOption] values
// to the server constructor, and stay in option form because the server does
// not resolve them itself.
//
// Values should not be modified after the server has been constructed.
type ServerConfig struct {
	BlogRoot        BlogRoot
	Port            Port
	Host            Host
	Logger          Logger
	CacheControlTTL CacheControlTTL
	HealthChecks    HealthChecks
	AssetsDir       AssetsDir
	TemplateDir     TemplateDir
	Middleware      []middleware.Middleware
	GeneratorOpts   []GeneratorOption
	RendererOpts    []RendererOption
}
