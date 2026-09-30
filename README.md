# myim


最初的基准路径：
1.用户注册，登录，落表，// 短链接
2.send msg,recv msg; // 长连接，websocket

arch:
1.servergo 单体 arch

## Agent 接续上下文

### 项目目标

`myim` 是一个正在增量开发的 Go 单体社交 IM 后端，产品方向参考 Soul 一类社交应用。开发顺序是：

1. 用户注册、登录、PostgreSQL 持久化和短连接认证。
2. WebSocket 长连接、发送消息和接收消息。
3. 用户资料、社交关系、内容和互动能力。

当前已完成用户注册登录链路，并新增可独立运行的 message、content 和 social service。四个进程仍保存在同一个 Go module 中，暂不引入 Redis、gRPC 和服务发现。

### 参考项目

同级工作区中的 `..\gorm_test\go-ws-srv` 是功能和工程模式参考。可以学习它的 service、DAO、HTTP 和后续 IM 实现，但 `myim` 当前保持简单单体架构，不直接搬入参考项目的全部基础设施。

### 当前架构

```text
apps/<service>/app        服务组装、依赖注入和生命周期
apps/<service>/cmd        服务进程入口（当前 user/message/content/social）
apps/<service>/config     固定配置（PostgreSQL、端口、JWT 等）
apps/<service>/dao        PostgreSQL 初始化、schema 和参数化 SQL
apps/<service>/model      持久化模型和领域状态常量
apps/<service>/router     Gin 初始化、HTTP/WebSocket 路由和协议绑定
apps/<service>/service    业务流程、鉴权和服务内状态
apps/content-service/event Content 点赞事件 envelope 与 Publisher
api/protobuf/<domain>     user/message/content/social 的 .proto 与生成的 .pb.go
internal/httpx            protobuf / JSON 统一响应和传输层错误
client                    Go WebSocket 测试客户端及 content 点赞客户端
test                      所有 Go 测试文件（统一目录）
docs                      API、进度、架构、Kafka 和 face 知识记录
skills                    本项目强制协作规则（含 face/facer 用语约定）
.vscode/launch.json       四个服务及全部服务的调试配置
go.mod、go.sum            Go module 和依赖锁定
```

服务默认端口为 user `:8080`、message `:8081`、content `:8082`、social `:8083`。四个服务仍位于同一个 Go module 中，可独立启动但共享 PostgreSQL、统一 JWT 配置和代码仓库；这属于当前的单体基础架构，不代表已经完成生产级微服务治理。

每个 `apps/<service>/app.App` 都是对应服务的 composition root。四个进程拥有独立端口和生命周期；鉴权实现集中在 `internal/auth`，运行时通过 `MYIM_JWT_SECRET` 和 `MYIM_DATABASE_DSN` 读取共享配置。应用启动会拒绝开发 JWT 密钥，详细边界见 [鉴权与配置进度记录](docs/progress-2026-09-30-auth-config.md)。

`App.Run` 启动服务并阻塞等待退出，`App.Stop` 负责只执行一次 HTTP 优雅关闭和 DAO 关闭。Content 进程还运行 outbox 投递 worker，停机时等待它退出再关闭 Publisher 和 DAO。VS Code 可以直接选择 `.vscode/launch.json` 中的 `Run user service` 配置启动调试；运行前需要保证固定配置所指向的 PostgreSQL 可连接。

当前四个服务已按 `apps/<service>` 形成可独立启动的服务模块。未来继续微服务化时，应在各自服务目录内增加独立基础设施，不再恢复根 `app` 目录。

Content 业务与 Kafka 学习需求记录在 `docs/content-kafka-learning-requirements.md`。Kafka 第一阶段只服务 Content、Follow、Notification 等异步事件，不改动 user 和 message 核心链路。

当前已新增独立 `content-service`，默认监听 `:8082`。阶段 A 提供动态草稿/发布、详情、作者公开列表、删除、点赞/取消点赞和评论/评论列表；点赞关系、计数及 `content.liked.v1`/`content.unliked.v1` 事件现写入同一 PostgreSQL 事务中的 outbox，后台按内容顺序重试投递 Kafka。Kafka 暂时不可用不会丢失已提交事件；消费方需按 `event_id` 去重。

当前已新增独立 `social-service`，默认监听 `:8083`，提供关注、取消关注、关注列表、粉丝列表和关注关系判断。Social 目前只使用 HTTP 短连接和 PostgreSQL，不引入 WebSocket、Redis、Kafka 或 gRPC；关注关系已从 content service 移至 social service。

Content 的 `/content/create` 用于创建草稿或直接发布（由 `publish` 字段决定）；`/content/publish` 用于后续将当前用户自己的草稿发布，适合“先保存、后发布”的流程。

### 当前接口

- user service：`POST /user-register`、`POST /user-login`、`POST /my-profile`、`POST /target-profile`。
- message service：`GET /health`、`GET /ws`、`POST /message/conversations`、`POST /message/history`；`/ws` 使用 Bearer Token 和 protobuf 二进制帧。
- content service：`POST /content/create`、`/content/publish`、`/content/get`、`/content/list`、`/content/delete`、`/content/like`、`/content/unlike`、`/content/comment`、`/content/comments`、`/content/comment/delete`。
- social service：`POST /social/follow`、`POST /social/unfollow`、`POST /social/following-list`、`POST /social/follower-list`、`POST /social/follow-check`。
- 四个服务均提供 `GET /health`，成功返回 HTTP 204。

请求字段、响应字段、状态码和 WebSocket 帧定义见 [`docs/api.md`](docs/api.md)；后续优化顺序和与 face 对话的推进方式见 [`docs/todo.md`](docs/todo.md)。

当前用户自己的资料接口与查看他人公开资料的接口分开设计。自己的资料可以包含手机号、邮箱、账号状态和登录时间；公开资料只能返回昵称、头像、简介等公开字段，避免敏感字段泄露。除健康检查外，当前所有业务 HTTP 请求统一使用 POST。

当前 message 协议按“枚举、可复用消息结构、CG/GC 业务对”排列。客户端到服务端使用 CG，服务端到客户端使用 GC。消息先写入 PostgreSQL `messages` 表，再向在线用户推送；ACK 表示消息已持久化，不表示对方已收到或已读。

登录成功后，客户端在访问需要身份的接口时携带 `Authorization: Bearer <access_token>`。当前访问令牌有效期为 24 小时；重复登录会签发新 token，但不会主动注销此前仍在有效期内的 token。
令牌缺失、格式错误或签名不匹配返回 `invalid access token`；签名和内容合法但超过 `exp` 返回 `expired access token`。两者 HTTP 状态码都是 401。

HTTP 请求支持 JSON 和 protobuf binding；响应根据 `Accept` 或请求 `Content-Type` 返回 protobuf 或 protobuf JSON。所有 protobuf 响应前两个字段固定为 `error_code`、`error_msg`。

`service.Service` 当前直接持有一个具体 `*dao.Dao`。所有 SQL 和 `sql.DB` 仍封装在 DAO 内，Service 不直接操作数据库连接池；新增 friend、message 业务时继续在 DAO 和 Service 中按文件组织，等出现独立部署和数据所有权后再拆服务。

message service 维护本实例内存连接，支持一个用户一条连接、单聊 PUSH 和发送方 ACK。每一对单聊用户对应一个正式 `conversation_id`，`messages.seq` 在会话内递增；消息先写入 PostgreSQL `messages` 表，ACK 表示已持久化，不表示对方已收到或已读。目标离线时消息仍保留；重连后可列出当前用户会话，并按每个会话的 `after_seq` 分页补拉。客户端以 `(conversation_id, seq)` 去重，处理并持久保存消息后才推进本地序号。真实 WebSocket 连接始终保存在所属进程内存，后续 Redis 只保存用户到 gateway 的路由元数据。详细演进计划见 `docs/message-architecture-evolution.md`。

### 本地运行

```bash
go run ./apps/user-service/cmd
go run ./apps/message-service/cmd
go run ./apps/content-service/cmd
go run ./apps/social-service/cmd
```

也可以在 VS Code 中分别选择 `Run user service`、`Run message service`、`Run content service` 或 `Run social service`，或使用 `Run all services` 同时启动四个进程。

### Go 测试客户端

`client` 目录提供两个 Go 客户端，分别固定为两个测试用户，并默认互相发送消息。客户端启动后会提示输入当前用户的登录用户名和密码，登录成功后自动携带 Bearer Token 建立 WebSocket 连接。

```bash
go run ./client/user-cda0
go run ./client/user-53c8
```

两个客户端可分别在两个终端中运行。直接输入文本并回车即可向另一个用户发送消息，输入 `/quit` 关闭连接。客户端直接复用项目生成的 Go protobuf 协议，发送和接收的数据均为 `message.proto` 定义的二进制帧。

### 数据库现状

当前使用 PostgreSQL，四个 DAO 启动时统一执行 `internal/migration` 中已登记的版本化迁移，使用事务、并发锁和 SQL 校验记录；表结构与索引按服务编号，SQLC 直接读取 migration SQL。手动升级和回滚可使用 `go run ./cmd/migrate`，详见 [数据库迁移说明](docs/database-migrations.md)。首次初始化要求空数据库或空 schema，不自动兼容旧开发库。User 目前包含账号、联系方式、基础社交资料、状态和时间字段；兴趣、关系、在线状态、内容等数据应放在独立业务表。

P0-1 migration 已于 2026-09-29 21:06 +08:00 完成，验证结果见 [本次进度记录](docs/progress-2026-09-29-database-migrations.md)。

P1-4 消息历史与离线补拉已于 2026-09-30 11:08 +08:00 完成，接口、权限和验证结果见 [消息历史进度记录](docs/progress-2026-09-30-message-history.md)。

P1-5 Content 点赞事件 Transactional Outbox 已于 2026-09-30 11:20 +08:00 完成，重试、租约和保留策略见 [Outbox 进度记录](docs/progress-2026-09-30-content-outbox.md)。

### 新 Agent 开始工作前

1. 完整阅读 `skills/skill.md`，其规则优先贯穿后续修改。
2. 阅读 `docs/progress-2026-08-07-user-auth.md` 了解已完成内容。
3. 阅读 `docs/face-knowledge.md` 了解现有技术决策和知识记录。
4. 修改前检查工作区状态，保留用户已有改动。
5. 每次代码更新都同步 docs，并确保 `go test ./...`、`go vet ./...`、`go build ./...` 通过。
6. 测试不得遗留后台服务或需要用户手动结束的测试可执行程序。
