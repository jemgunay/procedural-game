# Multiplayer Procedural Game Demo

A cheeky 2D multiplayer procedural terrain game demo made using the pixel game engine.

WASD to move character, RF to zoom in/out.

## Build & Run

```bash
go mod download
go build && ./procedural-game
```

## Package Executable & Assets

```bash
make package
```

Generated zip archive can be found in the `build` directory.

## TODO

- UI for stopping server.
- Procedurally generated houses with accessible interiors.
- Continuous world tile generation upon exploring unseen territory.
- Cars.
- Weapons (and ability for players to die).
- Persistent TOML config produced fro setting menu.
- Store/read server state to/from disk to allow restarts.
- FPS counter enable/disable.