package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"punkdom/config"
	"punkdom/internal/book"
)

func TestBookRegistryTouchListAndCurrent(t *testing.T) {
	root := t.TempDir()
	bookA := filepath.Join(root, "book-a")
	bookB := filepath.Join(root, "book-b")
	if err := os.MkdirAll(bookA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(bookB, 0o755); err != nil {
		t.Fatal(err)
	}

	registry := &BookRegistry{path: filepath.Join(root, "books.json")}
	if err := registry.Touch(bookA); err != nil {
		t.Fatalf("记录 bookA 失败: %v", err)
	}
	if err := registry.Touch(bookB); err != nil {
		t.Fatalf("记录 bookB 失败: %v", err)
	}

	if got := registry.Current(); got != bookB {
		t.Fatalf("当前书籍不符合预期: want=%s got=%s", bookB, got)
	}
	books := registry.List()
	if len(books) != 2 {
		t.Fatalf("书籍记录数量不符合预期: %d", len(books))
	}
	if books[0].Path != bookB || books[1].Path != bookA {
		t.Fatalf("书籍记录排序不符合预期: %#v", books)
	}
}

func TestBookRegistryListScansPunkdomDirBooks(t *testing.T) {
	root := t.TempDir()
	bookA := filepath.Join(root, "zeta")
	bookB := filepath.Join(root, "alpha")
	missingBook := filepath.Join(root, "missing")
	for _, dir := range []string{
		filepath.Join(bookA, ".punkdom"),
		filepath.Join(bookB, "chapters"),
		filepath.Join(root, "book_meta"),
		filepath.Join(root, "styles"),
		filepath.Join(root, "notes"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	registry := &BookRegistry{path: filepath.Join(root, "books.json"), punkdomDir: root}
	if err := registry.save(bookRegistryData{
		Books: []BookRecord{
			{Path: missingBook, LastOpenedAt: "2026-01-03T00:00:00Z"},
			{Path: bookA, LastOpenedAt: "2026-01-02T00:00:00Z"},
		},
	}); err != nil {
		t.Fatalf("写入注册表失败: %v", err)
	}

	books := registry.List()
	if len(books) != 2 {
		t.Fatalf("书籍数量不符合预期: %#v", books)
	}
	if books[0].Path != bookB || books[1].Path != bookA {
		t.Fatalf("书籍应来自 Punkdom 目录并按名称排序: %#v", books)
	}
	if books[1].LastOpenedAt != "2026-01-02T00:00:00Z" {
		t.Fatalf("应保留已有打开时间用于兼容展示: %#v", books[1])
	}
}

func TestBookRegistryMoveToTrash(t *testing.T) {
	root := t.TempDir()
	bookA := filepath.Join(root, "book-a")
	bookB := filepath.Join(root, "book-b")
	for _, dir := range []string{
		filepath.Join(bookA, ".punkdom"),
		filepath.Join(bookB, ".punkdom"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	registry := &BookRegistry{path: filepath.Join(root, "books.json"), punkdomDir: root}
	if err := registry.Touch(bookA); err != nil {
		t.Fatal(err)
	}
	if err := registry.Touch(bookB); err != nil {
		t.Fatal(err)
	}
	deleted, err := registry.MoveToTrash(bookB)
	if err != nil {
		t.Fatalf("移入 Trash 失败: %v", err)
	}

	if got := registry.Current(); got != bookA {
		t.Fatalf("删除当前项目后应回退到上一条记录: want=%s got=%s", bookA, got)
	}
	books := registry.List()
	if len(books) != 1 || books[0].Path != bookA {
		t.Fatalf("移入 Trash 后项目列表不符合预期: %#v", books)
	}
	if _, err := os.Stat(bookB); !os.IsNotExist(err) {
		t.Fatalf("原项目目录应被移动: %v", err)
	}
	if _, err := os.Stat(deleted.Path); err != nil {
		t.Fatalf("Trash 中应存在项目目录: %v", err)
	}
	deletedList := registry.Deleted()
	if len(deletedList) != 1 || deletedList[0].OriginalPath != bookB {
		t.Fatalf("已删除列表不符合预期: %#v", deletedList)
	}
}

func TestDeletedBooksPreserveProjectMetadata(t *testing.T) {
	root := t.TempDir()
	punkdomDir := filepath.Join(root, "punkdom")
	application, err := New(context.Background(), &config.Config{
		PunkdomDir:          punkdomDir,
		Workspace:           filepath.Join(root, "initial"),
		ResumeLastWorkspace: false,
		OpenAIModel:         "test-model",
	})
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}
	workspace, meta, err := application.CreateBook(context.Background(), punkdomDir, "虚数时间", "Wilton", "硬科幻项目简介")
	if err != nil {
		t.Fatalf("创建项目失败: %v", err)
	}
	if meta.CreatedAt == "" {
		t.Fatalf("创建项目应写入创建时间: %#v", meta)
	}
	readMeta, err := application.BookInfo(workspace)
	if err != nil {
		t.Fatalf("删除前读取项目元信息失败: %v", err)
	}
	if readMeta.CreatedAt != meta.CreatedAt {
		t.Fatalf("删除前项目元信息应可从存储读取: want=%s got=%s meta=%#v", meta.CreatedAt, readMeta.CreatedAt, readMeta)
	}

	if _, err := application.RemoveBook(workspace); err != nil {
		t.Fatalf("删除项目失败: %v", err)
	}
	deleted := application.DeletedBooks()
	if len(deleted) != 1 {
		t.Fatalf("已删除列表数量不符合预期: %#v", deleted)
	}
	if deleted[0].CreatedAt != meta.CreatedAt {
		t.Fatalf("已删除项目应保留原创建时间: want=%s got=%s record=%#v", meta.CreatedAt, deleted[0].CreatedAt, deleted[0])
	}
	if deleted[0].Description != "硬科幻项目简介" {
		t.Fatalf("已删除项目应保留简介: %#v", deleted[0])
	}
}

func TestBookRegistryRestoreAndPurgeDeletedBook(t *testing.T) {
	root := t.TempDir()
	bookA := filepath.Join(root, "book-a")
	bookB := filepath.Join(root, "book-b")
	for _, dir := range []string{
		filepath.Join(bookA, ".punkdom"),
		filepath.Join(bookB, ".punkdom"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	registry := &BookRegistry{path: filepath.Join(root, "books.json"), punkdomDir: root}
	deleted, err := registry.MoveToTrash(bookB)
	if err != nil {
		t.Fatalf("移入 Trash 失败: %v", err)
	}

	books := registry.List()
	if len(books) != 1 || books[0].Path != bookA {
		t.Fatalf("移入 Trash 后扫描列表应隐藏目标项目: %#v", books)
	}
	restoredPath, err := registry.Restore(deleted.Path)
	if err != nil {
		t.Fatalf("恢复项目失败: %v", err)
	}
	if restoredPath != bookB {
		t.Fatalf("恢复路径不符合预期: want=%s got=%s", bookB, restoredPath)
	}
	if len(registry.Deleted()) != 0 {
		t.Fatalf("恢复后已删除列表应为空: %#v", registry.Deleted())
	}
	deleted, err = registry.MoveToTrash(bookB)
	if err != nil {
		t.Fatalf("再次移入 Trash 失败: %v", err)
	}
	if err := registry.Purge(deleted.Path); err != nil {
		t.Fatalf("彻底删除失败: %v", err)
	}
	if _, err := os.Stat(deleted.Path); !os.IsNotExist(err) {
		t.Fatalf("彻底删除后 Trash 目录应不存在: %v", err)
	}
}

func TestBookRegistryReorderScannedPunkdomBooks(t *testing.T) {
	root := t.TempDir()
	bookA := filepath.Join(root, "alpha")
	bookB := filepath.Join(root, "zeta")
	for _, dir := range []string{
		filepath.Join(bookA, ".punkdom"),
		filepath.Join(bookB, ".punkdom"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	registry := &BookRegistry{path: filepath.Join(root, "books.json"), punkdomDir: root}
	if err := registry.Reorder([]string{bookB, bookA}); err != nil {
		t.Fatalf("保存排序失败: %v", err)
	}

	books := registry.List()
	if len(books) != 2 || books[0].Path != bookB || books[1].Path != bookA {
		t.Fatalf("书籍列表应遵循自定义排序: %#v", books)
	}

	if err := registry.Touch(bookA); err != nil {
		t.Fatalf("打开书籍失败: %v", err)
	}
	books = registry.List()
	if len(books) != 2 || books[0].Path != bookB || books[1].Path != bookA {
		t.Fatalf("打开书籍不应打乱自定义排序: %#v", books)
	}
}

func TestNewBookRegistryUsesPunkdomDir(t *testing.T) {
	punkdomDir := t.TempDir()
	registry := NewBookRegistry(punkdomDir)
	want := filepath.Join(punkdomDir, "books.json")
	if registry.path != want {
		t.Fatalf("注册表路径不符合预期: want=%s got=%s", want, registry.path)
	}
}

func TestBookRegistryLoadsLegacyPathAndMigratesOnSave(t *testing.T) {
	root := t.TempDir()
	bookDir := filepath.Join(root, "book")
	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		t.Fatal(err)
	}

	legacyPath := filepath.Join(root, "legacy-books.json")
	newPath := filepath.Join(root, "punkdom", "books.json")
	legacyData := bookRegistryData{
		Current: bookDir,
		Books: []BookRecord{{
			Name:         "旧书",
			Path:         bookDir,
			LastOpenedAt: "2026-01-01T00:00:00Z",
		}},
	}
	raw, err := json.Marshal(legacyData)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	registry := &BookRegistry{path: newPath, legacyPath: legacyPath}
	if got := registry.Current(); got != bookDir {
		t.Fatalf("未能读取旧注册表当前书籍: want=%s got=%s", bookDir, got)
	}
	if err := registry.Touch(bookDir); err != nil {
		t.Fatalf("保存迁移后的注册表失败: %v", err)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("新注册表未写入: %v", err)
	}
}

func TestUpdateBookInfoRenamesCurrentWorkspaceDirectory(t *testing.T) {
	root := t.TempDir()
	punkdomDir := filepath.Join(root, ".punkdom-user")
	oldWorkspace := filepath.Join(punkdomDir, "测试")
	if err := book.NewState(oldWorkspace).InitWorkspace(); err != nil {
		t.Fatalf("初始化旧工作区失败: %v", err)
	}

	application, err := New(context.Background(), &config.Config{
		PunkdomDir:          punkdomDir,
		Workspace:           oldWorkspace,
		ResumeLastWorkspace: false,
		OpenAIModel:         "test-model",
	})
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	result, err := application.UpdateBookInfo(oldWorkspace, "虚数时间", "Wilton", "简介")
	if err != nil {
		t.Fatalf("更新书籍信息失败: %v", err)
	}

	newWorkspace := filepath.Join(punkdomDir, "虚数时间")
	if result.Workspace != newWorkspace {
		t.Fatalf("返回的新 workspace 不符合预期: want=%s got=%s", newWorkspace, result.Workspace)
	}
	if application.Workspace() != newWorkspace {
		t.Fatalf("当前 workspace 应同步到新目录: want=%s got=%s", newWorkspace, application.Workspace())
	}
	if _, err := os.Stat(oldWorkspace); !os.IsNotExist(err) {
		t.Fatalf("旧目录应被重命名移走，实际错误: %v", err)
	}
	if _, err := os.Stat(filepath.Join(newWorkspace, ".punkdom")); err != nil {
		t.Fatalf("新目录应保留工作区数据: %v", err)
	}
	meta, err := application.BookInfo(newWorkspace)
	if err != nil {
		t.Fatalf("读取新路径元信息失败: %v", err)
	}
	if meta.Title != "虚数时间" || meta.Author != "Wilton" || meta.Description != "简介" {
		t.Fatalf("元信息未迁移到新路径: %#v", meta)
	}
	if got := application.bookRegistry.Current(); got != newWorkspace {
		t.Fatalf("注册表当前书籍应指向新路径: want=%s got=%s", newWorkspace, got)
	}
}

func TestProjectArchiveExportAndImportZip(t *testing.T) {
	root := t.TempDir()
	punkdomDir := filepath.Join(root, "punkdom")
	application, err := New(context.Background(), &config.Config{
		PunkdomDir:          punkdomDir,
		Workspace:           filepath.Join(root, "initial"),
		ResumeLastWorkspace: false,
		OpenAIModel:         "test-model",
	})
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}
	workspace, _, err := application.CreateBook(context.Background(), punkdomDir, "Source Project", "Wilton", "")
	if err != nil {
		t.Fatalf("创建源项目失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "chapters", "chapter.md"), []byte("hello project"), 0o644); err != nil {
		t.Fatalf("写入章节失败: %v", err)
	}

	archive, err := application.ExportProjectZip(workspace)
	if err != nil {
		t.Fatalf("导出项目失败: %v", err)
	}
	if archive.FileName != "Source Project.zip" {
		t.Fatalf("导出文件名不符合预期: %s", archive.FileName)
	}
	if !zipContains(t, archive.Data, "Source Project/chapters/chapter.md") {
		t.Fatalf("导出 zip 未包含项目文件")
	}
	targetPunkdomDir := filepath.Join(root, "target-punkdom")
	targetApp, err := New(context.Background(), &config.Config{
		PunkdomDir:          targetPunkdomDir,
		Workspace:           filepath.Join(root, "target-initial"),
		ResumeLastWorkspace: false,
		OpenAIModel:         "test-model",
	})
	if err != nil {
		t.Fatalf("创建目标应用失败: %v", err)
	}
	result, err := targetApp.ImportProjectZip(context.Background(), archive.FileName, bytes.NewReader(archive.Data), int64(len(archive.Data)))
	if err != nil {
		t.Fatalf("导入项目失败: %v", err)
	}
	if filepath.Base(result.Workspace) != "Source Project" {
		t.Fatalf("导入工作区名称不符合预期: %s", result.Workspace)
	}
	raw, err := os.ReadFile(filepath.Join(result.Workspace, "chapters", "chapter.md"))
	if err != nil {
		t.Fatalf("导入后章节不存在: %v", err)
	}
	if string(raw) != "hello project" {
		t.Fatalf("导入后章节内容不符合预期: %q", string(raw))
	}
	if got := targetApp.Workspace(); got != result.Workspace {
		t.Fatalf("导入后应切换到新项目: want=%s got=%s", result.Workspace, got)
	}
}

func TestImportProjectZipRejectsUnsafeEntries(t *testing.T) {
	root := t.TempDir()
	application, err := New(context.Background(), &config.Config{
		PunkdomDir:          filepath.Join(root, "punkdom"),
		Workspace:           filepath.Join(root, "initial"),
		ResumeLastWorkspace: false,
		OpenAIModel:         "test-model",
	})
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("../escape.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("bad")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := application.ImportProjectZip(context.Background(), "unsafe.zip", bytes.NewReader(buf.Bytes()), int64(buf.Len())); err == nil {
		t.Fatal("应拒绝包含越界路径的 zip")
	}
}

func zipContains(t *testing.T, data []byte, name string) bool {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("读取 zip 失败: %v", err)
	}
	for _, file := range reader.File {
		if file.Name == name {
			return true
		}
	}
	return false
}

func TestUpdateBookInfoRejectsExistingTargetDirectory(t *testing.T) {
	root := t.TempDir()
	punkdomDir := filepath.Join(root, ".punkdom-user")
	workspace := filepath.Join(punkdomDir, "测试")
	existing := filepath.Join(punkdomDir, "虚数时间")
	if err := book.NewState(workspace).InitWorkspace(); err != nil {
		t.Fatalf("初始化工作区失败: %v", err)
	}
	if err := book.NewState(existing).InitWorkspace(); err != nil {
		t.Fatalf("初始化冲突工作区失败: %v", err)
	}

	application, err := New(context.Background(), &config.Config{
		PunkdomDir:          punkdomDir,
		Workspace:           workspace,
		ResumeLastWorkspace: false,
		OpenAIModel:         "test-model",
	})
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	if _, err := application.UpdateBookInfo(workspace, "虚数时间", "", ""); !os.IsExist(err) {
		t.Fatalf("目标目录存在时应拒绝重命名，实际错误: %v", err)
	}
	if application.Workspace() != workspace {
		t.Fatalf("冲突后当前 workspace 不应改变: want=%s got=%s", workspace, application.Workspace())
	}
	if _, err := os.Stat(workspace); err != nil {
		t.Fatalf("冲突后旧目录应保留: %v", err)
	}
}
