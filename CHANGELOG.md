# 变更记录（Changelog）

遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 与 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Fixed

- 修复 `session` 包等待器（Record / Preset / Status / Control）的竞态：当设备响应先于 `Wait` 到达时，
  旧实现会因 `pending` 条目已被 `Handle` 删除而误报超时。`Wait` 现在接收 `Register` 返回的通道，
  以通道中的缓冲结果为准。
- `server` 包不再强制使用全局 `log`：新增 `Config.Logger`（实现 `Printf` 即可），为 `nil` 时回退
  `log.Default()`；宿主可注入自己的日志实现或使用空实现静默。

### Added

- `session` 等待器单元测试（响应先到、分片聚合、超时取消）。
- CI：`gofmt` / `go vet` / `go test`。
- `.gitignore`、`CONTRIBUTING.md`、`CHANGELOG.md`。

## [0.1.0]

- Phase 0：协议工具包（Digest / SSRC / Transport / SDP / PTZ / MANSCDP / MANSRTSP）。
- Phase 2：平台侧 SIP Server、INVITE 会话、录像/预置位等待、级联客户端。
