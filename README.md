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

当前只完成第一阶段的基础注册登录链路，不要提前引入微服务、Redis、Kafka 等尚未进入需求的组件。

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
api/protobuf  请求响应协议及生成代码
internal/httpx protobuf / JSON统一响应
test          统一测试目录
docs          进度和 face 知识记录
skills        本项目强制协作规则
```

根 `app.App` 持有 Config、DAO 和 HTTP Server，负责创建 Service、Router，并统一管理启动与关闭。一个可部署模块对应一个 `service.Service`，不同业务可以拆到 `service/user.go`、`service/message.go` 等文件，但当前不增加嵌套 Service 或 DTO 层。

`App.Run` 启动服务并阻塞等待退出，`App.Stop` 负责只执行一次 HTTP 优雅关闭和 DAO 关闭。VS Code 可以直接选择 `.vscode/launch.json` 中的 `Run myim` 配置启动调试；运行前需要保证固定配置所指向的 PostgreSQL 可连接。

当前整个单体业务都位于 `app` 目录，因此 composition root 保持在 `app/app.go`。未来拆分微服务时，再按 `apps/<service>/internal/...` 整体迁移服务边界，不单独移动一个 App 文件。

### 当前接口

- `POST /user-register`：注册用户，校验用户名和密码，使用 bcrypt 保存密码。
- `POST /user-login`：验证密码，签发 HS256 JWT，更新最后登录时间。
- `GET /health`：服务健康检查，成功返回 HTTP 204。

HTTP 请求支持 JSON 和 protobuf binding；响应根据 `Accept` 或请求 `Content-Type` 返回 protobuf 或 protobuf JSON。所有 protobuf 响应前两个字段固定为 `error_code`、`error_msg`。

### 数据库现状

当前使用 PostgreSQL，开发阶段由 DAO 执行幂等 schema 初始化。正式环境后续应迁移到版本化 migration。User 目前包含账号、联系方式、基础社交资料、状态和时间字段；兴趣、关系、在线状态、内容等数据应放在独立业务表。

### 新 Agent 开始工作前

1. 完整阅读 `skills/skill.md`，其规则优先贯穿后续修改。
2. 阅读 `docs/progress-2026-08-07-user-auth.md` 了解已完成内容。
3. 阅读 `docs/face-knowledge.md` 了解现有技术决策和知识记录。
4. 修改前检查工作区状态，保留用户已有改动。
5. 每次代码更新都同步 docs，并确保 `go test ./...`、`go vet ./...`、`go build ./...` 通过。
