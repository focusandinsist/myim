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

当前已完成用户注册登录链路，并新增可独立运行的基础 message service。两个进程仍保存在同一个 Go module 中，暂不引入 Redis、Kafka、gRPC 和服务发现。

### 参考项目

同级工作区中的 `..\gorm_test\go-ws-srv` 是功能和工程模式参考。可以学习它的 service、DAO、HTTP 和后续 IM 实现，但 `myim` 当前保持简单单体架构，不直接搬入参考项目的全部基础设施。

### 当前架构

```text
app/app.go    依赖组装、HTTP启动和资源生命周期
app/cmd       精简的进程入口
app/router    Gin初始化、路由和协议处理
app/service   业务流程，不增加无意义转发层
app/dao       PostgreSQL 初始化和参数化 SQL
app/model     持久化模型
apps/message-service/app      message进程组装与生命周期
apps/message-service/cmd      message进程入口
apps/message-service/config   message固定配置
apps/message-service/router   健康检查和WebSocket握手
apps/message-service/service  JWT校验、连接管理和在线消息投递
api/protobuf  请求响应协议及生成代码
internal/httpx protobuf / JSON统一响应
test          统一测试目录
docs          进度和 face 知识记录
skills        本项目强制协作规则
```

根 `app.App` 是用户服务的 composition root。`apps/message-service/app.App` 是消息服务的 composition root。两个进程拥有独立端口和生命周期，但当前共享代码仓库及 JWT 签名密钥。

`App.Run` 启动服务并阻塞等待退出，`App.Stop` 负责只执行一次 HTTP 优雅关闭和 DAO 关闭。VS Code 可以直接选择 `.vscode/launch.json` 中的 `Run myim` 配置启动调试；运行前需要保证固定配置所指向的 PostgreSQL 可连接。

当前整个单体业务都位于 `app` 目录，因此 composition root 保持在 `app/app.go`。未来拆分微服务时，再按 `apps/<service>/internal/...` 整体迁移服务边界，不单独移动一个 App 文件。

### 当前接口

- `POST /user-register`：注册用户，校验用户名和密码，使用 bcrypt 保存密码。
- `POST /user-login`：验证密码，签发 HS256 JWT，更新最后登录时间。
- `POST /my-profile`：通过 `Authorization: Bearer <token>` 获取当前登录用户的完整个人资料，不返回密码。
- `POST /target-profile`：登录后通过 `ReqTargetProfile.user_id` 查询目标用户的公开资料，不返回登录账号、联系方式、账号状态和登录时间。
- `GET /health`：服务健康检查，成功返回 HTTP 204。

message service 当前接口：

- `GET :8081/health`：消息服务健康检查。
- `GET :8081/ws`：携带 Bearer Token 完成 WebSocket Upgrade，之后使用 protobuf 二进制帧在线收发文本消息。

当前用户自己的资料接口与查看他人公开资料的接口分开设计。自己的资料可以包含手机号、邮箱、账号状态和登录时间；公开资料只能返回昵称、头像、简介等公开字段，避免敏感字段泄露。除健康检查外，当前所有业务 HTTP 请求统一使用 POST。

登录成功后，客户端在访问需要身份的接口时携带 `Authorization: Bearer <access_token>`。当前访问令牌有效期为 24 小时；重复登录会签发新 token，但不会主动注销此前仍在有效期内的 token。
令牌缺失、格式错误或签名不匹配返回 `invalid access token`；签名和内容合法但超过 `exp` 返回 `expired access token`。两者 HTTP 状态码都是 401。

HTTP 请求支持 JSON 和 protobuf binding；响应根据 `Accept` 或请求 `Content-Type` 返回 protobuf 或 protobuf JSON。所有 protobuf 响应前两个字段固定为 `error_code`、`error_msg`。

`service.Service` 当前直接持有一个具体 `*dao.Dao`。所有 SQL 和 `sql.DB` 仍封装在 DAO 内，Service 不直接操作数据库连接池；新增 friend、message 业务时继续在 DAO 和 Service 中按文件组织，等出现独立部署和数据所有权后再拆服务。

message service 第一版只维护本实例内存连接，支持一个用户一条连接、在线单聊、目标 PUSH 和发送方 ACK。消息没有落库，目标离线会直接返回错误；ACK 只表示服务端已接受并加入目标连接的发送队列，不表示对方已读或已经可靠持久化。详细演进计划见 `docs/message-architecture-evolution.md`。

### 本地运行

```bash
go run ./app/cmd
go run ./apps/message-service/cmd
```

也可以在 VS Code 中分别选择 `Run myim`、`Run message service`，或使用 `Run user + message services` 同时启动两个进程。

### 数据库现状

当前使用 PostgreSQL，开发阶段由 DAO 执行幂等 schema 初始化。正式环境后续应迁移到版本化 migration。User 目前包含账号、联系方式、基础社交资料、状态和时间字段；兴趣、关系、在线状态、内容等数据应放在独立业务表。

### 新 Agent 开始工作前

1. 完整阅读 `skills/skill.md`，其规则优先贯穿后续修改。
2. 阅读 `docs/progress-2026-08-07-user-auth.md` 了解已完成内容。
3. 阅读 `docs/face-knowledge.md` 了解现有技术决策和知识记录。
4. 修改前检查工作区状态，保留用户已有改动。
5. 每次代码更新都同步 docs，并确保 `go test ./...`、`go vet ./...`、`go build ./...` 通过。
6. 测试不得遗留后台服务或需要用户手动结束的测试可执行程序。
