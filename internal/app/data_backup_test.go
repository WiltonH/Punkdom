package app

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"punkdom/config"
)

func TestDataBackupExportAndRestore(t *testing.T) {
	root := t.TempDir()
	punkdomDir := filepath.Join(root, ".punkdom")
	if err := os.MkdirAll(filepath.Join(punkdomDir, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	targetFile := filepath.Join(punkdomDir, "demo", "note.md")
	if err := os.WriteFile(targetFile, []byte("backup version"), 0o644); err != nil {
		t.Fatal(err)
	}
	application, err := New(context.Background(), &config.Config{
		PunkdomDir:          punkdomDir,
		ResumeLastWorkspace: false,
		OpenAIModel:         "test-model",
	})
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	archive, err := application.ExportDataBackupZip()
	if err != nil {
		t.Fatalf("导出备份失败: %v", err)
	}
	if archive.FileName == "" || !zipContains(t, archive.Data, ".punkdom/demo/note.md") {
		t.Fatalf("备份 zip 内容不符合预期: %s", archive.FileName)
	}
	if err := os.WriteFile(targetFile, []byte("mutated"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(punkdomDir, "extra.txt"), []byte("remove me"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := application.RestoreDataBackupZip(context.Background(), archive.FileName, bytes.NewReader(archive.Data), int64(len(archive.Data))); err != nil {
		t.Fatalf("还原备份失败: %v", err)
	}
	raw, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("还原后文件不存在: %v", err)
	}
	if string(raw) != "backup version" {
		t.Fatalf("还原后文件内容不符合预期: %q", string(raw))
	}
	if _, err := os.Stat(filepath.Join(punkdomDir, "extra.txt")); !os.IsNotExist(err) {
		t.Fatalf("覆盖还原应移除备份中不存在的文件: %v", err)
	}
}

func TestDataBackupRestoreRejectsUnsafeZipEntries(t *testing.T) {
	root := t.TempDir()
	application, err := New(context.Background(), &config.Config{
		PunkdomDir:          filepath.Join(root, ".punkdom"),
		ResumeLastWorkspace: false,
		OpenAIModel:         "test-model",
	})
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}
	data := testZipBytes(t, map[string]string{
		"../escape.txt": "bad",
	})
	if _, err := application.RestoreDataBackupZip(context.Background(), "unsafe.zip", bytes.NewReader(data), int64(len(data))); err == nil {
		t.Fatal("应拒绝包含越界路径的备份 zip")
	}
}

func testZipBytes(t *testing.T, files map[string]string) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)
	for name, content := range files {
		w, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
