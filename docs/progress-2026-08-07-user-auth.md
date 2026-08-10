# 用户注册与登录链路进度

- 时间：2026-08-07 15:33:34 +08:00
- 任务：在不改变现有 `http -> service -> dao -> model` 架构的前提下，完善用户注册、登录、数据持久化和验证。
- 完成时间：2026-08-07 15:34:53 +08:00
- 修订时间：2026-08-07 15:57:10 +08:00
- 第二次修订时间：2026-08-07 16:26:51 +08:00
- 本轮验证复核时间：2026-08-07 16:29:17 +08:00
- Service 与 PostgreSQL 知识修订时间：2026-08-07 16:59:40 +08:00
- HTTP 与 Agent 上下文修订时间：2026-08-07 17:39:36 +08:00
- Router 与 App 生命周期修订时间：2026-08-07 17:59:20 +08:00
- App 配置与 HTTP 通用错误修订时间：2026-08-07 18:19:42 +08:00
- VS Code 调试与 App 停止生命周期修订时间：2026-08-10 10:41:01 +08:00
- 具体 DAO 与个人资料接口修订时间：2026-08-10 11:22:18 +08:00
- Git 忽略与他人资料接口修订时间：2026-08-10 11:54:04 +08:00
- 当前进度：100%，实现与验证均已完成。

## 已完成

- 扩展用户 protobuf，增加登录请求与登录响应，重新生成 Go 代码。
- 完善 `User` 模型，密码字段不再参与 JSON 输出，并增加最后登录时间。
- 使用 PostgreSQL 参数化 SQL 实现用户表初始化、创建、查询、重复检查和最后登录时间更新。
- 注册时验证用户名和密码，使用 bcrypt 保存密码哈希，并处理用户名重复。
- 登录时统一处理未知用户和错误密码，签发带过期时间的 HS256 JWT。
- 增加注册和登录 HTTP 路由，补齐请求绑定、状态码映射和服务启动链路。
- 在 `/test` 增加 service 行为测试，覆盖校验、重复注册、密码哈希、登录令牌和错误凭证。

## 2026-08-07 响应与代码规范修订

- 删除 `Register`、`Login` 纯转发封装，业务逻辑直接保留在 `HandleCGUserRegister`、`HandleCGUserLogin`。
- 删除 user handler 私有的 `writeUserResponse`，所有响应统一调用 `internal/httpx.Response`。
- 统一响应根据 `Accept` 或请求 `Content-Type` 返回 protobuf 或 protobuf JSON；JSON 保留 proto 字段名并输出零值字段。
- 请求绑定从仅 JSON 调整为按 `Content-Type` 自动选择 JSON 或 protobuf binding。
- 配置恢复为当前阶段的固定值，不读取环境变量。
- DAO SQL 按 `SELECT`、`FROM`、`WHERE` 等结构分行，提高可读性，不改变执行语义。
- 在 `skills/skill.md` 记录禁止无意义透传封装以及 model tag 对齐规则。

## 2026-08-07 用户模型与协议修订

- protobuf 响应的前两个字段统一改为 `error_code`、`error_msg`，并重新生成 Go 代码。
- User 增加手机号、邮箱、昵称、头像、简介、性别、生日、地区、状态和更新时间。
- PostgreSQL schema 增加兼容已有表的 `ADD COLUMN IF NOT EXISTS`，并为非空手机号、邮箱建立部分唯一索引。
- DAO 创建与查询同步覆盖 User 的全部持久化字段。
- 删除 `serviceError` 和一次性输入校验函数，注册登录流程改为直接、线性的业务代码。
- 新增 `docs/face-knowledge.md`，记录协议、安全、数据库、用户模型和 Service 边界知识点。
- `skills/skill.md` 增加 SQL 排版、直写代码、protobuf 响应头和 face 用语规则。

## 2026-08-07 Service 与 PostgreSQL 修订

- 参考 `go-ws-srv/content-service`，改为由启动入口创建 DAO 并管理关闭，`Service` 构造函数显式接收配置和 DAO 接口。
- `Service` 不再创建数据库、不再保存关闭函数，也不再复制 JWT 密钥和过期时间。
- 新增结构体、字段和包级错误变量补齐用途注释，并在 `skills/skill.md` 固化规则。
- `docs/face-knowledge.md` 增加 PostgreSQL 对象层级、启动初始化、版本化 migration、MongoDB 差异、基础 CRUD 和查询 API 知识。

## 2026-08-07 HTTP 与 Agent 接续修订

- 注释规范调整为字段、接口方法和包级错误变量统一写在声明右侧；删除简单声明的占位注释。
- Gin 初始化改为显式安装 Logger、Recovery，关闭默认代理信任，并在初始化阶段一次性注册路由。
- 新增 `GET /health` 健康检查。
- 使用标准库 `http.Server` 配置请求超时、Header 上限、系统信号监听和 5 秒优雅退出。
- 新增 HTTP 初始化测试，验证健康检查、注册路由和登录路由。
- `docs/face-knowledge.md` 增加 Gin、HTTP Server、中间件、超时、健康检查和优雅退出知识。
- README 增加业务目标、参考项目、当前架构、接口、数据库现状和新 agent 接续步骤。

## 2026-08-07 Router 与 App 生命周期修订

- `app/http` 重命名为 `app/router`，负责 Gin 初始化、路由和业务 handler。
- `internal/http` 重命名为 `internal/httpx`，避免与标准库 `net/http` 包名冲突。
- 新增根 `app.App`，集中创建 Config、DAO、Service、Router 和 `http.Server`。
- HTTP 启动、信号监听、优雅关闭和 PostgreSQL 关闭从 main 移入 `App.Run`。
- main 缩减为创建并运行 App，便于后续接入 WebSocket、gRPC 和后台任务。
- `skills/skill.md` 按工作流程、目录文档、架构实现、代码格式、协议和用语重新分类。
- README 和 face 文档同步新的目录与应用生命周期。

## 2026-08-07 App 配置与 HTTP 通用错误修订

- `App` 增加 Config 字段，与 DAO、HTTP Server 一起作为应用级依赖持有。
- DAO 局部变量统一命名为 `dao`，并在 skill 中禁止使用 `store` 替代。
- `internal/httpx` 增加 `ErrInvalidRequest`、`ErrInternalServer` 通用传输层错误。
- router 的请求绑定错误不再依赖 user service 错误。
- README 和 face 文档记录 `app/app.go` 当前保留理由以及未来微服务目录迁移方式。

## 验证结果

- `go test -count=1 ./...`：通过，`myim/test` 注册登录测试通过。
- `go vet ./...`：通过，无静态检查问题。
- `go build ./...`：通过，所有包编译成功。

## 2026-08-10 VS Code 调试与 App 生命周期修订

- 新增 `.vscode/launch.json`，可从 VS Code 直接调试 `app/cmd` 入口。
- 保留阻塞语义明确的 `App.Run`，没有改名为通常表示立即返回的 `Start`。
- 新增 `App.Stop(ctx)`，统一关闭 HTTP Server 和 PostgreSQL DAO，并保证关闭逻辑只执行一次。
- `Run` 在系统信号和 HTTP Server 退出两条路径上都调用 `Stop`，关闭失败时保留完整错误。
- README 和 face 文档同步调试方式、生命周期命名以及 `UserRepository` 的收益与边界。
- `go test -count=1 ./...`、`go vet ./...`、`go build ./...` 均通过。

## 当前固定配置

- PostgreSQL：`user=postgres password=123456 host=localhost port=5432 dbname=test sslmode=disable`。
- HTTP 监听地址：`:8080`。
- JWT 签名密钥：`myim-development-secret`。

## 2026-08-10 具体 DAO 与个人资料接口修订

- 删除 `UserRepository`，`service.Service` 直接持有一个具体 `*dao.Dao`，但 Service 仍不接触 `sql.DB` 或 SQL。
- `user.proto` 中所有响应的 `error_code`、`error_msg` 增加行尾注释，并通过 `api/protobuf/gen.bat` 重新生成 `user.pb.go`。
- 新增 `ResUserProfile` 协议和 `GET /user-profile`，从 Bearer Token 获取当前用户身份并返回不含密码的完整个人资料。
- 明确“我的资料”和“他人公开资料”使用独立接口及响应字段，避免手机号、邮箱、登录时间等私有数据泄露。
- Service 测试调整为输入校验、令牌拒绝和 HTTP 鉴权测试；具体 DAO 成功路径后续使用 PostgreSQL 集成测试覆盖。
- 移除 `.gitignore` 对 `/test` 的忽略，使项目规定目录中的测试可以进入版本管理。
- `go test -count=1 ./...`、`go vet ./...`、`go build ./...` 均通过。

## JWT 使用说明与空请求协议修订

- `user.proto` 增加空的 `ReqUserProfile`，并通过生成脚本更新 `user.pb.go`。
- `skills/skill.md` 增加无参数 protobuf 请求也必须声明空 Request message 的规则。
- 清理进度文档中每次修改的具体时间，并在 skill 中禁止继续记录此类时间。
- face 文档增加 JWT 结构、签发、请求携带、校验、重复登录和失效机制知识。
- `go test -count=1 ./...`、`go vet ./...`、`go build ./...` 均通过。

## 2026-08-10 Git 忽略与他人资料接口修订

- 恢复进度文档的具体时间记录，并在 skill 中恢复对应要求。
- 确认 `/test/` 未被 Git 跟踪；`/docs/`、`/skills/` 已被忽略，但本轮 Git 索引移除因 `.git` 写权限审批失败而未完成。
- 新增 `ReqOtherUserProfile`、`ResOtherUserProfile` 和 `GET /users/:user_id/profile`。
- 他人资料接口要求 Bearer Token，只返回公开资料字段，不返回登录账号、联系方式、状态和登录时间。
- `go test -count=1 ./...`、`go vet ./...`、`go build ./...` 均通过。
