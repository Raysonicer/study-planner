// Package frontend 内嵌前端构建产物（web/dist）。
package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Sub 返回内嵌前端静态文件的 fs.FS（根即 dist 目录内容）。
func Sub() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
