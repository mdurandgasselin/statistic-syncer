package utils

import (
	"github.com/fatih/color"
)

var Red func(...interface {}) string = color.New(color.FgRed).SprintFunc()
var Yellow func(...interface {}) string = color.New(color.FgYellow).SprintFunc()
var Cyan func(...interface {}) string = color.New(color.FgCyan).SprintFunc()
var Blue func(...interface {}) string = color.New(color.FgBlue).SprintFunc()
var Green func(...interface {}) string = color.New(color.FgGreen).SprintFunc()
var Magenta func(...interface {}) string = color.New(color.FgMagenta).SprintFunc()
var White func(...interface {}) string = color.New(color.FgWhite).SprintFunc()
