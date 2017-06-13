package reauth

import (
	"fmt"

	"github.com/freman/caddy-reauth/backend"
	_ "github.com/freman/caddy-reauth/backends"

	"github.com/mholt/caddy"
)

type Rule struct {
	path       string
	exceptions []string
	backends   []backend.Backend
}

func parseConfiguration(c *caddy.Controller) ([]Rule, error) {
	var rules []Rule
	for c.Next() {
		args := c.RemainingArgs()
		switch len(args) {
		case 0:
			r, err := parseBlock(c)
			if err != nil {
				return nil, err
			}
			rules = append(rules, r)
		default:
			// we want only 0 arguments max
			return nil, c.ArgErr()
		}
	}
	return rules, nil
}

func parseBlock(c *caddy.Controller) (Rule, error) {
	r := Rule{backends: []backend.Backend{}}
	for c.NextBlock() {
		switch c.Val() {
		case "path":
			// Path expects just one string argument and only one iteration
			if !c.NextArg() {
				return r, c.ArgErr()
			}
			if len(r.path) != 0 {
				return r, c.ArgErr()
			}
			r.path = c.Val()
			if c.NextArg() {
				return r, c.ArgErr()
			}
		case "except":
			// Except can be specified multiple times with one string argument to
			// provide exceptions
			if !c.NextArg() {
				return r, c.ArgErr()
			}
			r.exceptions = append(r.exceptions, c.Val())
			if c.NextArg() {
				return r, c.ArgErr()
			}
		default:
			// Handle backends which should all have just one argument after the plugin name
			name := c.Val()
			args := c.RemainingArgs()
			if len(args) != 1 {
				return r, fmt.Errorf("wrong number of arguments for %v: %v (%v:%v)", name, args, c.File(), c.Line())
			}

			config := args[0]

			f, err := backend.Lookup(name)
			if err != nil {
				return r, fmt.Errorf("%v for %v (%v:%v)", err, name, c.File(), c.Line())
			}

			b, err := f(config)
			if err != nil {
				return r, fmt.Errorf("%v for %v (%v:%v)", err, name, c.File(), c.Line())
			}

			r.backends = append(r.backends, b)
		}
	}

	if r.path == "" {
		return r, fmt.Errorf("path is a required parameter")
	}

	if len(r.backends) == 0 {
		return r, fmt.Errorf("at least one backend required")
	}

	return r, nil
}
