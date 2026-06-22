# Changelog

## [v0.1.1] - 2026-06-22

- Added Paper theme with Claude-style warm off-white, charcoal, terracotta orange, soft blue, and olive green palette.
- Added a top-right workbench theme button cycling through Light, Paper, and Dark modes.
- Persisted theme changes through user settings while preserving the existing settings panel theme selector.
- Fixed Paper theme persistence by allowing `paper` in backend settings validation.
- Removed the redundant workbench mode title text from the top bar.

## [v0.1.0] - 2026-06-22

- Imported Nova commit `8e4528c48875e0fdaacc5548b1b2ad197259fff6` as the Punkdom v0.1 functional baseline.
- Renamed the application, CLI, Go module, package metadata, configuration directory, environment variables, frontend storage keys, events, and user-facing product text from Nova to Punkdom.
- Changed new workspace metadata from `.nova/` to `.punkdom/`.
- Preserved Apache-2.0 licensing and added NOTICE attribution for the upstream project.
