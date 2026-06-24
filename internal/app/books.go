package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"punkdom/internal/book"
)

const maxBookRecords = 20

// BookRecord 表示 Punkdom 数据目录中的一个书籍工作目录。
type BookRecord struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Author       string `json:"author"`
	Description  string `json:"description,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	LastOpenedAt string `json:"last_opened_at"`
}

// DeletedBookRecord 表示移入 Punkdom Trash 的项目，可恢复或彻底删除。
type DeletedBookRecord struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	OriginalPath string `json:"original_path"`
	Author       string `json:"author"`
	Description  string `json:"description,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	DeletedAt    string `json:"deleted_at"`
}

type bookRegistryData struct {
	Current string              `json:"current"`
	Books   []BookRecord        `json:"books"`
	Order   []string            `json:"order,omitempty"`
	Hidden  []string            `json:"hidden,omitempty"`
	Deleted []DeletedBookRecord `json:"deleted,omitempty"`
}

// BookRegistry 持久化当前书籍，并从 Punkdom 数据目录发现实际存在的书籍工作目录。
type BookRegistry struct {
	path       string
	legacyPath string
	punkdomDir string
}

// NewBookRegistry 创建书籍记录管理器。
func NewBookRegistry(punkdomDir string) *BookRegistry {
	return &BookRegistry{
		path:       filepath.Join(punkdomDir, "books.json"),
		legacyPath: legacyBookRegistryPath(),
		punkdomDir: punkdomDir,
	}
}

// Current 返回上次打开且仍存在的工作目录。
func (r *BookRegistry) Current() string {
	data := r.load()
	if data.Current == "" {
		return ""
	}
	current, err := filepath.Abs(data.Current)
	if err != nil {
		return ""
	}
	if pathSet(data.Hidden)[current] {
		return ""
	}
	if info, err := os.Stat(current); err == nil && info.IsDir() {
		return current
	}
	return ""
}

// List 返回当前 Punkdom 数据目录下实际存在的书籍列表。
func (r *BookRegistry) List() []BookRecord {
	data := r.load()
	if strings.TrimSpace(r.punkdomDir) == "" {
		return sortedRegistryBooks(data)
	}

	books, err := r.scanPunkdomBooks(data)
	if err == nil {
		return books
	}
	return sortedRegistryBooks(data)
}

func sortedRegistryBooks(data bookRegistryData) []BookRecord {
	hidden := pathSet(data.Hidden)
	books := make([]BookRecord, 0, len(data.Books))
	for _, book := range data.Books {
		if book.Path == "" {
			continue
		}
		absPath, err := filepath.Abs(book.Path)
		if err != nil || hidden[absPath] {
			continue
		}
		book.Path = absPath
		books = append(books, book)
	}
	if len(data.Order) > 0 {
		sortBooksByOrder(books, data.Order, func(i, j int) bool {
			return books[i].LastOpenedAt > books[j].LastOpenedAt
		})
		return books
	}
	sort.SliceStable(books, func(i, j int) bool {
		return books[i].LastOpenedAt > books[j].LastOpenedAt
	})
	return books
}

func (r *BookRegistry) scanPunkdomBooks(data bookRegistryData) ([]BookRecord, error) {
	absPunkdomDir, err := filepath.Abs(r.punkdomDir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(absPunkdomDir)
	if err != nil {
		return nil, err
	}

	openedAt := make(map[string]string, len(data.Books))
	for _, book := range data.Books {
		if book.Path == "" {
			continue
		}
		absPath, err := filepath.Abs(book.Path)
		if err != nil {
			continue
		}
		openedAt[absPath] = book.LastOpenedAt
	}
	hidden := pathSet(data.Hidden)

	books := make([]BookRecord, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || isPunkdomUserDataDir(entry.Name()) {
			continue
		}
		bookPath := filepath.Join(absPunkdomDir, entry.Name())
		if !isBookWorkspace(bookPath) {
			continue
		}
		if hidden[bookPath] {
			continue
		}
		books = append(books, BookRecord{
			Name:         entry.Name(),
			Path:         bookPath,
			LastOpenedAt: openedAt[bookPath],
		})
	}

	sortBooksByOrder(books, data.Order, func(i, j int) bool {
		return strings.ToLower(books[i].Name) < strings.ToLower(books[j].Name)
	})
	return books, nil
}

func sortBooksByOrder(books []BookRecord, order []string, fallback func(i, j int) bool) {
	rank := make(map[string]int, len(order))
	for i, path := range order {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, exists := rank[absPath]; !exists {
				rank[absPath] = i
			}
		}
	}
	sort.SliceStable(books, func(i, j int) bool {
		leftRank, leftOrdered := rank[books[i].Path]
		rightRank, rightOrdered := rank[books[j].Path]
		if leftOrdered && rightOrdered {
			return leftRank < rightRank
		}
		if leftOrdered != rightOrdered {
			return leftOrdered
		}
		return fallback(i, j)
	})
}

func pathSet(paths []string) map[string]bool {
	set := make(map[string]bool, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		set[absPath] = true
	}
	return set
}

func isPunkdomUserDataDir(name string) bool {
	switch name {
	case "book_meta", "styles", "Trash":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}

func isBookWorkspace(path string) bool {
	markers := []string{
		filepath.Join(path, ".punkdom"),
		filepath.Join(path, "book.json"),
		filepath.Join(path, "ideas.md"),
		filepath.Join(path, "brainstorm.md"),
		filepath.Join(path, "chapters"),
		filepath.Join(path, "setting"),
	}
	for _, marker := range markers {
		if _, err := os.Stat(marker); err == nil {
			return true
		}
	}
	return false
}

// Touch 记录并置顶一个书籍工作目录。
func (r *BookRegistry) Touch(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("路径不是目录")
	}

	data := r.load()
	now := time.Now().Format(time.RFC3339)
	record := BookRecord{
		Name:         filepath.Base(absPath),
		Path:         absPath,
		LastOpenedAt: now,
	}
	data.Hidden = removePath(data.Hidden, absPath)
	if len(data.Order) > 0 {
		found := false
		for i, book := range data.Books {
			bookPath, err := filepath.Abs(book.Path)
			if err == nil && bookPath == absPath {
				data.Books[i] = record
				found = true
				break
			}
		}
		if !found {
			data.Books = append(data.Books, record)
		}
		if !pathSet(data.Order)[absPath] {
			data.Order = append(data.Order, absPath)
		}
	} else {
		books := []BookRecord{record}
		for _, book := range data.Books {
			bookPath, err := filepath.Abs(book.Path)
			if book.Path == "" || (err == nil && bookPath == absPath) {
				continue
			}
			books = append(books, book)
			if len(books) >= maxBookRecords {
				break
			}
		}
		data.Books = books
	}
	data.Current = absPath
	return r.save(data)
}

// MoveToTrash 将项目目录移动到 Punkdom Trash，并从正常项目列表移除。
func (r *BookRegistry) MoveToTrash(path string) (DeletedBookRecord, error) {
	return r.MoveToTrashWithMeta(path, book.BookMeta{})
}

// MoveToTrashWithMeta 将项目移入 Trash，并把调用方已读取的项目元信息写入删除记录。
func (r *BookRegistry) MoveToTrashWithMeta(path string, meta book.BookMeta) (DeletedBookRecord, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return DeletedBookRecord{}, err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return DeletedBookRecord{}, err
	}
	if !info.IsDir() {
		return DeletedBookRecord{}, errors.New("路径不是目录")
	}
	absPunkdomDir, err := filepath.Abs(r.punkdomDir)
	if err != nil || strings.TrimSpace(absPunkdomDir) == "" {
		return DeletedBookRecord{}, errors.New("Punkdom 数据目录未配置")
	}
	if !isPathInside(absPunkdomDir, absPath) || isPathInside(r.trashDir(), absPath) {
		return DeletedBookRecord{}, fmt.Errorf("只能删除 Punkdom 数据目录内的项目: %s", absPath)
	}

	deletedAt := time.Now().Format(time.RFC3339)
	trashDir := r.trashDir()
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return DeletedBookRecord{}, err
	}
	baseName := filepath.Base(absPath)
	trashPath := filepath.Join(trashDir, baseName)
	if _, err := os.Stat(trashPath); err == nil {
		trashPath = filepath.Join(trashDir, baseName+"-"+time.Now().Format("20060102-150405"))
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return DeletedBookRecord{}, err
	}
	if err := os.Rename(absPath, trashPath); err != nil {
		return DeletedBookRecord{}, fmt.Errorf("移入 Trash 失败: %w", err)
	}
	data := r.load()
	record := DeletedBookRecord{
		Name:         baseName,
		Path:         trashPath,
		OriginalPath: absPath,
		Author:       meta.Author,
		Description:  meta.Description,
		CreatedAt:    meta.CreatedAt,
		UpdatedAt:    meta.UpdatedAt,
		DeletedAt:    deletedAt,
	}
	if meta.Title != "" {
		record.Name = meta.Title
	}
	data.Hidden = appendUniquePath(data.Hidden, absPath)
	books := make([]BookRecord, 0, len(data.Books))
	for _, book := range data.Books {
		bookPath, err := filepath.Abs(book.Path)
		if err == nil && bookPath == absPath {
			if record.Name == "" || record.Name == baseName {
				record.Name = book.Name
			}
			if record.Author == "" {
				record.Author = book.Author
			}
			if record.Description == "" {
				record.Description = book.Description
			}
			if record.CreatedAt == "" {
				record.CreatedAt = book.CreatedAt
			}
			if record.UpdatedAt == "" {
				record.UpdatedAt = book.UpdatedAt
			}
			continue
		}
		books = append(books, book)
	}
	data.Order = removePath(data.Order, absPath)
	current, _ := filepath.Abs(data.Current)
	if current == absPath {
		data.Current = ""
		if len(books) > 0 {
			data.Current = books[0].Path
		}
	}
	data.Books = books
	data.Deleted = append([]DeletedBookRecord{record}, removeDeletedByPaths(data.Deleted, absPath, trashPath)...)
	return record, r.save(data)
}

// Deleted 返回仍存在于 Trash 中的已删除项目。
func (r *BookRegistry) Deleted() []DeletedBookRecord {
	data := r.load()
	deleted := make([]DeletedBookRecord, 0, len(data.Deleted))
	for _, record := range data.Deleted {
		if record.Path == "" {
			continue
		}
		absPath, err := filepath.Abs(record.Path)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			record.Path = absPath
			deleted = append(deleted, record)
		}
	}
	sort.SliceStable(deleted, func(i, j int) bool {
		return deleted[i].DeletedAt > deleted[j].DeletedAt
	})
	return deleted
}

// Restore 从 Trash 恢复项目目录，并重新加入项目列表。
func (r *BookRegistry) Restore(trashPath string) (string, error) {
	absTrashPath, err := filepath.Abs(trashPath)
	if err != nil {
		return "", err
	}
	data := r.load()
	var record DeletedBookRecord
	found := false
	for _, item := range data.Deleted {
		itemPath, err := filepath.Abs(item.Path)
		if err == nil && itemPath == absTrashPath {
			record = item
			found = true
			break
		}
	}
	if !found {
		return "", errors.New("已删除项目不存在")
	}
	target := record.OriginalPath
	if strings.TrimSpace(target) == "" {
		target = filepath.Join(r.punkdomDir, filepath.Base(absTrashPath))
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(target); err == nil {
		return "", os.ErrExist
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.Rename(absTrashPath, target); err != nil {
		return "", fmt.Errorf("恢复项目失败: %w", err)
	}
	data.Deleted = removeDeletedByPaths(data.Deleted, record.OriginalPath, absTrashPath)
	data.Hidden = removePath(data.Hidden, target)
	data.Books = append([]BookRecord{{
		Name:         filepath.Base(target),
		Path:         target,
		Author:       record.Author,
		LastOpenedAt: time.Now().Format(time.RFC3339),
	}}, data.Books...)
	data.Current = target
	if err := r.save(data); err != nil {
		return "", err
	}
	return target, nil
}

// Purge 永久删除 Trash 中的项目目录。
func (r *BookRegistry) Purge(trashPath string) error {
	absTrashPath, err := filepath.Abs(trashPath)
	if err != nil {
		return err
	}
	if !isPathInside(r.trashDir(), absTrashPath) {
		return fmt.Errorf("只能彻底删除 Trash 内的项目: %s", absTrashPath)
	}
	if err := os.RemoveAll(absTrashPath); err != nil {
		return err
	}
	data := r.load()
	data.Deleted = removeDeletedByPaths(data.Deleted, "", absTrashPath)
	return r.save(data)
}

func (r *BookRegistry) trashDir() string {
	return filepath.Join(r.punkdomDir, "Trash")
}

// Reorder 保存书籍管理页的自定义排序。
func (r *BookRegistry) Reorder(paths []string) error {
	data := r.load()
	seen := make(map[string]bool, len(paths))
	order := make([]string, 0, len(paths))
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil || seen[absPath] {
			continue
		}
		seen[absPath] = true
		order = append(order, absPath)
	}
	for _, book := range r.List() {
		if !seen[book.Path] {
			order = append(order, book.Path)
		}
	}
	data.Order = order
	return r.save(data)
}

// ReplacePath 将注册表中的旧 workspace 路径替换为新路径，并保持当前书籍、排序和隐藏状态一致。
func (r *BookRegistry) ReplacePath(oldPath, newPath string) error {
	oldAbs, err := filepath.Abs(oldPath)
	if err != nil {
		return err
	}
	newAbs, err := filepath.Abs(newPath)
	if err != nil {
		return err
	}

	data := r.load()
	for i, book := range data.Books {
		bookPath, err := filepath.Abs(book.Path)
		if err == nil && bookPath == oldAbs {
			data.Books[i].Path = newAbs
			data.Books[i].Name = filepath.Base(newAbs)
		}
	}
	for i, path := range data.Order {
		orderPath, err := filepath.Abs(path)
		if err == nil && orderPath == oldAbs {
			data.Order[i] = newAbs
		}
	}
	data.Hidden = removePath(data.Hidden, oldAbs)
	current, err := filepath.Abs(data.Current)
	if err == nil && current == oldAbs {
		data.Current = newAbs
	}
	return r.save(data)
}

func appendUniquePath(paths []string, path string) []string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return paths
	}
	for _, item := range paths {
		itemAbs, err := filepath.Abs(item)
		if err == nil && itemAbs == absPath {
			return paths
		}
	}
	return append(paths, absPath)
}

func removePath(paths []string, path string) []string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return paths
	}
	next := make([]string, 0, len(paths))
	for _, item := range paths {
		itemAbs, err := filepath.Abs(item)
		if err != nil || itemAbs == absPath {
			continue
		}
		next = append(next, itemAbs)
	}
	return next
}

func removeDeletedByPaths(records []DeletedBookRecord, originalPath, trashPath string) []DeletedBookRecord {
	originalAbs, _ := filepath.Abs(originalPath)
	trashAbs, _ := filepath.Abs(trashPath)
	next := make([]DeletedBookRecord, 0, len(records))
	for _, record := range records {
		recordOriginal, _ := filepath.Abs(record.OriginalPath)
		recordTrash, _ := filepath.Abs(record.Path)
		if originalAbs != "" && recordOriginal == originalAbs {
			continue
		}
		if trashAbs != "" && recordTrash == trashAbs {
			continue
		}
		next = append(next, record)
	}
	return next
}

func (r *BookRegistry) load() bookRegistryData {
	var data bookRegistryData
	raw, err := os.ReadFile(r.path)
	if err != nil && r.legacyPath != "" && r.legacyPath != r.path {
		raw, err = os.ReadFile(r.legacyPath)
	}
	if err != nil {
		return data
	}
	_ = json.Unmarshal(raw, &data)
	return data
}

func (r *BookRegistry) save(data bookRegistryData) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, raw, 0o644)
}

func legacyBookRegistryPath() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "punkdom", "books.json")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".punkdom", "books.json")
	}
	return filepath.Join(".", ".punkdom-books.json")
}
