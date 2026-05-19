# Go Data Structures Examples

Educational Go examples for common data structures (CS1/CS2 level) — a companion to the [Python version](https://github.com/gkthiruvathukal/introcs-cs2-examples-python).

## About This Repo

This is the beginning of an effort to make the CS1/CS2 data structure examples available in Go. The Python repo has been the primary home for these demos, but Go offers a compelling alternative: static typing, fast binaries, and a mature TUI ecosystem via the [Charmbracelet](https://github.com/charmbracelet) suite.

We started with the **doubly linked list** to prove viability — specifically to answer: can Go and its TUI frameworks match the interactive demo quality of the Python/Textual versions? The answer is yes. The doubly linked list TUI runs as a self-contained binary, supports the full command set (insert, remove, search, undo/redo, save/load, horizontal scrolling), and requires no runtime or virtual environment.

Future examples will follow the same pattern.

## Setup

Each example is a self-contained Go module. You need Go 1.22 or later.

```bash
cd data_structures/doubly_linked_list
go build -o dll-tui .
./dll-tui
```

Or run without building:

```bash
go run .
```

## Examples

| Example | Path | Status |
|---|---|---|
| Doubly Linked List | `data_structures/doubly_linked_list/` | Complete |

## Related

- **Python version**: [introcs-cs2-examples-python](https://github.com/gkthiruvathukal/introcs-cs2-examples-python) — the original repo with Python/Textual TUIs for stack, queue, deque, list, singly linked list, and doubly linked list.

## TUI Framework

All interactive demos use [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-style Model/Update/View architecture) with [Bubbles](https://github.com/charmbracelet/bubbles) for input/viewport widgets and [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling — all from the Charmbracelet project.

For walkthroughs of each example, see the `docs/` directory:

- [`docs/DoublyLinkedList.md`](docs/DoublyLinkedList.md)
