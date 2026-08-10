# 知识点记录

本文档持续记录 `myim` 开发过程中适合 face 复习的技术知识。与相关场景和参与者有关的描述统一使用 `face`、`facer`。

## Protobuf 与 JSON 双协议响应

### 核心知识

- Protobuf 是二进制协议，体积较小、解析速度快，并通过 `.proto` 文件约束字段编号和类型。
- Protobuf JSON 是 protobuf 官方定义的 JSON 映射，不等同于随意使用 `encoding/json` 序列化结构体。
- HTTP 客户端可以通过 `Accept` 声明期望的响应格式，通过 `Content-Type` 声明请求体格式。
- `application/x-protobuf` 返回 protobuf 二进制；普通请求返回 protobuf JSON。
- protobuf 字段编号一旦发布不能随意复用。字段改名但保留编号通常不影响二进制兼容性，但会影响 JSON 字段名和生成代码 API。

### 当前约定

所有 protobuf 响应的前两个字段固定为：

```proto
int32 error_code = 1;
string error_msg = 2;
```

成功时 `error_code` 为 `0`；失败时写入明确错误码和错误信息。统一响应器负责选择 protobuf 或 JSON 编码。

### Facer 可能追问

**为什么不能只看请求的 `Content-Type` 决定响应？**

`Content-Type` 描述请求体，`Accept` 才描述客户端期望的响应类型。为兼容简单客户端，可以在 `Accept` 缺失时参考请求 `Content-Type`。

**修改 protobuf 字段名是否兼容？**

二进制协议主要依赖字段编号和类型，因此保留编号和类型时通常兼容；JSON 客户端依赖字段名，所以字段改名会产生 JSON API 变化。

## HTTP 服务初始化

### Gin Engine 与 HTTP Server 的区别

Gin engine 主要负责中间件、路由匹配和 handler 调用，本质上实现了标准库的 `http.Handler` 接口。真正监听端口、接受 TCP 连接、处理超时和关闭连接的是标准库 `http.Server`。

```text
客户端请求
  -> http.Server 接受连接
  -> Gin middleware
  -> Gin router 匹配路由
  -> HTTP handler 绑定协议
  -> service 执行业务
  -> DAO 访问 PostgreSQL
  -> 统一 Response 编码
```

因此当前由根 `app.App` 完成依赖组装和服务生命周期：

```go
application, err := app.New()
err = application.Run()
```

`app.New` 按配置、DAO、Service、Router、HTTP Server 的顺序组装依赖；`App.Run` 启动 HTTP Server并等待退出信号，`App.Stop` 只执行一次 HTTP 优雅关闭和 PostgreSQL 关闭。`app/router.New` 只构建 Gin middleware、handler 和路由。

### `Run`、`Start` 与 `Stop`

方法名应该表达阻塞语义。当前 `App.Run` 会一直阻塞，直到 HTTP Server 退出或进程收到停止信号，因此 `Run` 比 `Start` 更准确。工程中常见的 `Start` 通常只负责拉起 goroutine 或监听器，然后立即返回；如果把当前方法改名为 `Start`，调用方容易误判其行为。

关闭逻辑单独放入 `Stop(ctx)` 有三个好处：外部测试或未来的生命周期管理器可以主动停止应用；HTTP 和数据库的关闭顺序集中在一个地方；未来增加 WebSocket、gRPC 或后台任务时有明确的资源回收入口。`Stop` 使用 `sync.Once` 保证资源只关闭一次，并用 `errors.Join` 保留多个资源各自的关闭错误。

VS Code 的 `.vscode/launch.json` 使用 Go debug 类型，以 `app/cmd` 为程序入口、项目根目录为工作目录。调试器仍然执行完整的 `app.New -> App.Run` 生命周期，所以启动前 PostgreSQL 必须可用。

### 为什么增加根 App

把所有初始化直接写在 `main` 中并不是错误，小型程序通常会这样开始。但当应用继续增加 WebSocket、gRPC、后台任务和更多数据库资源时，main 会逐渐混入依赖创建、启动顺序和关闭顺序。

根 `App` 是 composition root：

- 集中决定具体使用哪些 Config、DAO、Service 和 Server 实现。
- 业务层只依赖接口，不负责创建数据库或网络服务。
- 统一管理 HTTP、未来 gRPC、WebSocket 和后台任务的生命周期。
- main 只处理顶层错误，保持进程入口稳定。

参考项目的 Application 还管理 Redis、Kafka、遥测和多个 Server。当前 `myim` 只实现实际需要的 PostgreSQL 与 HTTP，避免提前复制复杂基础设施。

### `app.go` 应该放在 `app` 还是 `internal`

Go 的 `internal` 是编译器支持的可见性边界：`internal` 目录中的包只能被其父目录树内的代码导入。它不是“所有内部实现都必须放进去”的固定分层名称。

当前 `myim` 的 router、service、dao、model 都在 `app` 下，因此把 composition root 放在 `app/app.go` 最一致。只把 App 单独移动到 `internal`，并不能提高微服务可迁移性，反而会让一个服务的业务边界横跨两套目录。

如果以后拆分微服务，可以整体迁移成：

```text
apps/user-service/cmd/main.go
apps/user-service/internal/app/app.go
apps/user-service/internal/router
apps/user-service/internal/service
apps/user-service/internal/dao

apps/message-service/cmd/main.go
apps/message-service/internal/app/app.go
apps/message-service/internal/router
apps/message-service/internal/service
apps/message-service/internal/dao
```

真正决定微服务边界的是独立部署、独立配置、独立数据所有权和独立生命周期，而不是 App 文件是否提前放在 `internal`。

### 通用 HTTP 错误放在哪里

请求体无法解析、协议格式错误和隐藏服务端细节都属于 HTTP 传输层规则，统一放在 `internal/httpx`：

- `ErrInvalidRequest`：请求体格式错误或无法按声明协议解析。
- `ErrInternalServer`：服务端内部失败时返回给客户端的安全文案。

业务错误仍然留在各自 service，例如用户名重复或登录凭证错误。这样 router 不需要借用 user service 的错误表达 HTTP binding 失败。

### `gin.New` 与 `gin.Default`

- `gin.New()` 返回不带中间件的 engine，安装了哪些能力一目了然。
- `gin.Default()` 等价于创建 engine 后自动加入 Logger 和 Recovery。
- 当前使用 `gin.New()`，再显式加入 `gin.Logger()`、`gin.Recovery()`。
- Logger 记录请求方法、路径、状态码和耗时；Recovery 捕获 handler panic，避免整个进程退出。

### Trusted Proxies

Gin 可以从 `X-Forwarded-For` 等 Header 获取客户端 IP。如果默认信任所有代理，外部客户端可能伪造来源 IP。当前没有反向代理配置，因此使用：

```go
router.SetTrustedProxies(nil)
```

以后部署在 Nginx 或网关后面时，应只填写真实代理网段。

### 健康检查

`GET /health` 返回 HTTP 204，表示进程和 HTTP 路由已经可用。健康检查不执行数据库查询，因此属于轻量的存活检查。后续可以另加 readiness 接口，检查 PostgreSQL 等依赖是否就绪。

### HTTP 超时

- `ReadHeaderTimeout`：限制读取请求 Header 的时间，降低 Slowloris 攻击风险。
- `ReadTimeout`：限制读取完整请求的时间。
- `WriteTimeout`：限制写响应的时间。
- `IdleTimeout`：限制 keep-alive 空闲连接存活时间。
- `MaxHeaderBytes`：限制请求 Header 大小。

这些超时不能完全照搬到所有服务。文件上传、流式响应和 WebSocket 握手可能需要不同设置；WebSocket 完成连接升级后由连接层自行管理读写 deadline 和心跳。

### `ListenAndServe` 与优雅退出

`ListenAndServe` 会阻塞当前 goroutine。为了同时监听系统退出信号，当前把它放进 goroutine，并通过 channel 把启动或运行错误传回主流程。

收到 Ctrl+C、SIGINT 或 SIGTERM 后调用 `server.Shutdown`：

1. 停止接受新连接。
2. 等待正在处理的 HTTP 请求结束。
3. 超过 5 秒仍未结束则返回超时错误。
4. `App.Run` 随后关闭 PostgreSQL 连接池。

直接结束进程可能中断正在注册或登录的请求。优雅退出让请求和资源按顺序收尾，是服务发布和容器停止时的重要能力。

### HTTP 测试

`httptest.NewRecorder` 和 `httptest.NewRequest` 可以直接调用 Gin engine，不需要占用真实端口。适合验证路由是否注册、状态码、Header 和响应体。真实端口监听、代理和操作系统信号则属于更高层的集成测试范围。

## 密码存储

### 核心知识

- 密码不能明文存储，也不应使用可逆加密。
- bcrypt 会自动生成 salt，并把计算成本写入哈希结果。
- 登录时使用 `bcrypt.CompareHashAndPassword`，不需要手动取 salt。
- bcrypt 输入上限是 72 字节，因此注册校验必须限制密码字节长度。
- 用户不存在和密码错误应返回相同提示，降低账号枚举风险。

### Facer 可能追问

**为什么不能直接使用 SHA-256 存密码？**

SHA-256 速度太快，攻击者可以高并发暴力计算。bcrypt 是专门的慢哈希算法，可以通过成本参数增加破解代价。

## JWT

### 核心知识

- JWT 由 Header、Payload、Signature 三部分组成。
- HS256 使用 HMAC-SHA256 和共享密钥签名。
- Payload 只是 Base64URL 编码，不是加密，不能放密码等敏感数据。
- 验证时必须检查签名、过期时间和必要 claims。
- JWT 签发后难以主动失效。需要强制下线、多设备管理或撤销能力时，应增加 session、token version 或 Redis 黑名单。

## PostgreSQL 与 DAO

### PostgreSQL 对象层级

PostgreSQL 可以按下面的层级理解：

```text
PostgreSQL 服务实例
  -> Database
    -> Schema（默认 public）
      -> Table / Index / Sequence
```

应用连接串先指定 Database，例如当前的 `dbname=test`。连接成功后，SQL 默认在 `public` schema 中查找 `users` 表。

### 启动时是否每次都要建表

不需要每次启动都真正创建一遍表。当前 `myim` 在 DAO 初始化时执行：

```sql
CREATE TABLE IF NOT EXISTS users (...);
```

每次启动确实都会把这条 DDL 发给 PostgreSQL，但 `IF NOT EXISTS` 会先检查系统目录：

- 表不存在：创建表。
- 表已经存在：跳过创建，不清空数据。
- 新字段使用 `ADD COLUMN IF NOT EXISTS`，已有字段同样会跳过。

这种方式适合当前快速开发阶段，优点是启动简单，缺点是无法严谨记录数据库从哪个版本迁移到哪个版本，也不适合复杂的数据转换、回滚和多实例并发发布。

生产环境通常使用版本化 migration：

```text
000001_create_users.up.sql
000002_add_user_profile.up.sql
000003_add_email_index.up.sql
```

migration 工具会维护版本表，每个版本只执行一次。常见工具包括 `golang-migrate`、`goose` 和 Atlas。应用正常启动时只建立连接并检查依赖，不负责反复执行整套 schema。

### `sql.Open` 不等于已经连接

Go 的 `sql.Open` 主要创建一个连接池句柄，通常不会立即访问 PostgreSQL。真正执行 SQL 时才会建立连接。需要在启动阶段尽早发现错误时，可以调用：

```go
if err := db.PingContext(ctx); err != nil {
    return err
}
```

`*sql.DB` 本身是并发安全的连接池，应在应用生命周期内复用，而不是每个请求都重新创建和关闭。

### 与 MongoDB 初始化的差异

- PostgreSQL 是强 schema 数据库。表、列、类型和约束必须预先定义，插入不存在的列会失败。
- MongoDB collection 通常可以在第一次写入时自动出现，同一 collection 的 document 字段也可以不同。
- MongoDB 仍然需要显式管理索引、唯一约束和数据校验；“不建表”不等于完全不需要初始化。
- PostgreSQL 的 schema 变化通常通过 migration 管理；MongoDB 的字段演进更多由应用兼容旧 document，并配合数据修复脚本完成。

### 最基础 CRUD

创建用户：

```sql
INSERT INTO users (user_id, user_name, password, register_time, updated_time)
VALUES ($1, $2, $3, $4, $5);
```

查询用户：

```sql
SELECT user_id, user_name, nickname, status
FROM users
WHERE user_id = $1;
```

更新用户：

```sql
UPDATE users
SET nickname = $1,
    updated_time = $2
WHERE user_id = $3;
```

删除用户：

```sql
DELETE FROM users
WHERE user_id = $1;
```

社交服务通常不会立即物理删除用户，而是先把 `status` 更新为注销状态。物理删除可能破坏消息、关系和内容记录的引用，需要配合数据保留策略处理。

### 常见查询结果处理

- `QueryRowContext`：预期返回一行，调用 `Scan` 读取；没有结果时返回 `sql.ErrNoRows`。
- `QueryContext`：返回多行，需要 `defer rows.Close()`，循环 `rows.Next()` 并在最后检查 `rows.Err()`。
- `ExecContext`：用于 `INSERT`、`UPDATE`、`DELETE` 和 DDL，可以读取 `RowsAffected()`。
- 多步写操作要求同时成功或同时失败时，应使用 transaction。

### 参数化 SQL

使用 `$1`、`$2` 等占位符把 SQL 结构与用户数据分离，可以避免 SQL 注入，也有利于数据库复用执行计划。

SQL 原始字符串中的换行和缩进只会被视为空白，不改变语句含义。按 `SELECT`、字段、`FROM`、`WHERE`、`VALUES` 分行能提高审查和维护效率。

### 唯一索引

- 用户名使用 `LOWER(user_name)` 唯一索引，实现大小写不敏感的唯一性。
- 手机号和邮箱当前允许为空，因此使用部分唯一索引，只约束非空值。
- service 的“先查询再插入”可以提供友好提示，但不能替代数据库唯一索引，因为并发请求之间存在竞态窗口。

## 社交用户模型

当前 `User` 聚合账号认证和基础个人资料，包括用户名、密码、联系方式、昵称、头像、简介、性别、生日、地区、状态及时间字段。

以下数据不应继续塞进 `users` 表：

- 兴趣和标签：多值且经常用于匹配、推荐，应使用关联表。
- 关注、好友和拉黑：属于用户关系，应使用关系表。
- 在线状态和连接信息：变化频繁，适合独立会话存储或 Redis。
- 动态、评论和互动计数：属于内容或互动业务。

判断标准是字段是否具有独立生命周期、是否为多值关系、是否高频变化，以及是否需要单独查询和扩展。

## Service 边界

参考 `go-ws-srv/content-service` 后，当前采用“一个可部署服务模块对应一个 `Service`”的方式。`Service` 只保存 DAO、配置等共享依赖，内容、标签、评论等业务方法可以按不同文件组织，但仍挂在同一个 `Service` 上。只有模块之间具备独立依赖、生命周期或部署边界时，才继续拆成多个 Service。

DTO 不是层数越多越好。当前 protobuf 请求和响应已经承担传输对象职责；只有当 HTTP、gRPC、内部任务等多个入口需要不同模型，或者业务层需要摆脱生成代码依赖时，再增加独立 DTO 更合适。

### 为什么 Service 不直接使用 `sql.DB`

`sql.DB` 是数据库连接池和 SQL 执行工具，不是用户业务能力。Service 如果直接使用它，就必须同时负责 SQL、表名、字段扫描和数据库错误转换，业务层会与 PostgreSQL 细节混在一起。当前由 DAO 持有 `sql.DB` 并实现 SQL，Service 只调用 `UserRepository` 描述的用户持久化能力。

`UserRepository` 并不是所有小项目都必须存在。它目前最直接的价值是 service 测试可以注入内存 fake，不需要真实 PostgreSQL；以后也可以替换 DAO 实现，而不修改注册登录流程。如果项目只追求最少代码，也可以让 Service 依赖具体的 `*dao.Dao`，仍然由 DAO 执行 SQL，只是 service 测试会更依赖数据库或需要更重的测试方案。无论是否保留接口，都不建议让 Service 直接操作 `sql.DB`。

增加 friend、message 后，不应把所有方法合并成一个巨大的 Repository 接口。可以继续用小而聚焦的 `UserRepository`、`FriendRepository`、`MessageRepository`，并仅向业务注入它实际需要的能力。一个 `Service` 持有多个 DAO 依赖并不自动等于架构错误，但当这些业务出现独立配置、独立生命周期、独立数据所有权或独立部署需求时，才是拆分 Service 或微服务的明确信号。当前阶段保留 `UserRepository`，等 friend、message 的真实用例出现后再决定边界，避免提前设计。
