# Multiplayer Procedural Game Demo

A cheeky 2D multiplayer procedural terrain game demo made using the pixel game engine.

WASD to move character, Up/Down to zoom in/out.

## Build & Run

Install `pixel` package dependencies: https://github.com/faiface/pixel#requirements

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

- Weapons:
    - Bullet position & collisions processed server side?
    - Weapon & ammo types/ammo pick ups.
    - Player death & random position respawning.
- Switch sprites depending on active weapon/walking & shooting animations.
- Redesign message poller to serialise request processing - can then remove all Mutexes.
- Slow player down in sand/water:
    - Only show head in water and prevent shooting.
- Procedurally generated buildings.
    - Accessible interiors?
- Cars - using A* to navigate between road nodes.
- Store/read server state to/from disk to allow restarts.
- FPS counter enable/disable.
- Continuous world tile generation upon exploring unseen territory.
