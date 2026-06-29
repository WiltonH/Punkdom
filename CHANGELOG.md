# Changelog

## [Unreleased]

- Added multi-arch Docker image build and push to GitHub Container Registry (`ghcr.io/wiltonh/punkdom`) on tagged releases.
- Added `Dockerfile`, `.dockerignore`, and `deploy/docker-compose.yml` with Watchtower auto-update support.

## [v0.1.3] - 2026-06-25

- Changed the Punkdom wordmark icon from `Swords` to the Lucide `Stone` icon.
- Flattened chat message bubbles and tool cards by removing their drop shadows.
- Renamed the underlying book workspace folder when editing a book title from Book Management, with conflict protection for existing target folders.
- Renamed Book Management to Project Repository and moved it after Automation in the primary menu.
- Shortened the English Project Repository label to Projects.
- Changed the Projects icon to `Layers`, moved Versions between Automation and Projects, and renamed the Chinese Narratives label to `叙事模式`.
- Improved Paper theme primary sidebar active-state contrast with a warmer selected background and no animated active indicator.
- Recolored Paper theme user chat bubbles to use a deeper warm neutral instead of the terracotta accent.
- Aligned narrative injection rule switches with a fixed right column and centered switch thumb geometry.
- Renamed Project Repository labels from book/bookshelf language to project/text language in Chinese and English.
- Removed container image publishing and deployment assets from the v0.1.3 release scope.
- Added project zip export/download and project zip import for moving Punkdom projects between installations.
- Changed project deletion to move folders into `.punkdom/Trash`, with a Deleted section that supports restore and permanent deletion.
- Made the Deleted projects section collapsed by default on every page load.
- Added project card metadata for description, created time, and last edited time.
- Kept the Projects view open after saving metadata for the current project.
- Kept the Projects view open after restoring a deleted project.
- Preserved project metadata, including creation time, when moving projects to Trash.
- Added full Punkdom data backup export and overwrite restore controls in Settings.
- Removed the inactive Story Route Map footer hint/minimap bar while preserving node selection and branch creation.
- Replaced the dark selected-node highlight in the Story Route Map with a thicker accent outline.
- Added runtime application version reporting through `/api/status` and made the status bar update from that backend version.

## [v0.1.2] - 2026-06-23

- Added a `Swords` Lucide mark beside the Punkdom wordmark in the workbench top bar.
- Renamed the primary sidebar entries and fixed language switching so cached menu labels refresh immediately.
- Added a top-right language toggle that defaults to Chinese and cycles between Chinese and English.
- Added a content font-size slider to editor settings for the chapter body without changing UI or AI conversation font size.
- Aligned the AI composer with the conversation content area.
- Set AI conversation message text to 15px.
- Updated the status bar runtime credit to show `Powered by Memepop`.
- Made the status bar model label follow the currently effective model configuration instead of a fixed provider name.
- Unified the AI conversation canvas and composer gutter so the input box stays aligned with chat content.

## [v0.1.1] - 2026-06-22

- Added Paper theme with Claude-style warm off-white, charcoal, terracotta orange, soft blue, and olive green palette.
- Added a top-right workbench theme button cycling through Light, Paper, and Dark modes.
- Persisted theme changes through user settings while preserving the existing settings panel theme selector.
- Fixed Paper theme persistence by allowing `paper` in backend settings validation.
- Removed the redundant workbench mode title text from the top bar.
- Aligned the Writing Agent panel header divider with the center editor tab bar.
- Added a visible vertical divider between the center editor panel and right Agent panel.
- Fixed update checks by pointing the default release repository to `WiltonH/Punkdom` and falling back to the GitHub Release redirect page when the GitHub REST API is rate-limited.

## [v0.1.0] - 2026-06-22

- Imported Nova commit `8e4528c48875e0fdaacc5548b1b2ad197259fff6` as the Punkdom v0.1 functional baseline.
- Renamed the application, CLI, Go module, package metadata, configuration directory, environment variables, frontend storage keys, events, and user-facing product text from Nova to Punkdom.
- Changed new workspace metadata from `.nova/` to `.punkdom/`.
- Preserved Apache-2.0 licensing and added NOTICE attribution for the upstream project.
