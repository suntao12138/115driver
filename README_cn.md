# 115driver

> **中文 | [English](README.md) 🌐**

适用于 [115 云盘](https://115.com) 的全功能 Go 库、CLI 工具和 MCP 服务端。提供对 115.com API 的完整驱动支持，包括登录、文件操作、上传/下载、离线下载等。

[![Go Report Card](https://goreportcard.com/badge/github.com/SheltonZhu/115driver)](https://goreportcard.com/report/github.com/SheltonZhu/115driver)
[![Release](https://img.shields.io/github/release/SheltonZhu/115driver)](https://github.com/SheltonZhu/115driver/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/SheltonZhu/115driver/v4.svg)](https://pkg.go.dev/github.com/SheltonZhu/115driver)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/:License-MIT-orange.svg)](https://raw.githubusercontent.com/SheltonZhu/115driver/main/LICENSE)

## 目录

- [功能特性](#功能特性)
- [安装](#安装)
- [快速开始](#快速开始)
- [CLI 命令行](#cli-命令行)
- [MCP 服务端](#mcp-服务端)
- [API 参考](#api-参考)
- [故障排除](#故障排除)
- [项目结构](#项目结构)
- [参与贡献](#参与贡献)
- [许可证](#许可证)

> ✨ **Fork 增强版** — 本仓库（Suntoa12138/115driver）在原版基础上增加了以下上游未合并的改进：
> - **配置文件认证** — MCP 服务端通过 `--profile` 参数自动从 `~/.115driver/config.toml` 读取 Cookie
> - **默认离线下载目录** — CLI 和 MCP 均支持从配置文件读取默认保存目录
> - **API 兼容性修复** — 适配 115 接口 `imei_info` 和下载 URL 字段格式变更

## 功能特性

**认证登录** — 支持 Cookie 登录、二维码登录和身份验证。

**文件操作** — 列表、重命名、移动、复制、删除、下载、上传（支持 SHA1 秒传和阿里云 OSS 分片上传）、搜索过滤、文件信息/统计。

**离线下载** — 添加 HTTP、ED2K 和磁力链接下载任务；查看、删除、清理任务。

**分享功能** — 创建分享链接，通过分享码下载文件。

**回收站** — 列出、恢复和永久删除回收站项目。

**CLI 命令行** — 全功能命令行界面，支持彩色表格输出、JSON 机器输出模式、Shell 自动补全和多配置文件。

**MCP 服务端** — [Model Context Protocol](https://modelcontextprotocol.io/) 服务端，用于 AI 应用集成（Claude Desktop、Cursor 等）。

## 安装

> ⚠️ **这是一个增强版 fork。** 包含以下上游未合并的改进：
> - 配置文件认证（MCP 通过 `--profile` 从 `~/.115driver/config.toml` 自动读取 Cookie）
> - 默认离线目录（CLI 和 MCP 支持 `default_offline_save_dir` 配置项）
> - API 兼容修复（适配 `imei_info` 和下载 URL 字段格式变更）
>
> 下方所有安装命令均指向本 fork 的增强版本。

### 作为 Go 库使用

```bash
go get github.com/SheltonZhu/115driver

# 在你的 go.mod 中添加 replace 指令指向本 fork：
# go mod edit -replace github.com/SheltonZhu/115driver=github.com/suntao12138/115driver@latest
```

### 安装 CLI

```bash
go install github.com/suntao12138/115driver/cmd/115driver@latest
```

### 安装 MCP 服务端

**方式一：go install**

```bash
go install github.com/suntao12138/115driver/mcp@latest
```

**方式二：从源码编译**

```bash
git clone https://github.com/suntao12138/115driver.git
cd 115driver
go build -o 115driver-mcp-server ./mcp/
```

## 快速开始

### 基础用法

```go
package main

import (
    "github.com/SheltonZhu/115driver/pkg/driver"
    "log"
)

func main() {
    // 方式一：从 Cookie 字符串导入凭据
    cr, err := driver.CredentialFromCookie("your_cookie_string")
    if err != nil {
        log.Fatalf("创建凭据失败: %v", err)
    }

    // 方式二：手动创建凭据
    // cr := &driver.Credential{
    //     UID:  "your_uid",
    //     CID:  "your_cid",
    //     SEID: "your_seid",
    //     KID:  "your_kid",
    // }

    // 创建客户端并导入凭据
    client := driver.Default().ImportCredential(cr)

    // 检查登录状态
    if err := client.LoginCheck(); err != nil {
        log.Fatalf("登录失败: %v", err)
    }

    log.Println("登录成功！")
}
```

### 常用操作

以下示例假设你已经有一个已认证的 `client`（参见基础用法）。

```go
// 使用 pickcode 下载文件
downloadInfo, err := client.Download("pickcode")
if err != nil { /* 处理错误 */ }
fileReader, _ := downloadInfo.Get()
defer fileReader.Close()
// 将 fileReader 写入本地文件...
```

```go
// 上传文件（自动选择秒传或 OSS 分片上传）
file, _ := os.Open("/path/to/local/file.zip")
defer file.Close()
fileInfo, _ := file.Stat()
uploadID, err := client.RapidUploadOrByOSS(
    "0",            // 父目录 ID（"0" 为根目录）
    fileInfo.Name(),
    fileInfo.Size(),
    file,
)
```

```go
// 列出根目录文件
files, err := client.List("0")
for _, f := range files {
    log.Printf("文件: %s, 大小: %d, 类型: %s", f.Name, f.Size, f.Type)
}
```

```go
// 搜索文件
results, err := client.Search(&driver.SearchOption{
    SearchValue: "文档",
    Limit:       100,
})
for _, r := range results.Files {
    log.Printf("文件: %s, 大小: %d", r.Name, r.Size)
}
```

```go
// 添加离线下载任务
taskIDs, err := client.AddOfflineTaskURIs(
    []string{"https://example.com/file.zip"},
    "0", // "0" 为根目录
)
```

## CLI 命令行

115driver 包含一个命令行工具，用于通过终端操作 115 云盘。既支持人类友好的彩色表格输出，也支持 AI 智能体消费的 `--json` 模式。

### 认证登录

```bash
# 二维码登录（交互式）
115driver login

# Cookie 登录
115driver login --cookie "UID=xxx;CID=xxx;SEID=xxx;KID=xxx"

# 验证身份
115driver whoami

# 账户和存储信息
115driver info
```

凭据存储在 `~/.115driver/config.toml`，支持多配置文件。

### 认证优先级

1. `--cookie` 命令行参数
2. `DRIVER115_COOKIE` 环境变量
3. 配置文件（`~/.115driver/config.toml`）

附加环境变量：`DRIVER115_CONFIG`（配置文件路径）、`DRIVER115_PROFILE`（配置文件名）。

### 命令

```bash
# 列出文件
115driver ls /path/to/dir
115driver ls -l /path/to/dir          # 详细视图

# 文件信息
115driver stat /path/to/file

# 账户和存储信息
115driver info

# 创建目录
115driver mkdir /new/dir
115driver mkdir -p /deep/nested/dir   # 递归创建父目录

# 移动 / 复制 / 重命名 / 删除
115driver mv /source/file /dest/dir
115driver cp /source/file /dest/dir
115driver rename /path/to/file new_name
115driver rm /path/to/file

# 上传 & 下载
115driver upload /local/file /remote/dir
115driver download /remote/file /local/dir

# 搜索
115driver search keyword
115driver search keyword -t video     # 按类型过滤
115driver search keyword --sort size  # 排序结果

# 离线下载（HTTP/ED2K/磁力链）
115driver offline add <url>
115driver offline add <url> -d /save/dir
115driver offline list
115driver offline rm <hash>
```

### JSON 输出

所有命令支持 `--json` 参数，输出机器可读的 JSON 格式：

```bash
115driver --json ls /path/to/dir
115driver --json stat /path/to/file
115driver --json info
```

### Shell 自动补全

```bash
# Bash
echo 'source <(115driver completion bash)' >> ~/.bashrc

# Zsh
echo 'source <(115driver completion zsh)' >> ~/.zshrc

# Fish
115driver completion fish > ~/.config/fish/completions/115driver.fish
```

## MCP 服务端

115driver 包含一个 MCP（Model Context Protocol）服务端，用于集成到 AI 应用中（Claude Desktop、Cursor 等）。

### 用法

```bash
# go install 安装后：
mcp --profile main                    # 从配置文件读取 Cookie（推荐）
mcp --cookie="UID=xxx;CID=xxx;SEID=xxx;KID=xxx"  # 或直接传 Cookie

# 从源码编译后：
./115driver-mcp-server --profile main
```

### 可用工具

| 分类 | 工具 |
|------|------|
| **账户** | `getAccountInfo` |
| **目录** | `listDirectory` |
| **文件** | `stat`, `mkdir`, `delete`, `rename`, `move`, `copy`, `upload_from_url`, `upload_from_local`, `download_file`, `get_download_info` |
| **搜索** | `search` |
| **离线** | `listOfflineTasks`, `addOfflineTaskURIs`, `deleteOfflineTasks`, `clearOfflineTasks` |
| **分享** | `getShareSnap` |
| **回收站** | `listRecycleBin`, `revertRecycleBin`, `cleanRecycleBin` |

### 配置文件认证（Fork 增强特性）

`--cookie` 参数可以省略，前提是 `~/.115driver/config.toml` 中存在有效的 Cookie：

```bash
# 使用默认配置文件
./115driver-mcp-server --profile main

# 或通过环境变量指定
DRIVER115_PROFILE=main ./115driver-mcp-server
```

配置文件路径优先级：`--config` 参数 > `DRIVER115_CONFIG` 环境变量 > `~/.115driver/config.toml`。
配置文件名优先级：`--profile` 参数 > `DRIVER115_PROFILE` 环境变量 > `default_profile` 配置项 > `"main"`。

`addOfflineTaskURIs` 工具的 `save_dir_id` 参数现在是可选的——如果省略，会自动使用配置文件中的 `default_offline_save_dir`（如果已设置）。先运行 `115driver login` 生成配置文件。

### 配置到 Claude Desktop

在 `claude_desktop_config.json` 中添加：

```json
{
  "mcpServers": {
    "115driver": {
      "command": "mcp",
      "args": ["--profile", "main"]
    }
  }
}
```

> **提示：** 如果使用本 fork 编译的二进制文件，将 `command` 改为 `115driver-mcp-server` 或对应的二进制路径。使用 `--profile main` 会自动从 `~/.115driver/config.toml` 读取 Cookie，无需在配置中暴露密钥。请先运行 `115driver login` 生成配置文件。

## API 参考

详细的 API 文档请访问 [pkg.go.dev](https://pkg.go.dev/github.com/SheltonZhu/115driver)。

## 故障排除

### 登录问题

如果遇到登录问题：
1. 确认 Cookie 有效且未过期
2. 检查所有必填字段（UID、CID、SEID、KID）是否完整
3. 先通过 Web 界面登录获取新 Cookie

### 上传/下载问题

如果上传或下载失败：
1. 检查文件路径是否正确
2. 确认网络连接正常
3. 确保有足够的存储空间
4. 查看返回的错误信息以获取详细信息

### 频率限制

115 API 可能有频率限制。如果遇到频率限制错误：
1. 在操作之间添加延迟
2. 实现带指数退避的重试逻辑
3. 必要时考虑使用代理

## 项目结构

```
115driver/                    # Go 1.23+
├── cmd/
│   └── 115driver/            # CLI 入口（go install 二进制）
├── cli/                      # CLI 实现
│   ├── cmd/                  # Cobra 命令
│   └── internal/             # 内部包（auth, output, resolver）
├── internal/                 # 共享应用级辅助函数
├── pkg/
│   ├── driver/               # 核心驱动（client, login, file, upload, download, search, share, offline）
│   └── crypto/               # 加密工具（ECDH, AES, RSA）
└── mcp/                      # MCP 服务端（stdin/stdout JSON-RPC 2.0）
    ├── main.go               # 入口
    └── server/tools/         # 工具实现（account, dir, file, search, offline, share, recycle）
```

## Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=sheltonzhu/115driver&type=date&legend=top-left)](https://www.star-history.com/#sheltonzhu/115driver&type=date&legend=top-left)

## 参与贡献

欢迎贡献！请随时提交 Pull Request。

## 贡献者

<!-- readme: contributors -start -->
<table>
<tr>
    <td align="center">
        <a href="https://github.com/SheltonZhu">
            <img src="https://avatars.githubusercontent.com/u/26734784?v=4" width="100;" alt="SheltonZhu"/>
            <br />
            <sub><b>SheltonZhu</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/xhofe">
            <img src="https://avatars.githubusercontent.com/u/36558727?v=4" width="100;" alt="xhofe"/>
            <br />
            <sub><b>xhofe</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/Ovear">
            <img src="https://avatars.githubusercontent.com/u/1362137?v=4" width="100;" alt="Ovear"/>
            <br />
            <sub><b>Ovear</b></sub>
        </a>
    </td>
    <td align="center">
        <a href="https://github.com/power721">
            <img src="https://avatars.githubusercontent.com/u/2384040?v=4" width="100;" alt="power721"/>
            <br />
            <sub><b>power721</b></sub>
        </a>
    </td></tr>
    <td align="center">
        <a href="https://github.com/suntao12138">
            <img src="https://avatars.githubusercontent.com/u/168153569?v=4" width="100;" alt="suntao12138"/>
            <br />
            <sub><b>suntao12138</b></sub>
        </a>
    </td></tr>
</table>
<!-- readme: contributors -end -->

## 许可证

[MIT](LICENSE)
