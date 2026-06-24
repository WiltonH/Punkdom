package handlers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// HandleDataBackupExport GET /api/backup/export — 导出完整 Punkdom 数据目录备份。
func (h *Handlers) HandleDataBackupExport(ctx context.Context, c *app.RequestContext) {
	archive, err := h.app.ExportDataBackupZip()
	if err != nil {
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	disposition := fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(archive.FileName))
	c.Response.Header.Set("Content-Disposition", disposition)
	c.Data(consts.StatusOK, "application/zip", archive.Data)
}

// HandleDataBackupRestore POST /api/backup/restore — 上传备份 zip 并覆盖还原完整 Punkdom 数据目录。
func (h *Handlers) HandleDataBackupRestore(ctx context.Context, c *app.RequestContext) {
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

	result, err := h.app.RestoreDataBackupZip(ctx, formFile.Filename, file, formFile.Size)
	if err != nil {
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, result)
}
