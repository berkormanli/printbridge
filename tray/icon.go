//go:build windows

package tray

import (
    _ "embed"
)

//go:embed "iconwin.ico"
var Icon []byte