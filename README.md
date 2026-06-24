# Punkdom

Punkdom v0.1 是一个本地优先的 AI 创作工作台，用于小说、互动叙事和长篇创意项目的结构化生产。

v0.1 提供写作模式、互动模式、设定集、创作 Agent、技能点、智能体配置、自动化、版本管理、项目仓库、文本导入、角色卡导入、设置和本地 workspace 管理。

## v0.1.3 更新

- 品牌标识调整为 `Stone` 图标，并继续保持极简 IDE 风格。
- 项目仓库完成项目化改造：支持项目重命名时同步重命名底层文件夹、项目 zip 导出/导入、项目简介与创建/编辑时间展示。
- 删除项目会移动到 `.punkdom/Trash`，并在默认折叠的“已删除”栏目中支持恢复和彻底删除；恢复后仍停留在项目仓库。
- 设置页新增数据备份：可一键下载完整 `.punkdom` 数据 zip，也可上传备份 zip 覆盖还原全部本地数据。
- 新增 Docker 发布方案：GitHub Release 同步发布 GHCR 镜像，提供 Docker Compose 模板和 Watchtower 自动更新指引。
- 优化纸张模式：侧边栏高亮、用户对话气泡、对话卡片和工具卡片更扁平、更贴近主色调。
- 调整一级菜单顺序和命名：项目仓库放在版本管理之后，叙事编排改为叙事模式，英文项目仓库为 `Projects`。
- 修复叙事模式规则开关对齐、项目元数据保存/恢复后跳转、删除项目后创建时间丢失等问题。

## v0.1.2 更新

- 主界面左上角新增 `Swords` 图标标识，并保持轻量 IDE 风格。
- 一级菜单完成产品化命名：工作台、设定集、叙事编排、版本管理、书籍管理、技能点、智能体、自动化、收起、设置。
- 修复界面语言切换后一级菜单仍显示旧语言的问题；新增右上角 `EN/中` 语言切换按钮，默认中文，可循环切换中英文。
- 编辑器设置新增正文字体大小滑块，仅调整中间正文阅读/编辑字号，不影响 UI 字体。
- AI 对话消息字号固定为 15px，并优化输入框与对话内容区域的横向对齐。
- 状态栏右侧显示当前有效大模型，并追加 `Powered by Memepop`；模型名会随当前模型配置变化。
- 优化纸张模式下的顶部控制、移动端主题/语言按钮和 AI 对话区视觉一致性。

## v0.1.1 更新

- 新增纸张模式，采用暖米色、炭黑和陶土橙为核心的阅读友好配色。
- 主界面右上角新增主题切换按钮，按浅色模式、纸张模式、深色模式顺序循环。
- 主题切换会写入用户设置，刷新后保持当前选择；纸张模式现在可被后端配置校验正确保存。
- 移除顶部栏冗余的“写作工作台”模式标题，让工作台顶部更简洁。
- 对齐中间编辑区与创作 Agent 面板顶部栏分隔线，并在中间面板与右面板之间加入可见竖向分隔线。
- 修复应用更新检查：默认指向 `WiltonH/Punkdom` Release，并在 GitHub REST API 限流时自动使用 Release 页面兜底。

## Quick Start

需要 Go 1.26+、Node.js 20+ 和 pnpm。

```bash
corepack enable
./bootstrap.sh
```

默认地址：

- 前端：`http://localhost:5173`
- 后端：`http://localhost:8080`

指定作品目录启动：

```bash
./punkdom --workspace /path/to/your-workspace
```

## Configuration

Punkdom 使用 OpenAI 兼容接口，可通过环境变量快速配置：

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://api.deepseek.com"
export OPENAI_MODEL="deepseek-v4-pro"
```

常用 Punkdom 环境变量：

```bash
export PUNKDOM_WORKSPACE="/path/to/your-workspace"
export PUNKDOM_DIR="./.punkdom"
export PUNKDOM_SKILLS_DIR="./skills"
export PUNKDOM_WEB_DIR="./web"
export PUNKDOM_BACKEND_PORT="8080"
export PUNKDOM_FRONTEND_PORT="5173"
```

配置文件使用 `punkdom_dir` 指定 Punkdom 数据目录。用户级和工作区级配置会忽略该启动级定位参数。

## Docker

Punkdom 的 GitHub Release 会同步发布 Docker 镜像到 GHCR：

```bash
docker run -d --name punkdom \
  -p 8080:8080 \
  -v punkdom-data:/data \
  ghcr.io/wiltonh/punkdom:latest
```

Docker 版默认使用单数据卷 `/data` 保存作品、配置、日志和本地状态；镜像内的 `/app` 只保存程序、前端静态资源和内置 Skills。更新镜像不会覆盖 `/data` 中的用户内容。

使用 Docker Compose：

```bash
curl -L -o docker-compose.yml https://raw.githubusercontent.com/WiltonH/Punkdom/main/deploy/docker-compose.yml
docker compose up -d
```

手动更新 Docker 版：

```bash
docker compose pull punkdom
docker compose up -d punkdom
```

自动更新 Docker 版：

```bash
docker compose --profile auto-update up -d
```

该 profile 会启动 Watchtower。新镜像发布后，Watchtower 会自动拉取 `ghcr.io/wiltonh/punkdom:latest` 并重启 Punkdom 容器。Docker 环境下应用内“安装更新”不会替换容器里的二进制文件，设置页会显示 Docker 更新命令。

## Themes

Punkdom 提供三套主界面主题：

- 浅色模式
- 纸张模式：暖白 / 米白 `#faf9f5`、浅灰米色 `#e8e6dc`、炭黑 `#141413`、暖灰 `#b0aea5`、陶土橙 `#d97757`、柔和蓝 `#6a9bcc`、橄榄绿 `#788c5d`
- 深色模式

## Workspace

新建工作区默认结构：

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

`.punkdom/` 保存 Punkdom 的本地创作状态，例如资料库、会话、自动化、运行记录和工作区配置。

## Development

启动前后端：

```bash
./bootstrap.sh
```

仅启动前端或后端：

```bash
./bootstrap.sh fe
./bootstrap.sh be
```

构建：

```bash
./build.sh
```

## License

Punkdom is licensed under Apache-2.0. See [LICENSE](./LICENSE) for details.
