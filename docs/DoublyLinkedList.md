# Doubly Linked List TUI (Go)

`data_structures/doubly_linked_list/` is an interactive terminal demo for a doubly linked list, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). It is a Go port of the Python/Textual version in `introcs-cs2-examples-python`.

## Running

```bash
cd data_structures/doubly_linked_list
go run .
```

Or build a binary first:

```bash
go build -o dll-tui .
./dll-tui
```

## Layout

The screen has three sections:

1. **Panel** — a horizontal chain of boxed node cards showing the current list from head to tail.
2. **Log** — a scrollable history of command results, errors, and timing.
3. **Input** — a single command line at the bottom.

## Panel

Each node is rendered as a card:

```
          head                                    tail
┌──────────────────┐               ┌──────────────────┐
│ id: node-1       │ <->       <-> │ id: node-4       │
│ data: 42         │               │ data: 99         │
│ prev: /          │               │ prev: node-3     │
│ next: node-2     │               │ next: /          │
└──────────────────┘               └──────────────────┘
```

- `id` is the stable node label assigned at insertion time (e.g. `node-1`).
- `prev: /` and `next: /` mean the node has no predecessor or successor.
- When the full chain is too wide to fit the terminal, the middle collapses into a gap card (`N hidden`) while `head` and `tail` remain visible. Use `←`/`→` to scroll through the hidden middle.

## Commands

Enter commands in the input box and press Enter.

### Doubly Linked List

| Command | Description |
|---|---|
| `/append VALUE [VALUE ...]` | Add one or more values at the tail |
| `/insert-at INDEX VALUE` | Insert before position (`0` = new head, `size` = new tail) |
| `/insert-after NODE-ID VALUE` | Insert immediately after the node with that id |
| `/insert-before NODE-ID VALUE` | Insert immediately before the node with that id |
| `/remove-at INDEX` | Remove the node at zero-based position |
| `/remove-node NODE-ID` | Remove the node with that id |
| `/remove-value VALUE` | Remove the first node whose value matches |
| `/search VALUE` | Report index and node-id of the first match, or "not found" |
| `/at INDEX` | Inspect the value at zero-based offset from the head |
| `/show-backward` | Display all values from tail to head |

### Common

| Command | Description |
|---|---|
| `/random N` | Append N random values for the current type |
| `/clear` | Remove all values |
| `/show` or `/print` | Display current contents head → tail |
| `/save PATH` | Save the current type and contents to a JSON file |
| `/load PATH` | Load a previously saved session |
| `/undo` | Restore the previous state (structure + type) |
| `/redo` | Reapply the most recently undone state |
| `/type` | Show the current type constraint |
| `/type int\|float\|str\|bool\|any` | Set the type constraint |
| `/left` or `←` | Scroll the panel one node to the left |
| `/right` or `→` | Scroll the panel one node to the right |
| `/help` | Show the full command reference |
| `/quit` or `Ctrl+D` | Exit |

Every command logs its execution time in the log panel.

## Type System

The default type is `int`. All values entered via `/append` are converted to the current type before insertion.

| Type | `/append` input | `/random` generates |
|---|---|---|
| `int` | parsed as integer | random integers in [-100, 100] |
| `float` | parsed as float | random floats in [-100.0, 100.0] |
| `str` | kept as-is | random words from `data/words.txt` (or NATO alphabet fallback) |
| `bool` | `true`/`false`, `yes`/`no`, `on`/`off`, `1`/`0` | random `true` or `false` |
| `any` | kept as string | mix of int, string, and bool |

## Undo / Redo

Every mutating command (append, insert, remove, clear, type change) saves a snapshot before it runs. `/undo` restores the previous snapshot; `/redo` reapplies one that was undone. Any new mutation after an undo clears the redo stack.

Snapshots preserve both the node list (with stable node IDs) and the current type constraint.

## Save / Load

`/save PATH` writes a JSON file containing the current type and the list of values:

```json
{
  "type": "int",
  "items": [10, 20, 30]
}
```

`/load PATH` restores values from that file into a fresh list, replacing current contents. The saved type is applied first so values are converted correctly.

## How the Code Is Organised

### `dll.go` — data structure

`DoublyLinkedList` is a standard doubly linked list with one addition: each node carries a stable integer `nodeID` assigned at insertion. The ID counter never resets during a session, so IDs are unique even after removals.

Key methods:

- `Insert(data, nodeID...)` — append at tail; accepts an optional explicit ID for restore
- `InsertAt(index, data)` — insert before position
- `InsertAfter(nodeID, data)` — insert after a node found by ID
- `InsertBefore(nodeID, data)` — insert before a node found by ID
- `Remove(data)` — remove first match by value (string-formatted comparison)
- `RemoveAt(index)` — remove by position
- `RemoveByNodeID(id)` — remove by node ID
- `Find(data)` — returns `*FindResult{Index, NodeID}` or nil
- `NodeSnapshots()` — returns display-safe structs used by the renderer
- `RestoreFromSnapshots(snaps)` — used by undo/redo to rebuild the list with original IDs

### `render.go` — rendering

Pure functions with no TUI dependencies:

- `buildDLLSnapshots` converts a `[]NodeSnapshot` and visible-index slice into a `[]cardSnapshot`, inserting gap entries where indices are non-contiguous.
- `buildDLLCards` turns each snapshot into a slice of strings forming a Unicode box.
- `renderHorizontalDLL` joins cards with ` <-> ` connectors and prepends a `head`/`tail` label line.
- `buildHorizontalDLLView` implements the scrolling viewport: given available pixel width and a scroll offset, it decides which nodes are visible and how many are hidden left and right.

### `main.go` — Bubble Tea app

The `model` struct holds:
- `*DoublyLinkedList` and `elemType`
- `undoHistory` / `redoHistory` as `[]snapshot`
- `scrollOffset` for the horizontal panel
- Bubbles `textinput.Model` and `viewport.Model`

`Update` handles window resize, arrow keys (scroll), Ctrl+D (quit), and Enter (dispatch command). All slash-command logic is in `dispatch`, which returns a new model. `View` calls the rendering layer and assembles the panel, log viewport, and input into a single string.
