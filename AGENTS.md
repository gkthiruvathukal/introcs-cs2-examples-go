# Repository Guidelines

## Project Structure & Module Organization

This repository contains Go examples for common data structures, intended as a companion to the Python examples in `introcs-cs2-examples-python`. Each data structure lives in its own subdirectory under `data_structures/`, with a self-contained Go module (`go.mod`) per example. All source files for a given example live flat in that directory.

```
data_structures/
  doubly_linked_list/
    dll.go       ← data structure
    render.go    ← card/viewport rendering (pure functions)
    main.go      ← Bubble Tea app (model, update, view)
    go.mod
    go.sum
    README.md
```

Future data structures follow the same pattern: one subdirectory, one module, three logical layers (data structure, rendering, TUI app).

## Build & Run

Each example is a standalone binary. From the example directory:

```bash
go build -o <name> .
./<name>
```

Or run directly:

```bash
go run .
```

No Makefile is required at this stage, but one may be added if the number of examples grows.

## TUI Architecture

All TUI examples use [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-style Model/Update/View) with [Bubbles](https://github.com/charmbracelet/bubbles) for input and viewport widgets, and [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling.

The three layers in each example are:

### 1. Data structure (`dll.go` etc.)

Pure Go struct and methods. No TUI dependencies. Matches the interface of the Python counterpart:
- stable node IDs assigned at insertion time
- `NodeSnapshots()` returns a slice of display-safe structs for the renderer
- `RestoreFromSnapshots()` is used by undo/redo

### 2. Rendering (`render.go` etc.)

Pure functions that convert `NodeSnapshot` slices into printable lines. No Bubble Tea or Lipgloss imports. Three responsibilities:
- `buildDLLSnapshots` / `buildDLLCards` — turn snapshots into boxed card slices
- `renderHorizontalDLL` — assemble cards into a horizontal chain with `head`/`tail` labels
- `buildHorizontalDLLView` — viewport logic: which nodes are visible given available width and scroll offset

Keeping rendering separate from the TUI means it can be tested independently and reused.

### 3. TUI app (`main.go`)

A single Bubble Tea `model` struct containing the data structure, element type, undo/redo stacks, scroll state, Bubbles widgets, and a `shouldQuit` flag. The `Update` function handles `tea.WindowSizeMsg` and `tea.KeyMsg`; all slash-command logic is in `dispatch`. The `View` function calls the rendering layer and assembles the final string.

### Type system

All values are stored as `any`. The `elemType` enum (`typeInt`, `typeFloat`, `typeStr`, `typeBool`, `typeAny`) controls how `/append` and `/random` input is converted. Comparison in `Find` and `Remove` is done by `fmt.Sprintf("%v", ...)`, which is sufficient for the educational use case.

### Undo/redo

Every mutating command calls `recordUndo()` before modifying the structure. A `snapshot` captures the full node list and current type name. `applySnapshot` restores both. The redo stack is cleared on every new mutation.

## Coding Style

- Standard Go formatting (`gofmt`). Four-space indentation is enforced by the formatter.
- Keep each layer (data structure / rendering / TUI) in its own file.
- Rendering functions must remain pure: no global state, no I/O, no TUI imports.
- Prefer explicit error returns over panics.
- Node IDs are assigned by the data structure and must be stable across insert/remove operations; the TUI relies on them for `/insert-after`, `/insert-before`, and `/remove-node`.

## Testing

There are no automated tests yet. When tests are added, place them in `*_test.go` files in the same directory and run with `go test ./...`.

## Commit & PR Guidelines

Use short, imperative commit messages such as `Add stack TUI` or `Fix viewport scroll clamping`. Keep changes small and focused. Do not mix data-structure changes with rendering or TUI changes in the same commit when the diff is large enough to be confusing.
