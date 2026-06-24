import { useEffect, useRef, useState } from 'react'
import type { ComponentProps, ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { DndContext, KeyboardSensor, PointerSensor, closestCenter, useSensor, useSensors, type DragEndEvent } from '@dnd-kit/core'
import { SortableContext, arrayMove, rectSortingStrategy, sortableKeyboardCoordinates, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { BookOpen, Check, ChevronDown, Download, FileArchive, FileText, Folder, GripVertical, LibraryBig, Pencil, Plus, RotateCcw, Trash2, Upload, X } from 'lucide-react'
import { toast } from 'sonner'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { TooltipIconButton } from '@/components/common/tooltip-icon-button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { NovelImportDialog } from './NovelImportDialog'
import {
  createBook,
  exportProjectZip,
  getDeletedBooks,
  getBookInfo,
  importProjectZip,
  purgeDeletedBook,
  removeBook,
  reorderBooks,
  restoreBook,
  switchWorkspace,
  updateBookInfo,
  type BookMeta,
  type BookRecord,
  type DeletedBookRecord,
} from '@/lib/api'

interface HomeViewProps {
  /** 当前工作区路径，用于高亮当前书籍并作为父目录推断默认值 */
  workspace: string
  /** 用户 Punkdom 数据目录，新建书籍默认创建在该目录下 */
  punkdomDir: string
  /** Punkdom 数据目录下实际存在的书籍 */
  books: BookRecord[]
  /** 切换到指定 workspace 后由父组件刷新业务状态 */
  onSwitch: (path: string) => void
  /** 更新当前项目元信息后刷新当前 workspace，不改变当前一级菜单。 */
  onRefreshCurrentBook?: (path: string) => void
  /** 书籍记录有变更时通知父组件刷新列表 */
  onBooksChange: () => void
  /** 打开酒馆角色卡导入弹窗 */
  onOpenCharacterCardImport?: () => void
  /** 关闭全局项目仓库弹窗 */
  onClose?: () => void
}

const inputCls = 'punkdom-field w-full rounded-[var(--punkdom-radius)] border px-2.5 py-1.5 outline-none placeholder:text-[var(--punkdom-text-faint)] focus:border-[var(--punkdom-field-focus-border)] focus:bg-[var(--punkdom-surface-3)]'
const ghostButtonCls = 'punkdom-nav-item border border-transparent bg-transparent text-[var(--punkdom-text-muted)] hover:bg-[var(--punkdom-hover)] hover:text-[var(--punkdom-text)]'
const primaryButtonCls = 'border border-[var(--punkdom-border)] bg-[var(--punkdom-active)] text-[var(--punkdom-text)] hover:bg-[var(--punkdom-hover)]'
const iconButtonCls = 'punkdom-nav-item text-[var(--punkdom-text-faint)] hover:bg-[var(--punkdom-hover)] hover:text-[var(--punkdom-text)]'

/** 项目仓库视图：集中展示、创建、打开和编辑 Punkdom 数据目录中的书籍。 */
export function HomeView({ workspace, punkdomDir, books, onSwitch, onRefreshCurrentBook, onBooksChange, onOpenCharacterCardImport, onClose }: HomeViewProps) {
  const { t } = useTranslation()
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [createTitle, setCreateTitle] = useState('')
  const [createAuthor, setCreateAuthor] = useState('')
  const [createDesc, setCreateDesc] = useState('')
  const [createError, setCreateError] = useState('')
  const [creating, setCreating] = useState(false)
  const [showNovelImport, setShowNovelImport] = useState(false)

  const [editingBookPath, setEditingBookPath] = useState<string | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [editAuthor, setEditAuthor] = useState('')
  const [editDesc, setEditDesc] = useState('')
  const [editLoading, setEditLoading] = useState(false)
  const [editSaving, setEditSaving] = useState(false)
  const [orderedBooks, setOrderedBooks] = useState<BookRecord[]>(books)
  const [deletedBooks, setDeletedBooks] = useState<DeletedBookRecord[]>([])
  const [deletedExpanded, setDeletedExpanded] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<BookRecord | null>(null)
  const [purgeTarget, setPurgeTarget] = useState<DeletedBookRecord | null>(null)
  const [deleteError, setDeleteError] = useState('')
  const [deleting, setDeleting] = useState(false)
  const [restoringPath, setRestoringPath] = useState<string | null>(null)
  const [purging, setPurging] = useState(false)
  const [importingProject, setImportingProject] = useState(false)
  const [downloadingPath, setDownloadingPath] = useState<string | null>(null)
  const projectImportInputRef = useRef<HTMLInputElement | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )

  useEffect(() => {
    setOrderedBooks(books)
  }, [books])

  useEffect(() => {
    void refreshDeletedBooks()
  }, [])

  const refreshDeletedBooks = async () => {
    try {
      setDeletedBooks(await getDeletedBooks())
    } catch (e) {
      console.error('[home] load deleted projects failed', e)
    }
  }

  /** 打开新建书籍表单，新书统一创建在用户 Punkdom 数据目录下 */
  const openCreateForm = () => {
    setShowCreateForm(true)
    setCreateTitle('')
    setCreateAuthor('')
    setCreateDesc('')
    setCreateError('')
  }

  /** 提交新建书籍 */
  const handleCreate = async () => {
    if (!createTitle.trim()) { setCreateError(t('home.titleRequired')); return }
    if (!punkdomDir.trim()) { setCreateError(t('home.waitPunkdomDir')); return }
    setCreating(true)
    setCreateError('')
    try {
      const data = await createBook(
        createTitle.trim(),
        createAuthor.trim() || undefined,
        createDesc.trim() || undefined,
      )
      onSwitch(data.workspace)
      setShowCreateForm(false)
      onBooksChange()
    } catch (e: unknown) {
      setCreateError(e instanceof Error ? e.message : t('home.createError'))
    } finally {
      setCreating(false)
    }
  }

  /** 切换到指定书籍 */
  const handleSwitch = async (path: string) => {
    try {
      const data = await switchWorkspace(path)
      onSwitch(data.workspace || path)
    } catch (e) {
      console.error('切换 workspace 失败', e)
    }
  }

  /** 进入编辑模式，先拉取完整元信息 */
  const startEdit = async (book: BookRecord) => {
    setEditingBookPath(book.path)
    setEditTitle(book.name)
    setEditAuthor(book.author || '')
    setEditDesc('')
    setEditLoading(true)
    try {
      const meta: BookMeta = await getBookInfo(book.path)
      setEditTitle(meta.title)
      setEditAuthor(meta.author)
      setEditDesc(meta.description)
    } catch {
      // 拉取失败时回退使用列表里的基础信息
    } finally {
      setEditLoading(false)
    }
  }

  /** 保存书籍编辑 */
  const handleSaveEdit = async () => {
    if (!editingBookPath) return
    setEditSaving(true)
    try {
      const result = await updateBookInfo(editingBookPath, editTitle.trim(), editAuthor.trim(), editDesc.trim())
      setEditingBookPath(null)
      if (editingBookPath === workspace && result.workspace) {
        if (onRefreshCurrentBook) {
          onRefreshCurrentBook(result.workspace)
        } else {
          onSwitch(result.workspace)
        }
      }
      onBooksChange()
    } catch (e) {
      console.error('保存书籍信息失败', e)
    } finally {
      setEditSaving(false)
    }
  }

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIndex = orderedBooks.findIndex((book) => book.path === active.id)
    const newIndex = orderedBooks.findIndex((book) => book.path === over.id)
    if (oldIndex === -1 || newIndex === -1) return
    const nextBooks = arrayMove(orderedBooks, oldIndex, newIndex)
    setOrderedBooks(nextBooks)
    try {
      await reorderBooks(nextBooks.map((book) => book.path))
      await onBooksChange()
    } catch (e) {
      console.error('保存书籍排序失败', e)
      setOrderedBooks(books)
    }
  }

  const openDeleteDialog = (book: BookRecord) => {
    setDeleteTarget(book)
    setDeleteError('')
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    setDeleteError('')
    try {
      const result = await removeBook(deleteTarget.path)
      if (deleteTarget.path === workspace) {
        onSwitch(result.workspace || '')
      } else {
        await onBooksChange()
      }
      await refreshDeletedBooks()
      toast.success(t('home.moveToTrashSuccess'))
      setDeleteTarget(null)
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : String(e))
    } finally {
      setDeleting(false)
    }
  }

  const handleRestoreDeleted = async (book: DeletedBookRecord) => {
    if (restoringPath) return
    setRestoringPath(book.path)
    try {
      const result = await restoreBook(book.path)
      if (onRefreshCurrentBook) {
        onRefreshCurrentBook(result.workspace)
      } else {
        onSwitch(result.workspace)
      }
      await onBooksChange()
      await refreshDeletedBooks()
      toast.success(t('home.restoreSuccess', { name: book.name || t('home.unnamedBook') }))
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e)
      console.error('[home] restore deleted project failed', { path: book.path, error: e })
      toast.error(t('home.restoreFailed'), { description: message })
    } finally {
      setRestoringPath(null)
    }
  }

  const handlePurgeDeleted = async () => {
    if (!purgeTarget) return
    setPurging(true)
    try {
      await purgeDeletedBook(purgeTarget.path)
      await refreshDeletedBooks()
      toast.success(t('home.purgeSuccess', { name: purgeTarget.name || t('home.unnamedBook') }))
      setPurgeTarget(null)
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e)
      console.error('[home] purge deleted project failed', { path: purgeTarget.path, error: e })
      toast.error(t('home.purgeFailed'), { description: message })
    } finally {
      setPurging(false)
    }
  }

  const handleDownloadProject = async (book: BookRecord) => {
    if (downloadingPath) return
    setDownloadingPath(book.path)
    try {
      const blob = await exportProjectZip(book.path)
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = `${safeDownloadName(book.name || t('home.unnamedBook'))}.zip`
      document.body.appendChild(anchor)
      anchor.click()
      anchor.remove()
      URL.revokeObjectURL(url)
      toast.success(t('home.exportSuccess', { name: book.name || t('home.unnamedBook') }))
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e)
      console.error('[home] export project failed', { path: book.path, error: e })
      toast.error(t('home.exportFailed'), { description: message })
    } finally {
      setDownloadingPath(null)
    }
  }

  const handleProjectZipSelected = async (file: File | undefined) => {
    if (!file || importingProject) return
    setImportingProject(true)
    try {
      const result = await importProjectZip(file)
      onSwitch(result.workspace)
      onBooksChange()
      toast.success(t('home.importProjectSuccess'))
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e)
      console.error('[home] import project failed', { file: file.name, error: e })
      toast.error(t('home.importProjectFailed'), { description: message })
    } finally {
      setImportingProject(false)
      if (projectImportInputRef.current) projectImportInputRef.current.value = ''
    }
  }

  const currentBook = orderedBooks.find((book) => book.path === workspace)

  return (
    <div className="punkdom-sidebar flex h-full min-w-0 flex-col text-[var(--punkdom-text)]">
      <div className="punkdom-topbar flex h-10 shrink-0 items-center gap-2 border-b px-4 text-xs">
        <LibraryBig className="h-3.5 w-3.5 text-[var(--punkdom-text-muted)]" />
        <span className="font-medium text-[var(--punkdom-text)]">{t('home.title')}</span>
        <span className="text-[11px] text-[var(--punkdom-text-faint)]">{t('home.bookCount', { count: books.length })}</span>
        {onClose && (
          <button
            type="button"
            onClick={onClose}
            className={`${iconButtonCls} ml-auto rounded p-1`}
            aria-label={t('home.close')}
            title={t('home.close')}
          >
            <X className="h-3.5 w-3.5" />
          </button>
        )}
      </div>

      <ScrollArea className="flex-1">
        <div className="mx-auto flex max-w-4xl flex-col gap-5 px-6 py-6">
          {/* 当前书籍 */}
          <section className="border-b border-[var(--punkdom-border)] pb-5">
            <div className="mb-2 flex items-center gap-2 text-[11px] font-medium uppercase text-[var(--punkdom-text-faint)]">
              <BookOpen className="h-3.5 w-3.5" />
              {t('home.currentBook')}
            </div>
            <div className="flex min-w-0 flex-col gap-2 rounded-[var(--punkdom-radius)] border border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] p-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.035)] sm:flex-row sm:items-center sm:justify-between">
              <div className="min-w-0">
                <div className="truncate text-sm font-semibold text-[var(--punkdom-text)]">
                  {currentBook?.name || (workspace ? workspace.split('/').filter(Boolean).pop() : t('home.currentWorkspaceUnset'))}
                </div>
                <div className="mt-1 truncate text-[11px] text-[var(--punkdom-text-faint)]">{workspace || t('home.startHint')}</div>
              </div>
              {currentBook && (
                <div className="flex shrink-0 items-center gap-1.5 rounded border border-[var(--punkdom-border)] bg-[var(--punkdom-surface-2)] px-2 py-1 text-[11px] text-[var(--punkdom-text-muted)]">
                  <BookOpen className="h-3 w-3" />
                  {t('common.current')}
                </div>
              )}
            </div>
          </section>

          {/* 书籍列表 */}
          <section>
            <div className="mb-3 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2 text-[11px] font-medium uppercase text-[var(--punkdom-text-faint)]">
                <Folder className="h-3.5 w-3.5" />
                {t('home.bookshelf')}
              </div>
              <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">
                <Button
                  type="button"
                  size="xs"
                  variant="ghost"
                  className={ghostButtonCls}
                  onClick={() => setShowNovelImport(true)}
                >
                  <FileText className="h-3.5 w-3.5" />
                  {t('home.importNovel')}
                </Button>
                {onOpenCharacterCardImport && (
                  <Button
                    type="button"
                    size="xs"
                    variant="ghost"
                    className={ghostButtonCls}
                    onClick={onOpenCharacterCardImport}
                  >
                    <Upload className="h-3.5 w-3.5" />
                    {t('home.importCard')}
                  </Button>
                )}
                <Button
                  type="button"
                  size="xs"
                  variant="ghost"
                  className={ghostButtonCls}
                  disabled={importingProject}
                  onClick={() => projectImportInputRef.current?.click()}
                >
                  <FileArchive className="h-3.5 w-3.5" />
                  {t('home.importProject')}
                </Button>
                {!showCreateForm && books.length > 0 && (
                  <Button
                    type="button"
                    size="xs"
                    variant="ghost"
                    className={ghostButtonCls}
                    onClick={openCreateForm}
                  >
                    <Plus className="h-3.5 w-3.5" />
                    {t('home.createBook')}
                  </Button>
                )}
                <input
                  ref={projectImportInputRef}
                  type="file"
                  accept=".zip,application/zip,application/x-zip-compressed"
                  className="hidden"
                  onChange={(event) => void handleProjectZipSelected(event.target.files?.[0])}
                />
              </div>
            </div>

            {showCreateForm && (
              <div className="mb-4 space-y-3 rounded-[var(--punkdom-radius)] border border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] p-4 shadow-[inset_0_1px_0_rgba(255,255,255,0.035)]">
                <div className="flex items-center gap-2 text-xs font-medium text-[var(--punkdom-text)]">
                  <Plus className="h-3.5 w-3.5 text-[var(--punkdom-text-muted)]" />
                  {t('home.createBook')}
                </div>
                <Input
                  type="text"
                  value={createTitle}
                  onChange={(e) => setCreateTitle(e.target.value)}
                  placeholder={t('home.bookTitlePlaceholder')}
                  className={inputCls}
                  autoFocus
                />
                <Input
                  type="text"
                  value={createAuthor}
                  onChange={(e) => setCreateAuthor(e.target.value)}
                  placeholder={t('home.authorPlaceholder')}
                  className={inputCls}
                />
                <div className="flex min-w-0 items-center gap-2 rounded-[var(--punkdom-radius)] border border-[var(--punkdom-border)] bg-[var(--punkdom-surface-2)] px-2.5 py-1.5 text-xs text-[var(--punkdom-text-faint)]">
                  <Folder className="h-3.5 w-3.5 shrink-0 text-[var(--punkdom-text-muted)]" />
                  <span className="shrink-0">{t('home.createIn')}</span>
                  <span className="truncate text-[var(--punkdom-text-muted)]">{punkdomDir || t('home.punkdomDirLoading')}</span>
                </div>
                <Textarea
                  autoResize
                  value={createDesc}
                  onChange={(e) => setCreateDesc(e.target.value)}
                  placeholder={t('home.descriptionPlaceholder')}
                  rows={1}
                  className={inputCls + ' min-h-0 resize-none'}
                />
                {createError && <div className="text-xs text-[var(--punkdom-danger)]">{createError}</div>}
                <div className="flex items-center justify-end gap-2">
                  <Button type="button" size="xs" variant="ghost" className={ghostButtonCls} onClick={() => setShowCreateForm(false)}>{t('common.cancel')}</Button>
                  <Button type="button" size="xs" className={primaryButtonCls} disabled={creating || !punkdomDir.trim()} onClick={handleCreate}>
                    {creating ? t('common.creating') : t('common.create')}
                  </Button>
                </div>
              </div>
            )}

            {orderedBooks.length === 0 ? (
              <div className="flex flex-col items-center gap-3 rounded-[var(--punkdom-radius)] border border-dashed border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] px-4 py-8 text-center text-xs text-[var(--punkdom-text-faint)]">
                <div className="text-sm font-medium text-[var(--punkdom-text-muted)]">{t('home.empty')}</div>
                <div className="max-w-md leading-5">{t('home.emptyDescription')}</div>
                {!showCreateForm && (
                  <Button
                    type="button"
                    size="xs"
                    className={primaryButtonCls}
                    onClick={openCreateForm}
                  >
                    <Plus className="h-3.5 w-3.5" />
                    {t('home.createBook')}
                  </Button>
                )}
              </div>
            ) : (
              <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
                <SortableContext items={orderedBooks.map((book) => book.path)} strategy={rectSortingStrategy}>
                  <div className="grid grid-cols-[repeat(auto-fill,minmax(168px,1fr))] gap-3">
                    {orderedBooks.map((book) => {
                      const isCurrent = book.path === workspace
                      const isEditing = editingBookPath === book.path

                      if (isEditing) {
                        return (
                          <SortableBookCard key={book.path} book={book} disabled>
                            {() => (
                              <div className="min-h-[188px] space-y-2 rounded-[var(--punkdom-radius)] border border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] p-3 shadow-[inset_0_1px_0_rgba(255,255,255,0.035)]">
                                {editLoading ? (
                                  <div className="py-2 text-center text-xs text-[var(--punkdom-text-faint)]">{t('common.loading')}</div>
                                ) : (
                                  <>
                                    <Input
                                      type="text"
                                      value={editTitle}
                                      onChange={(e) => setEditTitle(e.target.value)}
                                      placeholder={t('home.bookTitlePlaceholder')}
                                      className={inputCls}
                                      autoFocus
                                    />
                                    <Input
                                      type="text"
                                      value={editAuthor}
                                      onChange={(e) => setEditAuthor(e.target.value)}
                                      placeholder={t('home.authorPlaceholder')}
                                      className={inputCls}
                                    />
                                    <Textarea
                                      autoResize
                                      value={editDesc}
                                      onChange={(e) => setEditDesc(e.target.value)}
                                      placeholder={t('common.description')}
                                      rows={1}
                                      className={inputCls + ' min-h-0 resize-none'}
                                    />
                                    <div className="flex items-center justify-end gap-2">
                                      <TooltipIconButton
                                        label={t('common.cancel')}
                                        className={iconButtonCls}
                                        onClick={() => setEditingBookPath(null)}
                                      >
                                        <X className="h-3.5 w-3.5" />
                                      </TooltipIconButton>
                                      <TooltipIconButton
                                        label={t('common.save')}
                                        className="punkdom-nav-item text-[var(--punkdom-accent-green)] hover:bg-[var(--punkdom-hover)]"
                                        disabled={editSaving}
                                        onClick={handleSaveEdit}
                                      >
                                        <Check className="h-3.5 w-3.5" />
                                      </TooltipIconButton>
                                    </div>
                                  </>
                                )}
                              </div>
                            )}
                          </SortableBookCard>
                        )
                      }

                      return (
                        <SortableBookCard
                          key={book.path}
                          book={book}
                        >
                          {(dragHandleProps) => (
                            <div
                              className={`group relative min-h-[236px] overflow-hidden rounded-[var(--punkdom-radius)] border text-xs transition-colors ${
                                isCurrent
                                  ? 'border-[var(--punkdom-accent)] bg-[var(--punkdom-active)] text-[var(--punkdom-text)] shadow-[inset_0_1px_0_rgba(255,255,255,0.04)]'
                                  : 'border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] text-[var(--punkdom-text-muted)] hover:bg-[var(--punkdom-hover)]'
                              }`}
                            >
                              {isCurrent && (
                                <div className="absolute left-0 top-0 bottom-0 w-[4px] bg-[var(--punkdom-accent)]" />
                              )}
                              <div className="pointer-events-none absolute inset-x-0 bottom-0 h-3 border-t border-[var(--punkdom-border)] bg-[var(--punkdom-surface-2)]" />
                              <button
                                type="button"
                                className="flex h-full min-h-[236px] w-full min-w-0 flex-col px-4 py-4 text-left"
                                onClick={() => handleSwitch(book.path)}
                              >
                                <div className="mb-3 flex items-center justify-between gap-2">
                                  <BookOpen className={`h-4 w-4 shrink-0 ${isCurrent ? 'text-[var(--punkdom-text)]' : 'text-[var(--punkdom-text-muted)]'}`} />
                                </div>
                                <div className="line-clamp-3 text-sm font-semibold leading-5 text-[var(--punkdom-text)]">{book.name || t('home.unnamedBook')}</div>
                                {book.author && <div className="mt-2 truncate text-[11px] text-[var(--punkdom-text-muted)]">{book.author}</div>}
                                <div className="mt-2 line-clamp-4 text-[11px] leading-4 text-[var(--punkdom-text-muted)]">
                                  {book.description || t('home.projectNoDescription')}
                                </div>
                                <div className="mt-auto space-y-1 pr-8 pt-4 text-[10px] leading-4 text-[var(--punkdom-text-faint)]">
                                  <div className="truncate">
                                    {t('home.projectCreatedAt')}: {formatProjectDate(book.created_at)}
                                  </div>
                                  <div className="truncate">
                                    {t('home.projectUpdatedAt')}: {formatProjectDate(book.updated_at)}
                                  </div>
                                </div>
                              </button>
                              <div className="absolute right-2 top-2 z-10 flex shrink-0 items-center gap-0.5">
                                <TooltipIconButton
                                  label={t('home.dragToSort')}
                                  className={`${iconButtonCls} cursor-grab bg-[var(--punkdom-surface)] opacity-100 sm:pointer-events-none sm:opacity-0 sm:group-hover:pointer-events-auto sm:group-hover:opacity-100`}
                                  {...dragHandleProps}
                                >
                                  <GripVertical className="h-3.5 w-3.5" />
                                </TooltipIconButton>
                                <TooltipIconButton
                                  label={t('home.editInfo')}
                                  className={`${iconButtonCls} bg-[var(--punkdom-surface)] opacity-100 sm:pointer-events-none sm:opacity-0 sm:group-hover:pointer-events-auto sm:group-hover:opacity-100`}
                                  onClick={() => startEdit(book)}
                                >
                                  <Pencil className="h-3.5 w-3.5" />
                                </TooltipIconButton>
                                <TooltipIconButton
                                  label={t('home.deleteBook')}
                                  className={`${iconButtonCls} bg-[var(--punkdom-surface)] text-[var(--punkdom-danger)] opacity-100 hover:text-[var(--punkdom-danger)] sm:pointer-events-none sm:opacity-0 sm:group-hover:pointer-events-auto sm:group-hover:opacity-100`}
                                  onClick={() => openDeleteDialog(book)}
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </TooltipIconButton>
                                {isCurrent && (
                                  <span className="rounded border border-[var(--punkdom-border)] bg-[var(--punkdom-surface-2)] px-1.5 py-0.5 text-[10px] text-[var(--punkdom-text-muted)]">
                                    {t('common.current')}
                                  </span>
                                )}
                              </div>
                              <div className="absolute bottom-3 right-2 z-10">
                                <TooltipIconButton
                                  label={t('home.exportProject')}
                                  className={`${iconButtonCls} bg-[var(--punkdom-surface)] opacity-100 sm:pointer-events-none sm:opacity-0 sm:group-hover:pointer-events-auto sm:group-hover:opacity-100`}
                                  disabled={downloadingPath === book.path}
                                  onClick={() => void handleDownloadProject(book)}
                                >
                                  <Download className="h-3.5 w-3.5" />
                                </TooltipIconButton>
                              </div>
                            </div>
                          )}
                        </SortableBookCard>
                      )
                    })}
                  </div>
                </SortableContext>
              </DndContext>
            )}
          </section>

          <section>
            <button
              type="button"
              className="punkdom-nav-item mb-3 flex w-full items-center justify-between gap-3 rounded-[var(--punkdom-radius)] px-1 py-1 text-left hover:bg-[var(--punkdom-hover)]"
              aria-expanded={deletedExpanded}
              onClick={() => setDeletedExpanded((value) => !value)}
            >
              <span className="flex items-center gap-2 text-[11px] font-medium uppercase text-[var(--punkdom-text-faint)]">
                <ChevronDown className={`h-3.5 w-3.5 transition-transform ${deletedExpanded ? '' : '-rotate-90'}`} />
                <Trash2 className="h-3.5 w-3.5" />
                <span>{t('home.deletedProjects')}</span>
              </span>
              <span className="text-[11px] text-[var(--punkdom-text-faint)]">{t('home.deletedCount', { count: deletedBooks.length })}</span>
            </button>
            {deletedExpanded && (
              deletedBooks.length === 0 ? (
                <div className="rounded-[var(--punkdom-radius)] border border-dashed border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] px-4 py-5 text-xs text-[var(--punkdom-text-faint)]">
                  {t('home.deletedEmpty')}
                </div>
              ) : (
                <div className="grid grid-cols-[repeat(auto-fill,minmax(168px,1fr))] gap-3">
                  {deletedBooks.map((book) => (
                    <div
                      key={book.path}
                      className="group relative flex min-h-[210px] flex-col overflow-hidden rounded-[var(--punkdom-radius)] border border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] px-4 py-4 text-xs text-[var(--punkdom-text-muted)] opacity-85"
                    >
                      <div className="mb-3 flex items-center justify-between gap-2">
                        <Trash2 className="h-4 w-4 shrink-0 text-[var(--punkdom-text-faint)]" />
                        <span className="rounded border border-[var(--punkdom-border)] bg-[var(--punkdom-surface-2)] px-1.5 py-0.5 text-[10px] text-[var(--punkdom-text-faint)]">
                          {t('home.deletedBadge')}
                        </span>
                      </div>
                      <div className="line-clamp-2 text-sm font-semibold leading-5 text-[var(--punkdom-text)]">{book.name || t('home.unnamedBook')}</div>
                      {book.author && <div className="mt-2 truncate text-[11px] text-[var(--punkdom-text-muted)]">{book.author}</div>}
                      <div className="mt-2 line-clamp-4 text-[11px] leading-4 text-[var(--punkdom-text-muted)]">
                        {book.description || t('home.projectNoDescription')}
                      </div>
                      <div className="mt-auto space-y-1 pr-8 pt-4 text-[10px] leading-4 text-[var(--punkdom-text-faint)]">
                        <div className="truncate">
                          {t('home.projectCreatedAt')}: {formatProjectDate(book.created_at)}
                        </div>
                        <div className="truncate">
                          {t('home.projectUpdatedAt')}: {formatProjectDate(book.updated_at)}
                        </div>
                      </div>
                      <div className="mt-3 flex items-center justify-end gap-1.5">
                        <TooltipIconButton
                          label={t('home.restoreProject')}
                          className={`${iconButtonCls} bg-[var(--punkdom-surface-2)] text-[var(--punkdom-accent-green)] hover:text-[var(--punkdom-accent-green)]`}
                          disabled={restoringPath === book.path}
                          onClick={() => void handleRestoreDeleted(book)}
                        >
                          <RotateCcw className="h-3.5 w-3.5" />
                        </TooltipIconButton>
                        <TooltipIconButton
                          label={t('home.purgeProject')}
                          className={`${iconButtonCls} bg-[var(--punkdom-surface-2)] text-[var(--punkdom-danger)] hover:text-[var(--punkdom-danger)]`}
                          onClick={() => setPurgeTarget(book)}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </TooltipIconButton>
                      </div>
                    </div>
                  ))}
                </div>
              )
            )}
          </section>

        </div>
      </ScrollArea>
      <NovelImportDialog
        open={showNovelImport}
        punkdomDir={punkdomDir}
        onOpenChange={setShowNovelImport}
        onImported={(result) => {
          onSwitch(result.workspace)
          onBooksChange()
          onClose?.()
        }}
      />
      <AlertDialog open={Boolean(deleteTarget)} onOpenChange={(open) => {
        if (!open && !deleting) setDeleteTarget(null)
      }}>
        <AlertDialogContent className="border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] text-[var(--punkdom-text)]">
          <AlertDialogHeader>
            <AlertDialogTitle>{t('home.deleteBook')}</AlertDialogTitle>
            <AlertDialogDescription className="text-[var(--punkdom-text-muted)]">
              {t('home.deleteBookDescription', { name: deleteTarget?.name || t('home.unnamedBook') })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="truncate rounded border border-[var(--punkdom-border)] bg-[var(--punkdom-surface-2)] px-2.5 py-2 text-xs text-[var(--punkdom-text-faint)]">
            {deleteTarget?.path}
          </div>
          {deleteError && <div className="text-xs text-[var(--punkdom-danger)]">{deleteError}</div>}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleting}>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction
              className="border border-[var(--punkdom-border)] bg-[var(--punkdom-surface-2)] text-[var(--punkdom-text)] hover:bg-[var(--punkdom-hover)]"
              disabled={deleting}
              onClick={(e) => {
                e.preventDefault()
                void handleDelete()
              }}
            >
              {t('home.softDeleteBook')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <AlertDialog open={Boolean(purgeTarget)} onOpenChange={(open) => {
        if (!open && !purging) setPurgeTarget(null)
      }}>
        <AlertDialogContent className="border-[var(--punkdom-border)] bg-[var(--punkdom-surface)] text-[var(--punkdom-text)]">
          <AlertDialogHeader>
            <AlertDialogTitle>{t('home.purgeProject')}</AlertDialogTitle>
            <AlertDialogDescription className="text-[var(--punkdom-text-muted)]">
              {t('home.purgeDescription', { name: purgeTarget?.name || t('home.unnamedBook') })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="truncate rounded border border-[var(--punkdom-border)] bg-[var(--punkdom-surface-2)] px-2.5 py-2 text-xs text-[var(--punkdom-text-faint)]">
            {purgeTarget?.path}
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={purging}>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction
              className="border border-[var(--punkdom-danger-border)] bg-[var(--punkdom-surface-2)] text-[var(--punkdom-danger)] hover:bg-[var(--punkdom-hover)]"
              disabled={purging}
              onClick={(e) => {
                e.preventDefault()
                void handlePurgeDeleted()
              }}
            >
              {t('home.purgeConfirm')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

function safeDownloadName(name: string) {
  return name.trim().replace(/[\\/:*?"<>|]+/g, '_') || 'punkdom-project'
}

function formatProjectDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function SortableBookCard({ book, disabled, children }: {
  book: BookRecord
  disabled?: boolean
  children: (dragHandleProps: ComponentProps<'button'>) => ReactNode
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: book.path, disabled })
  const dragHandleProps: ComponentProps<'button'> = disabled
    ? {}
    : { ...attributes, ...listeners }

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className={isDragging ? 'relative z-10 opacity-80' : undefined}
    >
      {children(dragHandleProps)}
    </div>
  )
}
