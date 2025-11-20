# Docker 使用说明

## 0. 重点注意

写在最前面。

- 启动后，会产生一个 `images/` 目录，用于存储发布的图片。它会挂载到 Docker 容器里面。
  如果要使用本地图片发布的话，请确保图片拷贝到 `./images/` 目录下，并且让 MCP 在发布的时候，指定文件夹为：`/app/images`，否则一定失败。

## 1. 获取 Docker 镜像

### 1.1 从 Docker Hub 拉取（推荐）

我们提供了预构建的 Docker 镜像，可以直接从 Docker Hub 拉取使用：

```bash
# 拉取最新镜像（AMD64 架构）
docker pull xpzouying/xiaohongshu-mcp

# 拉取 ARM64 架构镜像（Apple Silicon、ARM 服务器）
docker pull xpzouying/xiaohongshu-mcp:arm64
```

Docker Hub 地址：[https://hub.docker.com/r/xpzouying/xiaohongshu-mcp](https://hub.docker.com/r/xpzouying/xiaohongshu-mcp)

> **注意**：ARM64 镜像首次启动时会自动下载 Chromium 浏览器（约 1-2 分钟），后续启动无需重复下载。

### 1.2 自己构建镜像（可选）

在有项目的Dockerfile的目录运行

```bash
docker build -t xpzouying/xiaohongshu-mcp .
```

`xpzouying/xiaohongshu-mcp`为镜像名称和版本。

<img width="2576" height="874" alt="image" src="https://github.com/user-attachments/assets/fe7e87f1-623f-409f-8b54-e11d380fc7b8" />

## 2. 手动 Docker Compose

```bash
# 注意：在 docker-compose.yml 文件的同一个目录，或者手动指定 docker-compose.yml。

# --- 启动 docker 容器 ---
# 启动 docker-compose
docker compose up -d

# 查看日志
docker logs -f xpzouying/xiaohongshu-mcp

# 或者
docker compose logs -f
```

查看日志，下面表示成功启动。

<img width="1012" height="98" alt="image" src="https://github.com/user-attachments/assets/c374f112-a5b5-4cf6-bd9f-080252079b10" />


```bash
# 停止 docker-compose
docker compose stop

# 查看实时日志
docker logs -f xpzouying/xiaohongshu-mcp

# 进入容器
docker exec -it xiaohongshu-mcp bash

# 手动更新容器
docker compose pull && docker compose up -d
```

## 3. 使用 MCP-Inspector 进行连接

**注意 IP 换成你自己的 IP**

<img width="2606" height="1164" alt="image" src="https://github.com/user-attachments/assets/495916ad-0643-491d-ae3c-14cbf431c16f" />

对应的 Docker 日志一切正常。

<img width="1662" height="458" alt="image" src="https://github.com/user-attachments/assets/309c2dab-51c4-4502-a41b-cdd4a3dd57ac" />

## 4. 扫码登录

1. **重要**，一定要先把 App 提前打开，准备扫码登录。
2. 尽快扫码，有可能二维码会过期。

打开 MCP-Inspector 获取二维码和进行扫码。

<img width="2632" height="1468" alt="image" src="https://github.com/user-attachments/assets/543a5427-50e3-4970-b942-5d05d69596f4" />

<img width="2624" height="1222" alt="image" src="https://github.com/user-attachments/assets/4f38ca81-1014-4874-ab4d-baf02b750b55" />

扫码成功后，再次扫码后，就会提示已经完成登录了。

<img width="2614" height="994" alt="image" src="https://github.com/user-attachments/assets/5356914a-3241-4bfd-b6b2-49c1cc5e3394" />


## 5. ARM64 架构使用说明

### 5.1 适用场景

ARM64 镜像适用于：
- Apple Silicon Mac（M1/M2/M3/M4 芯片）
- ARM 架构的服务器

### 5.2 构建 ARM64 镜像

```bash
# 在项目根目录运行
docker build -f Dockerfile.arm64 -t xpzouying/xiaohongshu-mcp:arm64 .
```

### 5.3 首次登录（保存 Cookies）

由于 Docker 容器运行在无头模式，无法手动输入短信验证码，需要**先在本地登录一次**保存 cookies：

```bash
# 在项目根目录运行
go run cmd/login/main.go
```

这会：
1. 打开浏览器窗口显示二维码
2. 使用小红书 APP 扫码
3. 手动输入短信验证码
4. 登录成功后自动保存 cookies 到 `cookies.json`

### 5.4 使用 Cookies 运行容器

登录成功后，使用以下命令运行 Docker 容器（会自动使用已保存的登录状态）：

```bash
docker run -d \
  --name xiaohongshu-mcp \
  -p 18060:18060 \
  -e TZ=Asia/Shanghai \
  -e COOKIES_PATH=/app/data/cookies.json \
  -v $(pwd)/cookies.json:/app/data/cookies.json \
  -v $(pwd)/images:/app/images \
  xpzouying/xiaohongshu-mcp:arm64
```

**参数说明**：
- `-e TZ=Asia/Shanghai`：设置时区为中国时区
- `-e COOKIES_PATH=/app/data/cookies.json`：指定 cookies 文件路径
- `-v $(pwd)/cookies.json:/app/data/cookies.json`：挂载本地的 cookies 文件
- `-v $(pwd)/images:/app/images`：挂载图片目录

### 5.5 验证运行状态

```bash
# 查看日志
docker logs -f xiaohongshu-mcp

# 测试连接
curl -X POST http://localhost:18060/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
```

### 5.6 注意事项

1. **首次启动较慢**：ARM64 镜像首次启动时会自动下载 Chromium 浏览器（约 120MB），需要 1-2 分钟。后续启动会使用缓存，速度很快。

2. **Cookies 过期处理**：如果 cookies 过期导致登录失效，重新执行步骤 5.3 本地登录即可。

3. **时区问题**：容器已设置为 `Asia/Shanghai` 时区，但日志输出仍为 UTC 时间（这是标准做法），不影响功能使用。

