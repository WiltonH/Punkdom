package app

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxDataBackupImportBytes int64 = 2 << 30

// DataBackupArchive 表示一个完整 Punkdom 数据目录备份包。
type DataBackupArchive struct {
	FileName string
	Data     []byte
}

// DataBackupRestoreResult 描述完整数据备份还原后的运行状态。
type DataBackupRestoreResult struct {
	Workspace string `json:"workspace"`
	Message   string `json:"message"`
}

// ExportDataBackupZip 将整个 Punkdom 数据目录打包为 zip，zip 根目录固定为 .punkdom。
func (a *App) ExportDataBackupZip() (DataBackupArchive, error) {
	punkdomDir, err := a.effectivePunkdomDir()
	if err != nil {
		return DataBackupArchive{}, err
	}
	info, err := os.Stat(punkdomDir)
	if err != nil {
		return DataBackupArchive{}, err
	}
	if !info.IsDir() {
		return DataBackupArchive{}, fmt.Errorf("Punkdom 数据目录不是目录: %s", punkdomDir)
	}

	temp, err := os.CreateTemp("", "punkdom-data-backup-*.zip")
	if err != nil {
		return DataBackupArchive{}, err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()

	writer := zip.NewWriter(temp)
	if err := filepath.WalkDir(punkdomDir, func(itemPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if itemPath == punkdomDir {
			return nil
		}
		rel, err := filepath.Rel(punkdomDir, itemPath)
		if err != nil {
			return err
		}
		name := ".punkdom/" + filepath.ToSlash(rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			header := &zip.FileHeader{Name: name + "/", Method: zip.Deflate}
			header.SetMode(info.Mode())
			_, err := writer.CreateHeader(header)
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = name
		header.Method = zip.Deflate
		w, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		f, err := os.Open(itemPath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	}); err != nil {
		_ = writer.Close()
		_ = temp.Close()
		return DataBackupArchive{}, err
	}
	if err := writer.Close(); err != nil {
		_ = temp.Close()
		return DataBackupArchive{}, err
	}
	if err := temp.Close(); err != nil {
		return DataBackupArchive{}, err
	}
	data, err := os.ReadFile(tempPath)
	if err != nil {
		return DataBackupArchive{}, err
	}
	return DataBackupArchive{
		FileName: "Punkdom-" + time.Now().Format("20060102-150405") + ".zip",
		Data:     data,
	}, nil
}

// RestoreDataBackupZip 覆盖还原整个 Punkdom 数据目录，并尽量恢复到可用 workspace。
func (a *App) RestoreDataBackupZip(ctx context.Context, fileName string, reader io.Reader, size int64) (DataBackupRestoreResult, error) {
	if size > maxDataBackupImportBytes {
		return DataBackupRestoreResult{}, fmt.Errorf("备份压缩包超过大小上限 %d MB", maxDataBackupImportBytes>>20)
	}
	punkdomDir, err := a.effectivePunkdomDir()
	if err != nil {
		return DataBackupRestoreResult{}, err
	}
	temp, err := os.CreateTemp("", "punkdom-data-restore-*.zip")
	if err != nil {
		return DataBackupRestoreResult{}, err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()

	written, err := io.Copy(temp, io.LimitReader(reader, maxDataBackupImportBytes+1))
	closeErr := temp.Close()
	if err != nil {
		return DataBackupRestoreResult{}, err
	}
	if closeErr != nil {
		return DataBackupRestoreResult{}, closeErr
	}
	if written > maxDataBackupImportBytes {
		return DataBackupRestoreResult{}, fmt.Errorf("备份压缩包超过大小上限 %d MB", maxDataBackupImportBytes>>20)
	}

	readerZip, err := zip.OpenReader(tempPath)
	if err != nil {
		return DataBackupRestoreResult{}, fmt.Errorf("打开备份 zip 失败: %w", err)
	}
	defer readerZip.Close()

	files, err := prepareDataBackupEntries(fileName, readerZip.File)
	if err != nil {
		return DataBackupRestoreResult{}, err
	}
	if err := os.MkdirAll(punkdomDir, 0o755); err != nil {
		return DataBackupRestoreResult{}, err
	}
	if err := clearDirectoryContents(punkdomDir); err != nil {
		return DataBackupRestoreResult{}, fmt.Errorf("清空 Punkdom 数据目录失败: %w", err)
	}
	if err := extractDataBackup(files, punkdomDir); err != nil {
		return DataBackupRestoreResult{}, err
	}

	workspace, err := a.recoverWorkspaceAfterDataRestore(ctx)
	if err != nil {
		log.Printf("[backup] restore completed but workspace recovery failed file=%q err=%v", fileName, err)
	}
	log.Printf("[backup] restored data backup file=%q punkdom_dir=%q workspace=%q", fileName, punkdomDir, workspace)
	return DataBackupRestoreResult{Workspace: workspace, Message: "restored"}, nil
}

func (a *App) effectivePunkdomDir() (string, error) {
	layered, err := a.Settings()
	if err != nil {
		return "", err
	}
	punkdomDir := strings.TrimSpace(layered.Paths.PunkdomDir)
	if punkdomDir == "" {
		return "", errors.New("Punkdom 数据目录未设置")
	}
	abs, err := filepath.Abs(punkdomDir)
	if err != nil {
		return "", fmt.Errorf("Punkdom 数据目录无效: %w", err)
	}
	return abs, nil
}

type dataBackupEntry struct {
	file *zip.File
	rel  string
}

func prepareDataBackupEntries(fileName string, files []*zip.File) ([]dataBackupEntry, error) {
	_ = fileName
	prepared := make([]dataBackupEntry, 0, len(files))
	hasRoot := false
	for _, file := range files {
		name, err := cleanZipEntryName(file.Name)
		if err != nil {
			return nil, err
		}
		if name == "" {
			continue
		}
		rel := name
		if name == ".punkdom" {
			hasRoot = true
			continue
		}
		if strings.HasPrefix(name, ".punkdom/") {
			hasRoot = true
			rel = strings.TrimPrefix(name, ".punkdom/")
		}
		if rel == "" {
			continue
		}
		prepared = append(prepared, dataBackupEntry{file: file, rel: rel})
	}
	if len(prepared) == 0 {
		return nil, errors.New("备份 zip 中没有可还原的数据")
	}
	if !hasRoot {
		log.Printf("[backup] restoring zip without .punkdom root; treating entries as Punkdom data root")
	}
	return prepared, nil
}

func extractDataBackup(entries []dataBackupEntry, targetDir string) error {
	for _, entry := range entries {
		targetPath := filepath.Join(targetDir, filepath.FromSlash(entry.rel))
		if !isPathInside(targetDir, targetPath) {
			return fmt.Errorf("zip 条目越界: %s", entry.file.Name)
		}
		mode := entry.file.FileInfo().Mode()
		if entry.file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, mode.Perm()); err != nil {
				return err
			}
			continue
		}
		if !mode.IsRegular() {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		src, err := entry.file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
		if err != nil {
			_ = src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		closeSrcErr := src.Close()
		closeDstErr := dst.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeSrcErr != nil {
			return closeSrcErr
		}
		if closeDstErr != nil {
			return closeDstErr
		}
	}
	return nil
}

func clearDirectoryContents(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) recoverWorkspaceAfterDataRestore(ctx context.Context) (string, error) {
	if current := a.bookRegistry.Current(); current != "" {
		return a.SwitchWorkspace(ctx, current)
	}
	for _, record := range a.bookRegistry.List() {
		if record.Path == "" {
			continue
		}
		return a.SwitchWorkspace(ctx, record.Path)
	}
	a.mu.Lock()
	a.workspace = ""
	a.bookState = nil
	a.bookService = nil
	a.interactive = nil
	a.sessionStore = nil
	a.session = nil
	a.agentRunner = nil
	a.interactiveStoryRunner = nil
	a.versionService = nil
	a.cfg.Workspace = ""
	a.mu.Unlock()
	return "", nil
}
