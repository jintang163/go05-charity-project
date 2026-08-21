# BENZHI_README

## 项目是做什么的

**go05-charity-project** 是一个慈善项目管理系统，使用 **纯 Go 标准库**（零第三方依赖）实现。

- **公益机构（org）** 发布募捐项目（目标金额、截止日、受益对象），确认线下捐款，登记并公示支出，发布进展报告。
- **捐赠人（donor）** 浏览广场、关注项目、在线或线下捐款、查看个人捐款记录与收据、查看公开资金流向。
- **管理员（admin）** 核验机构资质、审核项目、冻结用户、调账、强制关闭、查看全局统计。
- **核心业务规则**：金额以分存储；确认捐款写入只追加账本；管理费不超过已筹 8%；退款需在窗口内且余额足够；结项前可用余额必须为 0；未核验机构不能发布项目。
- 内置前端页面（HTML/CSS/JS，`embed` 打包）与文件级 JSON 数据持久化（`data/store.json`，原子落盘，重启自动恢复）。
- 单一 Go 二进制，可通过 Docker 独立运行，适合离线受限环境交付。

技术栈：Go 1.22、`net/http`（Go 1.22 `ServeMux` 方法路由）、`encoding/json`、`embed`、`sync`、`crypto/rand`、`crypto/sha256`。

---

## 构建命令

```bash
# 本地构建（需本地安装 Go 1.22+）
go build ./...

# 质检镜像构建（基于 benzhi.Dockerfile，linux/amd64）
bash ./build_benzhi_docker.sh go05-charity-project
```

## 运行命令

```bash
# 方式一：本地直接运行
go run .

# 方式二：Docker Compose 一键起服务（后台常驻，:8080，种子管理员 admin/admin123）
bash ./go-run.sh
#   等价于：docker compose up -d --build
#   访问：http://localhost:8080/healthz
#   日志：docker compose logs -f
#   停止：docker compose down
```

## 测试命令

```bash
# 方式一：本地测试
go test ./...

# 方式二：质检环境测试（先构建 benzhi 镜像，再在容器内跑 go test）
bash ./go-test.sh go05-charity-project "go test ./..."
```

---

## 目录与质检文件说明

| 文件 | 是否可改 | 说明 |
|------|----------|------|
| `benzhi.Dockerfile` | ❌ 勿改 | 质检镜像（`golang:1.22`，`go mod download` + `go build ./...`） |
| `build_benzhi_docker.sh` | ❌ 勿改 | 质检镜像构建脚本 |
| `go-test.sh` | ✅ 可改 | 质检测试脚本（构建镜像后在容器内执行测试命令） |
| `go-run.sh` | ❌ 勿改 | 运行脚本（`docker compose up -d --build`） |
| `Dockerfile` | ✅ | 运行镜像（单阶段 `golang:1.22`，避免 alpine 拉取超时） |
| `docker-compose.yml` | ✅ | 服务编排（:8080，挂载 `./data` 持久化） |

> 约束：`go.mod` 声明 `go 1.22`，不使用 Go 1.23+ API（如 `crypto/pbkdf2`）；零第三方依赖，确保 `go mod download` 无需联网即可在 `golang:1.22` 镜像内离线构建与测试。

## 默认账号

- 管理员：`admin / admin123`（首次启动自动创建，可通过环境变量 `APP_ADMIN_USERNAME` / `APP_ADMIN_PASSWORD` 覆盖）
- 演示机构：`org / org123`
- 演示捐赠人：`alice / alice123`、`bob / bob123`（`APP_SEED_DEMO=true` 时若库中尚无捐赠人则写入）

## 快速验证

```bash
# 1. 登录获取 token
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"alice123"}'

# 2. 浏览募捐项目
curl -s http://localhost:8080/api/projects \
  -H "Authorization: Bearer <token>"

# 3. 对某项目捐款（金额单位：分）
curl -s -X POST http://localhost:8080/api/projects/<id>/donate \
  -H "Authorization: Bearer <token>" -H 'Content-Type: application/json' \
  -d '{"amount_cents":10000,"channel":"wechat","message":"加油","anonymous":false}'

# 4. 查看资金流向公示
curl -s http://localhost:8080/api/projects/<id>/ledger \
  -H "Authorization: Bearer <token>"

# 5. 浏览器访问 http://localhost:8080/login 查看前端页面
```
