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

	"punkdom/internal/book"
)

const maxProjectImportBytes int64 = 512 << 20

// ProjectArchive 表示一个已生成的项目导出包。
type ProjectArchive struct {
	FileName string
	Data     []byte
}

// ProjectImportResult 表示导入项目后的工作区路径和元信息。
type ProjectImportResult struct {
	Workspace string        `json:"workspace"`
	BookMeta  book.BookMeta `json:"book_meta"`
}

// ExportProjectZip 将指定项目目录打包为 zip，zip 根目录为项目目录名。
func (a *App) ExportProjectZip(path string) (ProjectArchive, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ProjectArchive{}, fmt.Errorf("路径无效: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return ProjectArchive{}, err
	}
	if !info.IsDir() {
		return ProjectArchive{}, fmt.Errorf("路径不是目录: %s", absPath)
	}
	if !isBookWorkspace(absPath) {
		return ProjectArchive{}, fmt.Errorf("不是有效的 Punkdom 项目目录: %s", absPath)
	}

	temp, err := os.CreateTemp("", "punkdom-project-*.zip")
	if err != nil {
		return ProjectArchive{}, err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()

	writer := zip.NewWriter(temp)
	rootName := sanitizeArchiveRootName(filepath.Base(absPath))
	if err := filepath.WalkDir(absPath, func(itemPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if itemPath == absPath {
			return nil
		}
		rel, err := filepath.Rel(absPath, itemPath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		name := rootName + "/" + rel
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
		return ProjectArchive{}, err
	}
	if err := writer.Close(); err != nil {
		_ = temp.Close()
		return ProjectArchive{}, err
	}
	if err := temp.Close(); err != nil {
		return ProjectArchive{}, err
	}
	data, err := os.ReadFile(tempPath)
	if err != nil {
		return ProjectArchive{}, err
	}
	return ProjectArchive{
		FileName: rootName + ".zip",
		Data:     data,
	}, nil
}

// ImportProjectZip 解包 Punkdom 项目 zip 到当前数据目录，并切换到导入后的项目。
func (a *App) ImportProjectZip(ctx context.Context, fileName string, reader io.Reader, size int64) (ProjectImportResult, error) {
	layered, err := a.Settings()
	if err != nil {
		return ProjectImportResult{}, err
	}
	if strings.TrimSpace(layered.Paths.PunkdomDir) == "" {
		return ProjectImportResult{}, errors.New("Punkdom 数据目录未设置")
	}
	if size > maxProjectImportBytes {
		return ProjectImportResult{}, fmt.Errorf("项目压缩包超过大小上限 %d MB", maxProjectImportBytes>>20)
	}
	temp, err := os.CreateTemp("", "punkdom-project-import-*.zip")
	if err != nil {
		return ProjectImportResult{}, err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	limited := io.LimitReader(reader, maxProjectImportBytes+1)
	written, err := io.Copy(temp, limited)
	closeErr := temp.Close()
	if err != nil {
		return ProjectImportResult{}, err
	}
	if closeErr != nil {
		return ProjectImportResult{}, closeErr
	}
	if written > maxProjectImportBytes {
		return ProjectImportResult{}, fmt.Errorf("项目压缩包超过大小上限 %d MB", maxProjectImportBytes>>20)
	}

	targetRoot, err := filepath.Abs(layered.Paths.PunkdomDir)
	if err != nil {
		return ProjectImportResult{}, fmt.Errorf("Punkdom 数据目录无效: %w", err)
	}
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return ProjectImportResult{}, err
	}

	readerZip, err := zip.OpenReader(tempPath)
	if err != nil {
		return ProjectImportResult{}, fmt.Errorf("打开项目 zip 失败: %w", err)
	}
	defer readerZip.Close()

	projectRoot, err := inferProjectRoot(fileName, readerZip.File)
	if err != nil {
		return ProjectImportResult{}, err
	}
	projectName := sanitizeArchiveRootName(projectRoot)
	targetDir := filepath.Join(targetRoot, projectName)
	if _, err := os.Stat(targetDir); err == nil {
		return ProjectImportResult{}, os.ErrExist
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return ProjectImportResult{}, err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return ProjectImportResult{}, err
	}
	extractOK := false
	defer func() {
		if !extractOK {
			_ = os.RemoveAll(targetDir)
		}
	}()
	if err := extractProjectZip(readerZip.File, projectRoot, targetDir); err != nil {
		return ProjectImportResult{}, err
	}
	if !isBookWorkspace(targetDir) {
		return ProjectImportResult{}, fmt.Errorf("zip 中未找到有效 Punkdom 项目结构")
	}
	meta, err := a.bookMetaStore.Read(targetDir)
	if err != nil {
		meta = book.BookMeta{Title: projectName}
	}
	if strings.TrimSpace(meta.Title) == "" {
		meta.Title = projectName
	}
	meta, err = a.bookMetaStore.Write(targetDir, meta)
	if err != nil {
		return ProjectImportResult{}, err
	}
	workspace, err := a.SwitchWorkspace(ctx, targetDir)
	if err != nil {
		return ProjectImportResult{}, err
	}
	extractOK = true
	log.Printf("[projects] imported zip=%q workspace=%q", fileName, workspace)
	return ProjectImportResult{Workspace: workspace, BookMeta: meta}, nil
}

func inferProjectRoot(fileName string, files []*zip.File) (string, error) {
	roots := make(map[string]bool)
	hasRootFile := false
	for _, file := range files {
		name, err := cleanZipEntryName(file.Name)
		if err != nil {
			return "", err
		}
		if name == "" {
			continue
		}
		parts := strings.Split(name, "/")
		if len(parts) == 1 {
			hasRootFile = true
			continue
		}
		roots[parts[0]] = true
	}
	if len(roots) == 1 && !hasRootFile {
		for root := range roots {
			return root, nil
		}
	}
	base := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
	base = strings.TrimSuffix(base, ".tar")
	return sanitizeArchiveRootName(base), nil
}

func extractProjectZip(files []*zip.File, archiveRoot, targetDir string) error {
	wrote := false
	for _, file := range files {
		name, err := cleanZipEntryName(file.Name)
		if err != nil {
			return err
		}
		if name == "" {
			continue
		}
		rel := name
		if archiveRoot != "" {
			prefix := strings.TrimSuffix(archiveRoot, "/") + "/"
			if name == archiveRoot {
				continue
			}
			if strings.HasPrefix(name, prefix) {
				rel = strings.TrimPrefix(name, prefix)
			}
		}
		if rel == "" {
			continue
		}
		targetPath := filepath.Join(targetDir, filepath.FromSlash(rel))
		if !isPathInside(targetDir, targetPath) {
			return fmt.Errorf("zip 条目越界: %s", file.Name)
		}
		mode := file.FileInfo().Mode()
		if file.FileInfo().IsDir() {
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
		src, err := file.Open()
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
		wrote = true
	}
	if !wrote {
		return errors.New("zip 中没有可导入的项目文件")
	}
	return nil
}

func cleanZipEntryName(name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "/")
	clean := filepath.ToSlash(filepath.Clean(name))
	if clean == "." {
		return "", nil
	}
	if strings.HasPrefix(clean, "../") || clean == ".." || filepath.IsAbs(clean) {
		return "", fmt.Errorf("zip 条目路径不安全: %s", name)
	}
	return clean, nil
}

func sanitizeArchiveRootName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "punkdom-project"
	}
	name = filepath.Base(name)
	name = strings.TrimSpace(strings.TrimSuffix(name, ".zip"))
	if err := book.ValidateNewName(name); err != nil {
		return "punkdom-project"
	}
	return name
}

func isPathInside(root, path string) bool {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}
