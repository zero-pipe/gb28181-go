# Architecture

## Phase 0 — Protocol toolkit

```
digest / ssrc / transport / sdp / ptz / manscdp / mansrtsp
```

No SIP stack. Pure encode/decode helpers.

## Phase 2 — Platform SIP server (current)

```
gb28181-go/
├── server/     # REGISTER / MESSAGE / INVITE / SUBSCRIBE (sipgo)
├── session/    # invite / record / preset waiters
├── cascade/    # upstream platform client
└── (phase 0 packages)
```

### Callback model

Host apps implement:

- `AuthResolver` — Digest password + known device?
- `RegisterHandler` — persist register / unregister
- `MessageHandler` — keepalive / catalog / deviceInfo / alarm / position
- `TelemetryHook` — optional (Redis timestamps, metrics)

Library types: `Peer`, `InviteTarget`, `manscdp.*` — **no host domain imports**.

### Host adapter (zero-web-server)

```
3rdpart usage:
  github.com/zero-pipe/gb28181-go/server
        ▲
internal/infrastructure/sip  (bridge + thin wrappers)
        ▲
application/device|play|playback|ptz|…
```

`sipinfra.Server` keeps the previous ZWS-facing API (`*domaindevice.Device`, etc.) and converts to library `Peer` at the boundary.
