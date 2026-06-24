package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// HandleProjectExport GET /api/books/export — 将指定项目目录打包为 zip。
func (h *Handlers) HandleProjectExport(ctx context.Context, c *app.RequestContext) {
	path := string(c.Query("path"))
	if strings.TrimSpace(path) == "" {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.pathRequired")
		return
	}
	archive, err := h.app.ExportProjectZip(path)
	if err != nil {
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	disposition := fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(archive.FileName))
	c.Response.Header.Set("Content-Disposition", disposition)
	c.Data(consts.StatusOK, "application/zip", archive.Data)
}

// HandleProjectImport POST /api/books/import-project — 导入 Punkdom 项目 zip。
func (h *Handlers) HandleProjectImport(ctx context.Context, c *app.RequestContext) {
	formFile, err := c.FormFile("file")
	if err != nil {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.invalidRequest")
		return
	}
	file, err := formFile.Open()
	if err != nil {
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	result, err := h.app.ImportProjectZip(ctx, formFile.Filename, file, formFile.Size)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			writeErrorKey(c, consts.StatusConflict, "api.workspace.targetExists")
			return
		}
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, result)
}
