package handlers

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// handleBooks GET /api/books — 返回当前 Punkdom 数据目录下实际存在的书籍工作目录。
func (h *Handlers) HandleBooks(ctx context.Context, c *app.RequestContext) {
	writeJSON(c, consts.StatusOK, map[string]interface{}{
		"books": h.app.Books(),
	})
}

// HandleDeletedBooks GET /api/books/deleted — 返回 Trash 中可恢复的项目。
func (h *Handlers) HandleDeletedBooks(ctx context.Context, c *app.RequestContext) {
	writeJSON(c, consts.StatusOK, map[string]interface{}{
		"books": h.app.DeletedBooks(),
	})
}

// handleCreateBook POST /api/books/create — 创建新书籍工作区。
func (h *Handlers) HandleCreateBook(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Title       string `json:"title"`
		Author      string `json:"author,omitempty"`
		Description string `json:"description,omitempty"`
	}
	if err := c.BindJSON(&req); err != nil {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.invalidRequest")
		return
	}
	if req.Title == "" {
		writeErrorKey(c, consts.StatusBadRequest, "api.books.titleRequired")
		return
	}
	layered, err := h.app.Settings()
	if err != nil {
		writeError(c, consts.StatusInternalServerError, err.Error())
		return
	}
	if layered.Paths.PunkdomDir == "" {
		writeErrorKey(c, consts.StatusInternalServerError, "api.books.punkdomDirMissing")
		return
	}
	workspace, meta, err := h.app.CreateBook(ctx, layered.Paths.PunkdomDir, req.Title, req.Author, req.Description)
	if err != nil {
		status := consts.StatusInternalServerError
		if strings.Contains(err.Error(), "已存在") {
			status = consts.StatusConflict
		}
		writeError(c, status, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, map[string]interface{}{
		"workspace": workspace,
		"book_meta": meta,
	})
}

// handleBookRemove POST /api/books/remove — 将项目移动到 Trash。
func (h *Handlers) HandleBookRemove(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.BindJSON(&req); err != nil || req.Path == "" {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.pathRequired")
		return
	}
	workspace, err := h.app.RemoveBook(req.Path)
	if err != nil {
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, map[string]string{
		"message":   messageKey(c, "api.books.removed"),
		"workspace": workspace,
	})
}

// HandleBookRestore POST /api/books/restore — 从 Trash 恢复项目。
func (h *Handlers) HandleBookRestore(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.BindJSON(&req); err != nil || req.Path == "" {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.pathRequired")
		return
	}
	workspace, err := h.app.RestoreBook(ctx, req.Path)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			writeErrorKey(c, consts.StatusConflict, "api.workspace.targetExists")
			return
		}
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, map[string]string{
		"message":   messageKey(c, "api.books.restored"),
		"workspace": workspace,
	})
}

// HandleBookPurge POST /api/books/purge — 彻底删除 Trash 中的项目文件。
func (h *Handlers) HandleBookPurge(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.BindJSON(&req); err != nil || req.Path == "" {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.pathRequired")
		return
	}
	if err := h.app.PurgeDeletedBook(req.Path); err != nil {
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, map[string]string{"message": messageKey(c, "api.books.purged")})
}

// handleBookReorder POST /api/books/reorder — 保存书籍管理页自定义排序。
func (h *Handlers) HandleBookReorder(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Paths []string `json:"paths"`
	}
	if err := c.BindJSON(&req); err != nil {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.invalidRequest")
		return
	}
	if err := h.app.ReorderBooks(req.Paths); err != nil {
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, map[string]string{"message": messageKey(c, "api.books.reordered")})
}

// handleBookInfo GET /api/books/info — 读取指定工作区的书籍元信息。
func (h *Handlers) HandleBookInfo(ctx context.Context, c *app.RequestContext) {
	path := string(c.Query("path"))
	if path == "" {
		writeErrorKey(c, consts.StatusBadRequest, "api.books.pathQueryRequired")
		return
	}
	meta, err := h.app.BookInfo(path)
	if err != nil {
		writeError(c, consts.StatusBadRequest, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, meta)
}

// handleUpdateBookInfo PUT /api/books/info — 更新指定工作区的书籍元信息。
func (h *Handlers) HandleUpdateBookInfo(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Path        string `json:"path"`
		Title       string `json:"title"`
		Author      string `json:"author"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&req); err != nil {
		writeErrorKey(c, consts.StatusBadRequest, "api.common.invalidRequest")
		return
	}
	if req.Path == "" {
		writeErrorKey(c, consts.StatusBadRequest, "api.books.pathRequired")
		return
	}
	result, err := h.app.UpdateBookInfo(req.Path, req.Title, req.Author, req.Description)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			writeErrorKey(c, consts.StatusConflict, "api.workspace.targetExists")
			return
		}
		if strings.Contains(err.Error(), "新名称") || strings.Contains(err.Error(), "隐藏文件名") || strings.Contains(err.Error(), "路径不是目录") {
			writeError(c, consts.StatusBadRequest, err.Error())
			return
		}
		writeError(c, consts.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(c, consts.StatusOK, result)
}
