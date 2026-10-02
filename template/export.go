// Package template is the ZATRANO addon for HTML templates.
//
// It owns framework wiring (boot, HTTP bridge, CLI, container key "template").
// The render engine is pluggable via Engine / SetFactory; the default is Canvas
// (github.com/zatrano/canvas), independent like rawhttp.
package template
