# myim

init功能：
1.用户注册，登录，落表，// 短链接
2.send msg,recv msg; // 长连接，websocket

arch:
1.单体

当前 message 协议按“枚举、可复用消息结构、CG/GC 业务对”排列。客户端到服务端使用 CG，服务端到客户端使用 GC。当前消息只通过 WebSocket 在内存连接中投递，尚未持久化；下一阶段将使用 PostgreSQL。
