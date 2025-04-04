// Package init provides automatic initialization for the planks-go/slog package.
// Importing this package will automatically configure and set the default slog logger
// based on environment variables.
package init

import (
	"github.com/nakat-t/planks-go/slog"
)

func init() {
	slog.Init()
}
