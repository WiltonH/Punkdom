# Punkdom

Punkdom v0.1 is a local-first AI creative workspace for novels, interactive storytelling, and long-form creative projects.

v0.1 includes Writing Mode, Interactive Mode, structured lore, creative Agents, Skills, automations, version management, novel import, character-card import, settings, and local workspace management.

## v0.1.1 Update

- Added Paper Mode with a reading-friendly palette centered on warm off-white, charcoal, and terracotta orange.
- Added a top-right workbench theme button cycling through Light Mode, Paper Mode, and Dark Mode.
- Theme changes are saved to user settings and persist after refresh; Paper Mode is now accepted by backend settings validation.
- Removed the redundant top-bar workbench mode title for a cleaner shell.

## Quick Start

Requirements: Go 1.26+, Node.js 20+, and pnpm.

```bash
corepack enable
./bootstrap.sh
```

Default URLs:

- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`

Start with a workspace:

```bash
./punkdom --workspace /path/to/your-workspace
```

## Configuration

Punkdom uses an OpenAI-compatible API:

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://api.deepseek.com"
export OPENAI_MODEL="deepseek-v4-pro"
```

Common Punkdom environment variables:

```bash
export PUNKDOM_WORKSPACE="/path/to/your-workspace"
export PUNKDOM_DIR="./.punkdom"
export PUNKDOM_SKILLS_DIR="./skills"
export PUNKDOM_WEB_DIR="./web"
export PUNKDOM_BACKEND_PORT="8080"
export PUNKDOM_FRONTEND_PORT="5173"
```

Use `punkdom_dir` in `config.toml` for the startup-level Punkdom data directory. User-level and workspace-level settings ignore this locator field.

## Themes

Punkdom includes three main interface themes:

- Light Mode
- Paper Mode: warm off-white `#faf9f5`, light greige `#e8e6dc`, charcoal `#141413`, warm gray `#b0aea5`, terracotta orange `#d97757`, soft blue `#6a9bcc`, and olive green `#788c5d`
- Dark Mode

## Workspace

New workspaces use this default structure:

```text
my-project/
├── CREATOR.md
├── ideas.md
├── chapters/
├── setting/
│   ├── progress.md
│   ├── character-states.md
│   └── chapter-groups/
├── drafts/
└── .punkdom/
    ├── lore/
    └── sessions/
```

`.punkdom/` stores local creative state such as lore, sessions, automations, run traces, and workspace configuration.

## Development

Start both frontend and backend:

```bash
./bootstrap.sh
```

Start only one side:

```bash
./bootstrap.sh fe
./bootstrap.sh be
```

Build:

```bash
./build.sh
```

## License

Punkdom is licensed under Apache-2.0. See [LICENSE](./LICENSE) for details.
