# 藏叶 (Cangye) 🍃

> 虚拟文件夹——将分散在各处的文件聚合到统一界面中浏览和预览。

藏叶是一款 Windows 桌面端工具，让你创建**虚拟文件夹**，将本地文件和网络资源聚合到统一界面中进行浏览和预览。支持多源路径扫描、标签管理、前缀映射等功能。

## 技术栈

- **后端**: Go + HTTP + SQLite
- **前端**: Vue 3 + TypeScript
- **构建**: 前端 dist 嵌入 Go 二进制文件，双击 exe 即开即用

## 快速开始

### 下载

在 [Releases](https://github.com/yehuoshun/cangye/releases) 页面下载最新的 `.exe` 文件，双击启动即可。会自动在 `http://127.0.0.1:27138` 打开界面。

### 从源码构建

```bash
# 1. 安装依赖
cd frontend && npm install

# 2. 构建前端
npm run build:skip-lint

# 3. 构建后端
cd ../backend
go mod tidy
go build -o cangye.exe .

# 4. 运行
./cangye.exe
```

## 功能

### 虚拟文件夹
- 创建虚拟文件夹，聚合多源文件
- 支持子文件夹层级
- 文件夹图标自定义

### 路径扫描
- 添加本地路径，扫描文件
- 自动识别文件类型（文本、图片、视频、音频等）
- 缓存扫描结果

### 虚拟文件
- 手动添加文件（本地或网络路径）
- 前缀映射系统（如 `115:` 表示 115 网盘文件，`tg:` 表示 Telegram 文件）

### 浏览与预览
- 网格/列表两种视图切换
- 文本/代码文件内联预览
- 图片缩略图显示
- 视频/音频 HTML5 播放
- 右键菜单：预览、外部程序打开、复制路径

### 标签系统
- 动态搜索已有标签
- 输入时无匹配则回车创建
- 为文件和文件夹添加标签分类

### 前缀映射
内置默认前缀：`115:`(本地)、`tg:`(网页)、`notion:`(网页)
可在设置中编辑类型和映射路径。

### 设置
- 布局切换（侧边栏 / 顶部导航 / 双栏对开）
- 明暗主题切换
- 前缀映射管理
- 标签管理

## 项目结构

```
cangye/
├── backend/
│   ├── main.go           # 入口：HTTP 服务 + embed + 自动开浏览器
│   ├── go.mod
│   ├── db/               # SQLite 初始化 + 迁移
│   ├── file/             # 文件模块（集合、路径、虚拟文件、扫描、浏览、预览）
│   ├── rss/              # RSS 模块（TODO）
│   ├── checkin/          # 签到模块（TODO）
│   ├── settings/         # 设置模块（前缀、标签）
│   ├── common/           # 共享工具
│   └── web/dist/         # 前端构建产物（由 embed.FS 嵌入）
├── frontend/
│   ├── src/
│   │   ├── views/        # 页面视图
│   │   ├── components/   # 通用组件
│   │   ├── composables/  # Vue 组合式函数
│   │   ├── api/          # API 客户端
│   │   └── assets/       # 样式
│   ├── index.html
│   └── package.json
├── .github/workflows/    # GitHub Actions
├── README.md
└── .gitignore
```

## API

完整的 RESTful API，见 `backend/file/file.go`、`backend/settings/settings.go` 等源文件。

主要端点：

| 模块 | 端点 | 方法 | 说明 |
|------|------|------|------|
| 文件夹 | `/api/collections` | GET/POST | 列表/创建 |
| 文件夹 | `/api/collections/:id` | GET/PUT/DELETE | 详情/编辑/删除 |
| 路径 | `/api/collections/:id/paths` | GET/POST | 路径列表/添加 |
| 路径 | `/api/paths/:id/scan` | POST | 扫描路径 |
| 虚拟文件 | `/api/collections/:id/vfiles` | GET/POST | 列表/创建 |
| 浏览 | `/api/collections/:id/browse` | GET | 聚合文件列表 |
| 预览 | `/api/preview/content?path=` | GET | 文本内容 |
| 预览 | `/api/preview/thumbnail?path=` | GET | 图片缩略图 |
| 标签 | `/api/tags/search?q=` | GET | 搜索标签 |
| 总览 | `/api/overview/stats` | GET | 统计数据 |
| 设置 | `/api/settings/:key` | GET/PUT | 获取/更新设置 |

## License

MIT
