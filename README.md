# Game Test

This is a multiplayer game test with the pixel game engine package.

## Build & Run

```bash
go mod download
cd cmd/client
go build && ./client
```

## TODO

- Colourise grass tiles based on height/z, i.e. lighter/darker.
- Hook server up to client via event channel.
- UI for starting/stopping/joining server.
- Add sand around water.
- Procedural road generation:
    - Use perlin noise to pick random points on map.
    - Join procedurally selected points with straight roads.
- Procedurally generated houses with accessible interiors.
- Continuous world tile generation upon exploring unseen territory.
- Cars.
- Weapons (and ability for players to die).
- TOML config.
- FPS counter enable/disable.