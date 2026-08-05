# gb28181-go

标准、通用的 [GB/T 28181](https://www.gb688.cn/)（国标）协议库，Go 实现。

> Phase 0：纯协议层（Digest / SSRC / 流传输模式 / SDP / PTZ / MANSCDP / MANSRTSP）  
> Phase 2：平台侧 SIP Server（REGISTER / MESSAGE / INVITE / SUBSCRIBE）+ 级联客户端  

## 设计原则

- **只做协议**：不依赖数据库、Redis、流媒体、HTTP 框架
- **类型自洽**：公共 API 不引用任何业务工程的 domain 模型
- **回调接入**：REGISTER / MESSAGE 通过 `Handlers` 交给宿主处理业务

## 安装

```bash
go get github.com/zero-pipe/gb28181-go@latest
```

本地与 zero-web-server 联调：

```go
replace github.com/zero-pipe/gb28181-go => ../gb28181-go
```

## 包一览

| 包 | 职责 |
|----|------|
| [`digest`](./digest) | SIP Digest 鉴权 |
| [`ssrc`](./ssrc) | 国标播放 SSRC |
| [`transport`](./transport) | UDP / TCP-ACTIVE / TCP-PASSIVE |
| [`sdp`](./sdp) | Play / Playback / Download SDP |
| [`ptz`](./ptz) | PTZ / 前端命令字 |
| [`manscdp`](./manscdp) | MANSCDP XML |
| [`mansrtsp`](./mansrtsp) | 回放 INFO |
| [`server`](./server) | 平台 SIP Server + 出站控制 |
| [`session`](./session) | INVITE / 录像 / 预置位等待器 |
| [`cascade`](./cascade) | 上级平台注册 / 保活 / 目录推送 |

## 最小 Server 示例

```go
srv, err := server.New(server.Config{
    ID: "34020000002000000001", Domain: "3402000000",
    Password: "12345678", IP: "192.168.1.5", Port: 5060,
}, server.Handlers{
    Auth:     myAuth,
    Register: myRegister,
    Message:  myMessage,
})
_ = srv.Start(ctx)
```

## 路线图

1. ~~Phase 0 — 纯协议工具包~~  
2. ~~Phase 2 — `server` + `session` + `cascade`~~  
3. ~~Wave A — 抓包对齐：NOTIFY / DeviceStatus / MediaStatus / RecordCmd / RecordInfo 扩展 / Catalog Event / BYE 清理~~  
4. ~~Wave B — Guard/TeleBoot/IFrame/AlarmCmd、ConfigDownload、2022 巡航/看守位/PTZPosition、级联 Digest~~  
5. Phase 3 — 示例程序、更多厂商 XML 金样、语音对讲/抓拍完整会话  

## License

MIT
