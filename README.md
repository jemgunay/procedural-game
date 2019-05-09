# 2D Multiplayer Game Demo

A cheeky multiplayer procedural terrain game demo made using the pixel game engine.

## Build & Run

```bash
go mod download
go build && ./game
```

## TODO

- UI for stopping server.
- Procedurally generated houses with accessible interiors.
- Continuous world tile generation upon exploring unseen territory.
- Cars.
- Weapons (and ability for players to die).
- Persistent TOML config produced fro setting menu.
- Store/read server state to/from disk to allow restarts.
- FPS counter enable/disable.