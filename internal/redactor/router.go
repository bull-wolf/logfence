package redactor

import (
	"regexp"
	"strings"
)

// RouteRule maps a log source pattern to a named pipeline configuration.
type RouteRule struct {
	// SourcePattern is a regex matched against the log source label (e.g. container name).
	SourcePattern string `yaml:"source_pattern"`
	// Pipeline is the name of the pipeline config to apply.
	Pipeline string `yaml:"pipeline"`
}

// RouterConfig holds the routing table and a fallback pipeline name.
type RouterConfig struct {
	Routes          []RouteRule `yaml:"routes"`
	DefaultPipeline string      `yaml:"default_pipeline"`
}

// DefaultRouterConfig returns a RouterConfig that routes everything to the
// built-in default pipeline.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		DefaultPipeline: "default",
	}
}

// compiledRoute is a RouteRule with its pattern pre-compiled.
type compiledRoute struct {
	pattern  *regexp.Regexp
	pipeline string
}

// Router selects a pipeline name for a given log source.
type Router struct {
	routes          []compiledRoute
	defaultPipeline string
}

// NewRouter compiles all route patterns and returns a Router.
// Returns an error if any pattern fails to compile.
func NewRouter(cfg RouterConfig) (*Router, error) {
	routes := make([]compiledRoute, 0, len(cfg.Routes))
	for _, r := range cfg.Routes {
		re, err := regexp.Compile(r.SourcePattern)
		if err != nil {
			return nil, err
		}
		routes = append(routes, compiledRoute{pattern: re, pipeline: r.Pipeline})
	}
	dp := strings.TrimSpace(cfg.DefaultPipeline)
	if dp == "" {
		dp = "default"
	}
	return &Router{routes: routes, defaultPipeline: dp}, nil
}

// Route returns the pipeline name for the given source string.
// Routes are evaluated in order; the first match wins.
// If no route matches, the default pipeline name is returned.
func (r *Router) Route(source string) string {
	for _, cr := range r.routes {
		if cr.pattern.MatchString(source) {
			return cr.pipeline
		}
	}
	return r.defaultPipeline
}
