# MyIM 协作规则

## 工作流程

- 不要使用 superpower 的 skill。
- 修改前先阅读 README、当前进度文档和相关代码，保留用户已有改动。
- 每次更新代码后都必须保证 `go test ./...`、`go vet ./...`、`go build ./...` 通过。

## 目录与文档

- 所有测试文件统一放在 `/test` 目录。
- 所有文档文件统一放在 `/docs` 目录。
- 每次更新代码都要同步 `/docs` 中的进度文档和 README，记录具体时间、任务、进度、改动和验证结果。
- README 保存业务目标、当前架构和 agent 接续上下文；skill 只保存强制协作规则。

## 架构与实现

- 当前保持 Go 单体架构，参考 `go-ws-srv` 的工程模式，但不要提前引入尚未进入需求的基础设施。
- 根 `app.App` 是 composition root，负责依赖组装、服务启动和资源关闭；`main` 保持精简。
- DAO 实例的局部变量统一命名为 `dao`，不要使用 `store` 等替代名称。
- 一个可部署模块对应一个 `service.Service`，共享配置和 DAO 依赖，具体业务按 `service/user.go`、`service/message.go` 等文件组织。
- 不要增加只做参数透传的函数封装。业务逻辑直接放在对外 service 处理方法中，例如 `HandleCGUserRegister`。
- 代码优先直接、线性、易读。注册、登录等短业务流程允许使用清晰的面条式代码。
- 只有存在真实复用、独立复杂性或能显著降低理解成本时才增加封装。

## 代码格式与注释

- model 的 struct tag 和行尾注释按列纵向对齐，保持 `json`、`gorm` tag 和注释位置整齐一致。
- SQL query 使用多行原始字符串，按 `SELECT`、字段、`FROM`、`WHERE`、`VALUES` 等语义分行并缩进，不把完整 SQL 挤在一行。
- 只有含义不直观的结构体才写结构体用途注释；`Service`、`User`、`Dao`、`Config`、`New` 等一眼可理解的声明不要写注释，也不要写 `// Xxx .` 这类占位注释。
- 所有字段、接口方法和包级错误变量的注释都写在声明右侧，不要单独写在声明上方；struct tag 和右侧注释继续保持纵向对齐。

## 协议约定

- HTTP 请求支持 JSON 和 protobuf binding。
- HTTP 响应统一通过 `internal/httpx.Response` 输出，根据 `Accept` 或请求 `Content-Type` 选择 protobuf 或 protobuf JSON。
- protobuf 请求即使没有参数，也必须声明对应的空 Request message，例如 `ReqUserProfile`。
- 所有 protobuf 响应消息的前两个字段固定为 `int32 error_code = 1;` 和 `string error_msg = 2;`，并分别添加“错误码，0表示成功”和“错误信息”的行尾注释。
- 修改 `.proto` 后必须重新生成对应的 `.pb.go`，不要手工修改生成文件。

## 用语约定

- 环境敏感的相关记录统一称为 `face`，相关参与者统一称为 `facer`。
- 在交流、代码注释和文档中都使用上述替代词。
