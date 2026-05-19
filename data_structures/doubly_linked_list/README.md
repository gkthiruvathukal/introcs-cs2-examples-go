# Doubly Linked List TUI (Go)

A port of the Python/Textual doubly linked list demo to Go using [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Requirements

- Go 1.22+

## Build and run

```bash
cd go-dll
go build -o dll-tui .
./dll-tui
```

Or run directly:

```bash
go run .
```

## Commands

| Command | Description |
|---|---|
| `/append VALUE [...]` | Add one or more values at the tail |
| `/insert-at INDEX VALUE` | Insert before position (0 = new head) |
| `/insert-after NODE-ID VALUE` | Insert after the node with that id |
| `/insert-before NODE-ID VALUE` | Insert before the node with that id |
| `/remove-at INDEX` | Remove node at position |
| `/remove-node NODE-ID` | Remove node by id |
| `/remove-value VALUE` | Remove first matching node |
| `/search VALUE` | Search for a value |
| `/at INDEX` | Inspect value at index |
| `/show-backward` | Display from tail to head |
| `/random N` | Append N random values |
| `/clear` | Remove all values |
| `/show` / `/print` | Display current contents |
| `/save PATH` | Save session to JSON file |
| `/load PATH` | Load session from JSON file |
| `/undo` | Restore previous state |
| `/redo` | Restore next undone state |
| `/type int\|float\|str\|bool\|any` | Get or set element type |
| `/left` / `/right` or `←` / `→` | Scroll list horizontally |
| `/help` | Show command reference |
| `/quit` or `Ctrl+D` | Exit |

## Type system

The default type is `int`. Use `/type str`, `/type float`, `/type bool`, or `/type any` to switch. Values entered via `/append` or `/random` are converted to the current type.

## Notes

- Node ids are stable across insert/remove and shown in each card as `node-N`.
- The panel scrolls horizontally when the list is too wide to fit: use `←`/`→` keys or `/left`/`/right`.
- If `../data/words.txt` exists relative to the working directory, `/random` under `str` or `any` draws from it; otherwise a built-in NATO phonetic alphabet word list is used.
