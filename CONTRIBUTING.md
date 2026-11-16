# 贡献指南（Contributing）

`gb28181-go` 是零依赖业务模型的 GB28181 协议库，欢迎提交 Issue / PR。

## 约定

- 仅做协议：不引入数据库、Redis、流媒体、HTTP 框架等业务依赖。
- 公共 API 不引用任何业务工程的领域模型；宿主通过 `Handlers` 与 `Logger` 接入。
- 使用 Conventional Commits：`feat:` / `fix:` / `refactor:` / `docs:` / `test:` / `chore:`。
- 提交前确保：

```bash
gofmt -l .
go vet ./...
go test ./...
```

## 新增能力

- 协议报文：优先落在 `manscdp` / `mansrtsp` / `sdp` / `ptz` 等部门包，保持纯函数。
- 会话等待：在 `session` 包实现 `Register` / `Handle` / `Wait(sn, ch, timeout)` / `Cancel`
  四件套；`Wait` 必须读取 `Register` 返回的通道，避免响应先到导致的竞态。
- 日志：库内统一使用 `s.logf`，不要直接调用 `log.Printf`。

## 分支与 PR

从 `master` 切 `feature/<name>` 或 `fix/<name>`，PR 描述写清背景、改动、验证方式。
