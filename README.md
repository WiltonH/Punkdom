# Punkdom

Punkdom v0.1 是一个本地优先的 AI 创作工作台，用于小说、互动叙事和长篇创意项目的结构化生产。

v0.1 提供写作模式、互动模式、结构化资料库、创作 Agent、Skills、自动化、版本管理、小说导入、角色卡导入、设置和本地 workspace 管理。

## v0.1.1 更新

- 新增纸张模式，采用暖米色、炭黑和陶土橙为核心的阅读友好配色。
- 主界面右上角新增主题切换按钮，按浅色模式、纸张模式、深色模式顺序循环。
- 主题切换会写入用户设置，刷新后保持当前选择；纸张模式现在可被后端配置校验正确保存。
- 移除顶部栏冗余的“写作工作台”模式标题，让工作台顶部更简洁。

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
