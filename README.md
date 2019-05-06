# 2D Multiplayer Game Demo

This is a cheeky multiplayer game test with the pixel game engine package.

## Build & Run

```bash
go mod download
cd cmd/client
go build && ./client
```

## TODO

- UI for starting/stopping/joining server.
- Procedurally generated houses with accessible interiors.
- Continuous world tile generation upon exploring unseen territory.
- Cars.
- Weapons (and ability for players to die).
- TOML config.
- Store/read server state to/from disk.
- FPS counter enable/disable.