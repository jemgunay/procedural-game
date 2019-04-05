# Game Test

This is a multiplayer game test with the pixel game engine package.

## Build & Run

```bash
go mod download
cd cmd/client
go build && ./client
```

## TODO

- UI for starting/stopping/joining server.
- Procedural road generation:
    - Use perlin noise to pick random points on map.
    - Join procedurally selected points with straight roads.
- Procedurally generated houses with accessible interiors.
- Continuous world tile generation upon exploring unseen territory.
- Cars.
- Weapons (and ability for players to die).
- TOML config.
- Store/read server state to/from disk.
- FPS counter enable/disable.