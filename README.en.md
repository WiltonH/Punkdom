# Punkdom

Punkdom v0.1 is a local-first AI creative workspace for novels, interactive storytelling, and long-form creative projects.

v0.1 includes Writing Mode, Interactive Mode, premise management, creative Agents, Skills, agent configuration, automations, version management, Projects, text import, character-card import, settings, and local workspace management.

## v0.1.3 Update

- Changed the brand mark to the Lucide `Stone` icon while keeping the minimal IDE-style shell.
- Expanded Projects into a project-first repository: project title edits rename the underlying folder, project cards show description plus created/edited times, and projects can be exported/imported as zip files.
- Project deletion now moves folders into `.punkdom/Trash`; the Deleted section is collapsed by default and supports restore plus permanent deletion while staying in Projects.
- Added Data Backup in Settings for downloading the full `.punkdom` data directory as a zip and restoring from a backup zip with overwrite semantics.
- Added Docker publishing: GitHub Releases now publish GHCR images, Docker Compose assets, and Watchtower auto-update guidance.
- Polished Paper Mode with stronger sidebar active states, warmer user message bubbles, and flatter chat/tool cards.
- Updated primary navigation order and naming: Projects now follows Versions, and the Chinese Narratives label is `叙事模式`.
- Fixed narrative rule switch alignment, Projects navigation after metadata save/restore, and missing created time after moving projects to Trash.

## v0.1.2 Update

- Added a lightweight `Swords` mark beside the Punkdom wordmark in the main shell.
- Renamed primary navigation for the product shell: Workbench, Premise, Narratives, Versions, Repository, Skills, Agents, Automation, Fold, and Settings.
- Fixed stale primary navigation labels after interface language changes; added a top-right `EN/中` language toggle that defaults to Chinese and cycles between Chinese and English.
- Added a body font-size slider in editor settings for chapter reading/editing text only, without changing UI font size.
- Set AI conversation message text to 15px and improved horizontal alignment between the chat canvas and composer.
- Updated the status bar to show the currently effective model and `Powered by Memepop`; the model label follows active model configuration changes.
- Polished Paper Mode top controls, mobile theme/language buttons, and AI conversation visual consistency.

## v0.1.1 Update

- Added Paper Mode with a reading-friendly palette centered on warm off-white, charcoal, and terracotta orange.
- Added a top-right workbench theme button cycling through Light Mode, Paper Mode, and Dark Mode.
- Theme changes are saved to user settings and persist after refresh; Paper Mode is now accepted by backend settings validation.
- Removed the redundant top-bar workbench mode title for a cleaner shell.
- Aligned the center editor and Writing Agent panel header dividers, and added a visible vertical divider between the center and right panels.
- Fixed update checks by using the `WiltonH/Punkdom` Release source and falling back to the GitHub Release page when the GitHub REST API is rate-limited.

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

## Docker

Punkdom GitHub Releases also publish Docker images to GHCR:

```bash
docker run -d --name punkdom \
  -p 8080:8080 \
  -v punkdom-data:/data \
  ghcr.io/wiltonh/punkdom:latest
```

The Docker build uses a single `/data` volume for projects, settings, logs, and local state. The image `/app` directory only contains the Punkdom binary, static web assets, and bundled Skills. Updating the image does not overwrite user content in `/data`.

Use Docker Compose:

```bash
curl -L -o docker-compose.yml https://raw.githubusercontent.com/WiltonH/Punkdom/main/deploy/docker-compose.yml
docker compose up -d
```

Manually update Docker deployments:

```bash
docker compose pull punkdom
docker compose up -d punkdom
```

Enable automatic Docker updates:

```bash
docker compose --profile auto-update up -d
```

This profile starts Watchtower. When a new image is published, Watchtower pulls `ghcr.io/wiltonh/punkdom:latest` and restarts the Punkdom container. In Docker, the in-app update installer does not replace the binary inside the container; the settings page shows the Docker update command instead.

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
