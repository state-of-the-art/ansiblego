package template

import (
	"strings"

	"github.com/MarioJim/gonja"
)

// Answers a simple question if the string contains the jinja template
func IsTemplate(input string) bool {
	cfg := gonja.DefaultEnv.Config
	// Check for block "{%  %}"
	start := strings.Index(input, cfg.BlockStartString)
	if start >= 0 && start < strings.LastIndex(input, cfg.BlockEndString) {
		// We have template block inside the string
		return true
	}
	// Check for var "{{  }}"
	start = strings.Index(input, cfg.VariableStartString)
	if start >= 0 && start < strings.LastIndex(input, cfg.VariableEndString) {
		// We have template variable inside the string
		return true
	}
	// Check for comment "{#  #}"
	start = strings.Index(input, cfg.CommentStartString)
	if start >= 0 && start < strings.LastIndex(input, cfg.CommentEndString) {
		// We have template variable inside the string
		return true
	}

	return false
}

// Processing the input string
func Process(input string, context map[string]any) (string, error) {
	tpl, err := gonja.FromString(input)
	if err != nil {
		return "", err
	}
	out, err := tpl.Execute(context)
	if err != nil {
		return "", err
	}
	return out, nil
}
