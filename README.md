# 慈善项目管理系统（go05-charity-project）

一个使用 **纯 Go 标准库** 从 0 到 1 构建的慈善项目管理系统。公益机构发布募捐项目，捐赠人在线捐款并查看记录，系统以公开账本公示每一笔资金流入与支出。系统内置前端页面与文件级 JSON 数据持久化，可通过 Docker 独立运行。

---

## 一、项目简介

公益募捐常见的痛点是：项目散落在表格和朋友圈，捐了多少事后对不上、钱花到哪里看不见、退款与管理费口径不统一。本系统把三条主链路产品化：

- **项目展示**：机构创建募捐项目（目标金额、截止日、受益对象），管理员审核后进入广场，公众可浏览进度与进展报告。
- **捐款记录**：捐赠人选择金额与支付方式完成捐款；系统生成编号、收据与个人流水，支持匿名、匹配捐赠与有条件退款。
- **资金流向公示**：每一笔确认收入、支出、退款、匹配款写入 **只追加账本**；广场详情页公开余额、管理费占比与透明度评分。

角色：

- **捐赠人（donor）**：注册登录、浏览广场、关注项目、捐款、查看个人捐款与收据、留言。
- **公益机构（org）**：维护机构资料、发布项目、确认线下捐款、登记支出并公示、发布进展、申请匹配款。
- **管理员（admin）**：用户与机构治理、项目审核、强制关闭、调账、全局统计与透明度抽查。

系统使用 Go 1.22 + 标准库（`net/http`、`encoding/json`、`embed`、`sync` 等），**零第三方依赖**，可完全离线构建与运行。金额一律以 **分（int64）** 存储，避免浮点误差。

---

## 二、功能特性

### 2.1 用户与权限

| 角色 | 能力 |
|------|------|
| 管理员 | 用户列表/创建/冻结/解冻，审核机构资质，审核/强制关闭项目，调账，查看全局统计与审计 |
| 公益机构 | 维护机构档案、创建项目、提交审核、确认线下捐款、登记支出并公示、发布进展、申请匹配捐赠 |
| 捐赠人 | 注册登录、浏览项目、关注、捐款、查看捐款记录与收据、留言、申请退款 |
| 未登录访客 | 仅登录/注册页、健康检查、公开项目列表与资金流向（只读） |

- 首次启动自动创建种子管理员（默认 `admin / admin123`）。
- 捐赠人可自助注册；机构账号由管理员创建（避免随意发项目）。
- 会话：Bearer Token，带过期时间；登出、改密、冻结即失效。
- 口令：盐值 + 多轮迭代 SHA-256（演示级，生产应替换为 bcrypt/argon2）。
- 账号状态：`active` / `frozen`（冻结后不可捐款、发项目） / `banned`（无法登录）。

### 2.2 机构档案

每个 `org` 账号绑定一份 **机构档案**：

- 机构全称、统一社会信用代码（演示字段）、联系人、简介。
- 资质状态：`unverified` / `verified` / `rejected`。
- **未核验机构不能发布项目**（可保存草稿）。管理员核验后才允许提交审核。
- 机构累计筹款、累计支出、进行中项目数、透明度均分写入档案，供广场展示。

### 2.3 项目生命周期（项目展示）

状态流转：

```
[ 草稿 draft ]
      │ 提交审核 submit
      ▼
[ 待审 pending_review ] ──管理员驳回──► [ 草稿 draft ]
      │ 管理员通过
      ▼
[ 已发布 published ] ──机构关闭募捐 / 截止日惰性关闭──► [ 募捐结束 closed ]
      │                                                      │
      │ 目标达成（Raised >= Goal）仍可超募（若允许）              │ 支出公示完毕
      ▼                                                      ▼
[ 已发布 published ]                                   [ 已结项 completed ]

任意非终态可由机构或管理员 ──取消──► [ 已取消 cancelled ]
管理员可 ──强制关闭──► [ 已关闭 closed ]
```

项目字段要点：

- **分类 `Category`**：教育、医疗、救灾、扶贫、环保、动物保护、社区、其他。
- **目标金额 `GoalCents`**：以分为单位，默认最低 10000 分（100 元），最高 1 亿元。
- **募捐窗口**：`StartAt` ~ `EndAt`；未到开始时间不可捐；超过截止且未开 `AllowLateDonation` 则拒绝新捐。
- **是否允许超募 `AllowOverGoal`**：关闭后，已达目标即拒绝新捐（匹配款除外，由管理员操作）。
- **最低/最高单笔捐款**、是否允许匿名、是否允许线下转账待确认。
- **受益对象 `Beneficiary`**：文字说明（村庄、学校、病患群体等）。
- **封面 URL**（演示用外链/占位，不存二进制）。
- **所属机构 `OrgID`**：机构只能维护本机构项目；管理员可跨机构。

列表规则：

- 广场默认只展示 `published`（含已达目标仍在募的项目）。
- 访客可按分类、关键词、是否达目标过滤。
- 机构/管理员可按状态查看草稿、待审、已关闭。
- 惰性推进：读取项目时若已过 `EndAt` 且状态仍为 published，自动置为 `closed`（不再接受新捐，仍可登记支出）。

### 2.4 捐款记录（核心）

> **定义**：一笔捐款对应一条 `Donation`。在线模拟支付创建后立即 `confirmed`；线下转账先 `pending`，机构确认后入账。只有 **confirmed** 计入项目已筹金额与公开账本。

1. 不能给自己机构的项目捐款。
2. 项目必须为 `published` 才接受新捐；`closed` 仅处理已有 pending 线下单（机构仍可确认或拒绝）。
3. **金额**：`MinDonationCents`～`MaxDonationCents`（项目可收紧，但不能突破平台上下限）。
4. **日限额**：同一捐赠人自然日累计不超过 `DailyCapCents`（默认 5 万元）。
5. **匿名**：`Anonymous=true` 时公开列表只显示“爱心人士”，个人中心仍可见全额。
6. **留言**：0–140 字，写入捐款记录，可在项目页展示（匿名则不展示用户名）。
7. **支付方式**：`wechat` / `alipay` / `bank` / `offline`。前三者为模拟在线，创建即确认；`offline` 为 pending。
8. **入账**：确认时在同一把锁内：累加 `Project.RaisedCents`、写入收入账本、累加捐赠人 `TotalDonatedCents`、若达到目标记录 `GoalReachedAt`。
9. **收据**：确认金额 ≥ `ReceiptThresholdCents`（默认 100 元）自动生成可核验收据编号。
10. **退款**：`confirmed` 且距确认不超过 `RefundWindowDays`（默认 7 天）；项目未 `completed`；剩余可用余额 ≥ 退款额。退款写入账本 `refund`，扣减已筹金额。pending 单由机构拒绝或捐赠人撤销，不进账本。
11. **匹配捐赠**：管理员或机构（需有匹配预算说明）可追加 `matching` 收入，计入已筹但不占用某位捐赠人的日限额。

捐款状态：`pending` / `confirmed` / `rejected` / `refunded` / `cancelled`。

### 2.5 资金流向公示（核心）

> **定义**：每个项目维护一份 **只追加（append-only）账本 `LedgerEntry`**。公开页按时间正序展示。余额由账本重算，禁止直接改余额字段作为唯一真实来源。

分录类型：

| 类型 | 方向 | 来源 |
|------|------|------|
| `income` | 入 | 确认捐款 |
| `matching` | 入 | 匹配捐赠 |
| `expense` | 出 | 已公示支出 |
| `refund` | 出 | 已确认退款 |
| `adjust` | 入或出 | 仅管理员调账（必须填写原因） |

可用余额：

```
available = sum(income) + sum(matching) + adjust_in
          - sum(expense) - sum(refund) - adjust_out
```

`Project.RaisedCents` 应等于收入 + 匹配 − 退款（不含支出）。  
`Project.SpentCents` 应等于已公示支出。  
结项前若 `available != 0`，机构必须继续公示支出或由管理员调账（例如结转至同类项目），否则不可 `completed`。

支出规则：

1. 仅机构（本项目）或管理员可创建支出草稿。
2. 字段：标题、类目、金额、发生日、受益方、票据号、说明。
3. **类目 `ExpenseCategory`**：物资、劳务、物流、医疗、教育、管理费、其他。
4. **管理费封顶**：累计 `admin_fee` ≤ `RaisedCents * MaxAdminFeeRate`（默认 8%）。超限拒绝公示。
5. 公示时校验 `amount ≤ available`，通过后追加 `expense` 分录并累加 `SpentCents`。
6. 草稿可撤回；已公示支出不可删除，只能由管理员 `adjust` 冲正。
7. 项目详情页展示：已筹 / 目标进度条、已支出、可用余额、管理费占比、透明度评分。

透明度评分（0–100）：

- 有目标与受益说明 +20
- 至少 1 条进展报告 +20
- 已筹 > 0 且支出公示覆盖已筹的 80% 以上（或已结项且余额为 0）+30
- 机构已核验 +15
- 近 30 天有账本更新或尚在募捐期 +15

### 2.6 进展、关注、留言、收据

- **进展报告**：机构在执行期发布图文进展（标题+正文）；广场详情只读。结项前建议至少 1 条。
- **关注**：捐赠人可关注 published 项目；关闭后仍可查看但提示不可再捐。
- **留言**：对项目提问或鼓励，机构可回复；不当内容管理员可删除。
- **通知**：审核结果、捐款确认、退款结果、支出公示、进展更新（关注者）。
- **审计日志**：发布、审核、入账、退款、公示支出、调账、强制关闭写 `AuditLog`。
- **收据核验**：`GET /api/receipts/{code}/verify` 公开接口，返回金额、项目名、是否匿名。

### 2.7 统计

管理员看板：用户数、机构数、项目数（按状态）、本月筹款、本月支出、平均透明度、待审项目。  
机构看板：本机构在募项目、待确认线下捐款、待公示支出、可用余额。  
捐赠人主页：累计捐赠、进行中关注、最近收据。

---

## 三、业务对象与持久化

数据全部保存在 `data/store.json`（路径由 `APP_DATA_PATH` 配置）。内存结构变更后通过钩子 **原子写盘**（临时文件 + rename），进程重启后恢复。

主要集合：

- Users、Organizations、Projects、Donations、LedgerEntries、Expenses
- ProgressReports、Follows、Comments、Receipts
- Notifications、AuditLogs

并发：`MemoryStore` 使用 `sync.RWMutex`；跨实体操作（确认捐款+入账+更新已筹、公示支出+扣余额、退款+冲减）在同一把锁内完成，避免半更新。

---

## 四、API 一览

前缀 `/api`。健康检查、登录注册、公开项目列表/详情、公开账本、收据核验可不登录；写操作需 `Authorization: Bearer <token>`。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 健康检查 |
| POST | `/api/auth/register` | 捐赠人注册 |
| POST | `/api/auth/login` | 登录 |
| POST | `/api/auth/logout` | 登出 |
| GET | `/api/auth/me` | 当前用户 |
| PUT | `/api/me/profile` | 改资料 |
| PUT | `/api/me/password` | 改密 |
| GET | `/api/categories` | 项目分类与支出类目枚举 |
| GET | `/api/orgs` | 机构列表 |
| POST | `/api/orgs` | 创建机构档案（机构本人或管理员） |
| GET | `/api/orgs/{id}` | 机构详情 |
| POST | `/api/orgs/{id}/verify` | 核验机构（管理员） |
| GET/POST | `/api/projects` | 列表 / 创建草稿 |
| GET/PUT | `/api/projects/{id}` | 详情 / 更新草稿 |
| POST | `/api/projects/{id}/submit` | 提交审核 |
| POST | `/api/projects/{id}/approve` | 审核通过（管理员） |
| POST | `/api/projects/{id}/reject` | 审核驳回 |
| POST | `/api/projects/{id}/close` | 关闭募捐 |
| POST | `/api/projects/{id}/complete` | 结项（余额须为 0） |
| POST | `/api/projects/{id}/cancel` | 取消 |
| POST | `/api/projects/{id}/donate` | 捐款 |
| GET | `/api/projects/{id}/donations` | 捐款记录（公开列表脱敏） |
| GET | `/api/projects/{id}/ledger` | 资金流向公示 |
| GET | `/api/projects/{id}/summary` | 余额与透明度摘要 |
| POST | `/api/projects/{id}/follow` | 关注 |
| GET/POST | `/api/projects/{id}/progress` | 进展列表 / 发布 |
| GET/POST | `/api/projects/{id}/comments` | 留言 |
| POST | `/api/comments/{id}/reply` | 回复留言 |
| GET | `/api/me/donations` | 我的捐款 |
| POST | `/api/donations/{id}/confirm` | 确认线下捐款 |
| POST | `/api/donations/{id}/reject` | 拒绝线下捐款 |
| POST | `/api/donations/{id}/refund` | 申请退款 |
| POST | `/api/projects/{id}/expenses` | 登记支出草稿 |
| POST | `/api/expenses/{id}/publish` | 公示支出（写入账本） |
| POST | `/api/projects/{id}/match` | 匹配捐赠 |
| POST | `/api/projects/{id}/adjust` | 管理员调账 |
| GET | `/api/receipts/{code}/verify` | 收据核验（公开） |
| GET | `/api/me/notifications` | 通知 |
| GET | `/api/stats` | 统计（管理员） |
| GET | `/api/users` | 用户列表（管理员） |
| POST | `/api/users/{id}/freeze` | 冻结 |

前端页面（`embed`）：`/login`、`/app`（项目广场）、`/projects/{id}`、`/me`、`/org`、`/admin`。

---

## 五、技术架构

```
main.go
  ├─ config          环境变量
  ├─ money           分/元换算、管理费封顶、透明度评分
  ├─ receipt         收据编号
  ├─ store           MemoryStore + FileStore 原子持久化
  ├─ auth            口令哈希 / 会话
  ├─ policy          限额、窗口、管理费率等策略常量
  ├─ service         项目 / 捐款 / 支出账本 / 通知
  ├─ handler         HTTP JSON + 页面
  ├─ middleware      鉴权、角色、CORS、日志、Recover
  └─ web/assets      内置 HTML/CSS/JS
```

约束：

- `go.mod` 声明 `go 1.22`，不使用 Go 1.23+ API。
- 零第三方模块，质检镜像 `golang:1.22` 内 `go mod download` 与 `go build ./...` 可离线完成。
- 运行镜像见项目根 `Dockerfile` + `docker-compose.yml`（端口 8080，挂载 `./data`）。

---

## 六、默认账号与演示数据

| 账号 | 口令 | 角色 |
|------|------|------|
| admin | admin123 | 管理员 |
| org | org123 | 公益机构（阳光公益） |
| alice | alice123 | 捐赠人 |
| bob | bob123 | 捐赠人 |

`APP_SEED_DEMO=true` 时写入上述账号、一个已核验机构，以及一个已发布的助学募捐项目（含一笔演示捐款与一笔已公示支出，便于查看资金流向）。

环境变量：`APP_ADDR`、`APP_DATA_PATH`、`APP_SESSION_TTL`、`APP_ADMIN_USERNAME`、`APP_ADMIN_PASSWORD`、`APP_SEED_ADMIN`、`APP_SEED_DEMO`、`APP_DAILY_CAP_CENTS`、`APP_MAX_ADMIN_FEE_RATE`。

---

## 七、本地与 Docker 运行

见 `BENZHI_README.md`。简述：

```bash
go run .
# 或
bash ./go-run.sh
```

浏览器打开 http://localhost:8080/login 。

---

## 八、核心规则速查

| 场景 | 结果 |
|------|------|
| 机构未核验 | 不可提交审核，只能草稿 |
| 项目非 published | 不可新捐 |
| 给自己机构项目捐款 | 禁止 |
| 超过日限额 / 单笔上下限 | 400 校验失败 |
| 已达目标且不允许超募 | 409 冲突 |
| 在线支付 | 立即 confirmed 并入账 |
| 线下转账 | pending，机构确认后入账 |
| 7 日内退款且余额足够 | 账本 refund，扣已筹 |
| 管理费累计超过已筹 8% | 支出不可公示 |
| 支出金额 > 可用余额 | 409 冲突 |
| 结项时可用余额 ≠ 0 | 拒绝结项 |
| 已公示支出 | 不可删除，只能管理员调账 |
