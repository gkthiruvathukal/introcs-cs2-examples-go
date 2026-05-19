package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	panelStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	logStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	cyanStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	boldStyle   = lipgloss.NewStyle().Bold(true)
)

// ── type system ───────────────────────────────────────────────────────────────

type elemType int

const (
	typeInt elemType = iota
	typeFloat
	typeStr
	typeBool
	typeAny
)

func (t elemType) String() string {
	switch t {
	case typeInt:
		return "int"
	case typeFloat:
		return "float"
	case typeStr:
		return "str"
	case typeBool:
		return "bool"
	default:
		return "any"
	}
}

func parseElemType(s string) (elemType, error) {
	switch strings.ToLower(s) {
	case "int":
		return typeInt, nil
	case "float":
		return typeFloat, nil
	case "str":
		return typeStr, nil
	case "bool":
		return typeBool, nil
	case "any", "":
		return typeAny, nil
	}
	return typeAny, fmt.Errorf("unknown type %q — use: int float str bool any", s)
}

func convertValue(raw string, t elemType) (any, error) {
	switch t {
	case typeInt:
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to int", raw)
		}
		return v, nil
	case typeFloat:
		v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to float", raw)
		}
		return v, nil
	case typeBool:
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "true", "t", "1", "yes", "y", "on":
			return true, nil
		case "false", "f", "0", "no", "n", "off":
			return false, nil
		}
		return nil, fmt.Errorf("cannot convert %q to bool", raw)
	default: // str or any → keep as string
		return raw, nil
	}
}

func convertMany(tokens []string, t elemType) ([]any, error) {
	out := make([]any, len(tokens))
	for i, tok := range tokens {
		v, err := convertValue(tok, t)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// ── shlex tokeniser ───────────────────────────────────────────────────────────

func shellSplit(s string) ([]string, error) {
	var tokens []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	for i := 0; i < len(s); i++ {
		ch := rune(s[i])
		switch {
		case inSingle:
			if ch == '\'' {
				inSingle = false
			} else {
				cur.WriteRune(ch)
			}
		case inDouble:
			if ch == '"' {
				inDouble = false
			} else if ch == '\\' && i+1 < len(s) {
				i++
				cur.WriteByte(s[i])
			} else {
				cur.WriteRune(ch)
			}
		case ch == '\'':
			inSingle = true
		case ch == '"':
			inDouble = true
		case unicode.IsSpace(ch):
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(ch)
		}
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote")
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens, nil
}

// ── random generation ─────────────────────────────────────────────────────────

var builtinWords = []string{
	"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel",
	"india", "juliet", "kilo", "lima", "mike", "november", "oscar", "papa",
	"quebec", "romeo", "sierra", "tango", "uniform", "victor", "whiskey",
	"xray", "yankee", "zulu",
}

func loadWords() []string {
	path := filepath.Join("..", "data", "words.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return builtinWords
	}
	var words []string
	for _, line := range strings.Split(string(data), "\n") {
		if w := strings.TrimSpace(line); w != "" {
			words = append(words, w)
		}
	}
	if len(words) == 0 {
		return builtinWords
	}
	return words
}

func randomValue(t elemType, rng *rand.Rand, words []string) any {
	switch t {
	case typeInt:
		return rng.Intn(200) - 100
	case typeFloat:
		return rng.Float64()*200 - 100
	case typeBool:
		return rng.Intn(2) == 1
	case typeStr:
		return words[rng.Intn(len(words))]
	default: // any
		builders := []func() any{
			func() any { return rng.Intn(200) - 100 },
			func() any { return words[rng.Intn(len(words))] },
			func() any { return rng.Intn(2) == 1 },
		}
		return builders[rng.Intn(len(builders))]()
	}
}

// ── undo/redo snapshot ────────────────────────────────────────────────────────

type snapshot struct {
	typeName string
	nodes    []NodeSnapshot
}

func takeSnapshot(dll *DoublyLinkedList, t elemType) snapshot {
	return snapshot{typeName: t.String(), nodes: dll.NodeSnapshots()}
}

func applySnapshot(dll *DoublyLinkedList, s snapshot) elemType {
	dll.RestoreFromSnapshots(s.nodes)
	t, _ := parseElemType(s.typeName)
	return t
}

// ── save / load ───────────────────────────────────────────────────────────────

type sessionFile struct {
	Type  string `json:"type"`
	Items []any  `json:"items"`
}

func saveSession(path string, dll *DoublyLinkedList, t elemType) error {
	items := dll.ToList()
	sf := sessionFile{Type: t.String(), Items: items}
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func loadSession(path string) (sessionFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return sessionFile{}, err
	}
	var sf sessionFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return sessionFile{}, err
	}
	return sf, nil
}

// ── timing ────────────────────────────────────────────────────────────────────

func formatElapsed(d time.Duration) string {
	ns := d.Nanoseconds()
	switch {
	case ns < 1_000:
		return fmt.Sprintf("%d ns", ns)
	case ns < 1_000_000:
		return fmt.Sprintf("%.1f us", float64(ns)/1_000)
	case ns < 1_000_000_000:
		return fmt.Sprintf("%.1f ms", float64(ns)/1_000_000)
	default:
		return fmt.Sprintf("%.3f s", float64(ns)/1_000_000_000)
	}
}

// ── model ─────────────────────────────────────────────────────────────────────

const maxSize = 1000

type model struct {
	dll          *DoublyLinkedList
	elemType     elemType
	input        textinput.Model
	logVP        viewport.Model
	logLines     []string
	undoHistory  []snapshot
	redoHistory  []snapshot
	scrollOffset int
	width        int
	height       int
	words        []string
	rng          *rand.Rand
	ready        bool
	shouldQuit   bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "/help  /append <v>  /insert-at <i> <v>  /insert-after <id> <v>  /insert-before <id> <v>  /remove-at <i>  /remove-node <id>  /remove-value <v>  /search <v>  /at <i>  /show-backward  /quit"
	ti.Focus()
	ti.CharLimit = 512

	return model{
		dll:      NewDoublyLinkedList(),
		elemType: typeInt,
		input:    ti,
		words:    loadWords(),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// ── update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
			m.logVP = viewport.New(m.logWidth(), m.logHeight())
		} else {
			m.logVP.Width = m.logWidth()
			m.logVP.Height = m.logHeight()
		}
		m.logVP.SetContent(strings.Join(m.logLines, "\n"))
		m.logVP.GotoBottom()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlD, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyLeft:
			m.scrollOffset = max(m.scrollOffset-1, 0)
			return m, nil
		case tea.KeyRight:
			m.scrollOffset++
			return m, nil
		case tea.KeyEnter:
			raw := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")
			if raw != "" {
				started := time.Now()
				m = m.dispatch(raw)
				elapsed := time.Since(started)
				m = m.appendLog(dimStyle.Render("time: " + formatElapsed(elapsed)))
			}
			if m.shouldQuit {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) dispatch(raw string) model {
	parts := strings.SplitN(raw, " ", 2)
	verb := strings.ToLower(parts[0])
	arg := ""
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}

	switch verb {

	case "/append":
		if arg == "" {
			return m.appendLog(redStyle.Render("Usage: /append <value> [more ...]"))
		}
		tokens, err := shellSplit(arg)
		if err != nil || len(tokens) == 0 {
			return m.appendLog(redStyle.Render("Usage: /append <value> [more ...]"))
		}
		values, err := convertMany(tokens, m.elemType)
		if err != nil {
			return m.appendLog(redStyle.Render(err.Error()))
		}
		if m.dll.size+len(values) > maxSize {
			return m.appendLog(redStyle.Render(fmt.Sprintf("Overflow: capacity is %d", maxSize)))
		}
		m = m.recordUndo()
		for _, v := range values {
			m.dll.Insert(v)
		}
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("append(%d values) → size=%d", len(values), m.dll.size)))

	case "/insert-at":
		tokens, err := shellSplit(arg)
		if err != nil || len(tokens) != 2 {
			return m.appendLog(redStyle.Render("Usage: /insert-at <index> <value>"))
		}
		idx, err := strconv.Atoi(tokens[0])
		if err != nil || idx < 0 {
			return m.appendLog(redStyle.Render("index must be a non-negative integer"))
		}
		vals, err := convertMany(tokens[1:], m.elemType)
		if err != nil {
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m = m.recordUndo()
		if err := m.dll.InsertAt(idx, vals[0]); err != nil {
			m.undoHistory = m.undoHistory[:len(m.undoHistory)-1]
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("insert-at(%d, %v) → size=%d", idx, vals[0], m.dll.size)))

	case "/insert-after":
		tokens, err := shellSplit(arg)
		if err != nil || len(tokens) != 2 {
			return m.appendLog(redStyle.Render("Usage: /insert-after <node-id> <value>"))
		}
		nodeID, err := strconv.Atoi(tokens[0])
		if err != nil || nodeID < 0 {
			return m.appendLog(redStyle.Render("node-id must be a non-negative integer"))
		}
		vals, err := convertMany(tokens[1:], m.elemType)
		if err != nil {
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m = m.recordUndo()
		if err := m.dll.InsertAfter(nodeID, vals[0]); err != nil {
			m.undoHistory = m.undoHistory[:len(m.undoHistory)-1]
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("insert-after(node-%d, %v) → size=%d", nodeID, vals[0], m.dll.size)))

	case "/insert-before":
		tokens, err := shellSplit(arg)
		if err != nil || len(tokens) != 2 {
			return m.appendLog(redStyle.Render("Usage: /insert-before <node-id> <value>"))
		}
		nodeID, err := strconv.Atoi(tokens[0])
		if err != nil || nodeID < 0 {
			return m.appendLog(redStyle.Render("node-id must be a non-negative integer"))
		}
		vals, err := convertMany(tokens[1:], m.elemType)
		if err != nil {
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m = m.recordUndo()
		if err := m.dll.InsertBefore(nodeID, vals[0]); err != nil {
			m.undoHistory = m.undoHistory[:len(m.undoHistory)-1]
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("insert-before(node-%d, %v) → size=%d", nodeID, vals[0], m.dll.size)))

	case "/remove-at":
		idx, err := strconv.Atoi(arg)
		if err != nil || idx < 0 {
			return m.appendLog(redStyle.Render("Usage: /remove-at <non-negative-index>"))
		}
		m = m.recordUndo()
		if err := m.dll.RemoveAt(idx); err != nil {
			m.undoHistory = m.undoHistory[:len(m.undoHistory)-1]
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("remove-at(%d) → size=%d", idx, m.dll.size)))

	case "/remove-node":
		nodeID, err := strconv.Atoi(arg)
		if err != nil || nodeID < 0 {
			return m.appendLog(redStyle.Render("Usage: /remove-node <node-id>"))
		}
		m = m.recordUndo()
		if err := m.dll.RemoveByNodeID(nodeID); err != nil {
			m.undoHistory = m.undoHistory[:len(m.undoHistory)-1]
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("remove-node(node-%d) → size=%d", nodeID, m.dll.size)))

	case "/remove-value":
		if arg == "" {
			return m.appendLog(redStyle.Render("Usage: /remove-value <value>"))
		}
		tokens, err := shellSplit(arg)
		if err != nil || len(tokens) != 1 {
			return m.appendLog(redStyle.Render("Usage: /remove-value <value>"))
		}
		vals, err := convertMany(tokens, m.elemType)
		if err != nil {
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m = m.recordUndo()
		if err := m.dll.Remove(vals[0]); err != nil {
			m.undoHistory = m.undoHistory[:len(m.undoHistory)-1]
			return m.appendLog(redStyle.Render(err.Error()))
		}
		return m.appendLog(greenStyle.Render(fmt.Sprintf("remove-value(%v) → size=%d", vals[0], m.dll.size)))

	case "/search":
		if arg == "" {
			return m.appendLog(redStyle.Render("Usage: /search <value>"))
		}
		tokens, err := shellSplit(arg)
		if err != nil || len(tokens) != 1 {
			return m.appendLog(redStyle.Render("Usage: /search <value>"))
		}
		vals, err := convertMany(tokens, m.elemType)
		if err != nil {
			return m.appendLog(redStyle.Render(err.Error()))
		}
		result := m.dll.Find(vals[0])
		if result == nil {
			return m.appendLog(cyanStyle.Render(fmt.Sprintf("search(%v) → not found", vals[0])))
		}
		return m.appendLog(cyanStyle.Render(fmt.Sprintf("search(%v) → found at index %d  node-id: %d", vals[0], result.Index, result.NodeID)))

	case "/at":
		idx, err := strconv.Atoi(arg)
		if err != nil || idx < 0 {
			return m.appendLog(redStyle.Render("Usage: /at <non-negative-index>"))
		}
		v, err := m.dll.At(idx)
		if err != nil {
			return m.appendLog(redStyle.Render(err.Error()))
		}
		return m.appendLog(cyanStyle.Render(fmt.Sprintf("at(%d) → %v", idx, v)))

	case "/show-backward":
		items := m.dll.ToListBackward()
		parts := make([]string, len(items))
		for i, v := range items {
			parts[i] = fmt.Sprintf("%v", v)
		}
		return m.appendLog(cyanStyle.Render("tail → head: [" + strings.Join(parts, " ") + "]"))

	case "/show", "/print":
		items := m.dll.ToList()
		if len(items) == 0 {
			return m.appendLog(yellowStyle.Render("(empty)"))
		}
		parts := make([]string, len(items))
		for i, v := range items {
			parts[i] = fmt.Sprintf("%v", v)
		}
		return m.appendLog(cyanStyle.Render("None ← " + strings.Join(parts, " ↔ ") + " → None"))

	case "/random":
		count, err := strconv.Atoi(arg)
		if err != nil || count <= 0 {
			return m.appendLog(redStyle.Render("Usage: /random <positive-count>"))
		}
		actual := min(count, maxSize-m.dll.size)
		if actual == 0 {
			return m.appendLog(yellowStyle.Render("List is already full; no values added."))
		}
		m = m.recordUndo()
		for i := 0; i < actual; i++ {
			m.dll.Insert(randomValue(m.elemType, m.rng, m.words))
		}
		m.scrollOffset = 0
		msg := fmt.Sprintf("random(%d) added %d values → size=%d", count, actual, m.dll.size)
		if actual < count {
			msg += fmt.Sprintf(" (truncated to capacity)")
		}
		return m.appendLog(greenStyle.Render(msg))

	case "/clear":
		m = m.recordUndo()
		prev := m.dll.size
		m.dll.Clear()
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("clear() removed %d values → size=0", prev)))

	case "/type":
		if arg == "" {
			return m.appendLog(cyanStyle.Render("current type: " + m.elemType.String()))
		}
		t, err := parseElemType(arg)
		if err != nil {
			return m.appendLog(redStyle.Render(err.Error()))
		}
		m = m.recordUndo()
		m.elemType = t
		return m.appendLog(greenStyle.Render("type constraint set to: " + t.String()))

	case "/save":
		if arg == "" {
			return m.appendLog(redStyle.Render("Usage: /save <path>"))
		}
		if err := saveSession(arg, m.dll, m.elemType); err != nil {
			return m.appendLog(redStyle.Render("Save failed: " + err.Error()))
		}
		return m.appendLog(greenStyle.Render(fmt.Sprintf("save() wrote %d values to %q", m.dll.size, arg)))

	case "/load":
		if arg == "" {
			return m.appendLog(redStyle.Render("Usage: /load <path>"))
		}
		sf, err := loadSession(arg)
		if err != nil {
			return m.appendLog(redStyle.Render("Load failed: " + err.Error()))
		}
		t, err := parseElemType(sf.Type)
		if err != nil {
			return m.appendLog(redStyle.Render("Load failed: " + err.Error()))
		}
		m = m.recordUndo()
		m.dll.Clear()
		m.elemType = t
		for _, v := range sf.Items {
			converted, err := convertValue(fmt.Sprintf("%v", v), t)
			if err != nil {
				m.dll.Clear()
				return m.appendLog(redStyle.Render("Load failed: " + err.Error()))
			}
			m.dll.Insert(converted)
		}
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("load() restored %d values from %q", m.dll.size, arg)))

	case "/undo":
		if len(m.undoHistory) == 0 {
			return m.appendLog(yellowStyle.Render("Nothing to undo."))
		}
		m.redoHistory = append(m.redoHistory, takeSnapshot(m.dll, m.elemType))
		s := m.undoHistory[len(m.undoHistory)-1]
		m.undoHistory = m.undoHistory[:len(m.undoHistory)-1]
		m.elemType = applySnapshot(m.dll, s)
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("undo() → size=%d  type=%s", m.dll.size, m.elemType)))

	case "/redo":
		if len(m.redoHistory) == 0 {
			return m.appendLog(yellowStyle.Render("Nothing to redo."))
		}
		m.undoHistory = append(m.undoHistory, takeSnapshot(m.dll, m.elemType))
		s := m.redoHistory[len(m.redoHistory)-1]
		m.redoHistory = m.redoHistory[:len(m.redoHistory)-1]
		m.elemType = applySnapshot(m.dll, s)
		m.scrollOffset = 0
		return m.appendLog(greenStyle.Render(fmt.Sprintf("redo() → size=%d  type=%s", m.dll.size, m.elemType)))

	case "/left":
		m.scrollOffset = max(m.scrollOffset-1, 0)
		return m
	case "/right":
		m.scrollOffset++
		return m

	case "/help":
		lines := []string{
			boldStyle.Render("Commands"),
			boldStyle.Render("  --- Doubly Linked List ---"),
			cyanStyle.Render("  /append") + " VALUE [VALUE ...]       add one or more values at the tail",
			cyanStyle.Render("  /insert-at") + " INDEX VALUE           insert before position (0 = new head)",
			cyanStyle.Render("  /insert-after") + " NODE-ID VALUE       insert after node",
			cyanStyle.Render("  /insert-before") + " NODE-ID VALUE      insert before node",
			cyanStyle.Render("  /remove-at") + " INDEX                  remove node at position",
			cyanStyle.Render("  /remove-node") + " NODE-ID              remove node by id",
			cyanStyle.Render("  /remove-value") + " VALUE               remove first matching node",
			cyanStyle.Render("  /search") + " VALUE                     search for a value",
			cyanStyle.Render("  /at") + " INDEX                         inspect value at index",
			cyanStyle.Render("  /show-backward") + "                    display from tail to head",
			"",
			boldStyle.Render("  --- Common ---"),
			cyanStyle.Render("  /random") + " N                         append N random values",
			cyanStyle.Render("  /clear") + "                            remove all values",
			cyanStyle.Render("  /show") + "  " + cyanStyle.Render("/print") + "                   display current contents",
			cyanStyle.Render("  /save") + " PATH                        save session to file",
			cyanStyle.Render("  /load") + " PATH                        load session from file",
			cyanStyle.Render("  /undo") + "                             restore previous state",
			cyanStyle.Render("  /redo") + "                             restore next undone state",
			cyanStyle.Render("  /type") + " int|float|str|bool|any      get or set type constraint",
			cyanStyle.Render("  /left") + "  " + cyanStyle.Render("/right") + "  ← →               scroll list horizontally",
			cyanStyle.Render("  /help") + "                             show this help",
			cyanStyle.Render("  /quit") + "  " + cyanStyle.Render("Ctrl+D") + "                    exit",
		}
		return m.appendLog(strings.Join(lines, "\n"))

	case "/quit", "/exit":
		m.shouldQuit = true
		return m
	}

	return m.appendLog(redStyle.Render(fmt.Sprintf("Unknown command: %q — type /help", verb)))
}

func (m model) recordUndo() model {
	m.undoHistory = append(m.undoHistory, takeSnapshot(m.dll, m.elemType))
	m.redoHistory = nil
	return m
}

func (m model) appendLog(line string) model {
	m.logLines = append(m.logLines, line)
	if m.ready {
		m.logVP.SetContent(strings.Join(m.logLines, "\n"))
		m.logVP.GotoBottom()
	}
	return m
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m model) logWidth() int  { return max(m.width-4, 20) }
func (m model) logHeight() int { return max(m.height-m.panelHeight()-6, 4) }

func (m model) panelHeight() int { return 8 } // header + 6 card rows + gap

func (m model) View() string {
	if !m.ready {
		return "Initialising…"
	}

	// panel
	snaps := m.dll.NodeSnapshots()
	availWidth := max(m.width-6, 20)
	view := buildHorizontalDLLView(snaps, availWidth, m.scrollOffset)

	header := fmt.Sprintf("Doubly Linked List (%d/%d)  type: %s",
		m.dll.size, maxSize, m.elemType)
	if view.hiddenBefore > 0 || view.hiddenAfter > 0 {
		header += fmt.Sprintf("  visible: %d/%d", len(view.visibleItems), view.itemCount)
	}
	if view.hiddenBefore > 0 {
		header += fmt.Sprintf("  ←%d hidden", view.hiddenBefore)
	}
	if view.hiddenAfter > 0 {
		header += fmt.Sprintf("  %d hidden→", view.hiddenAfter)
	}

	viewSnaps := buildDLLSnapshots(view.visibleItems, view.visibleIndices, view.itemCount)
	cards := buildDLLCards(viewSnaps)
	renderLines := renderHorizontalDLL(cards)
	panelContent := header + "\n" + strings.Join(renderLines, "\n")
	panel := panelStyle.Width(m.width - 4).Render(panelContent)

	// log viewport
	logBox := logStyle.Width(m.width - 4).Height(m.logHeight()).Render(m.logVP.View())

	// input
	inputLine := m.input.View()

	return strings.Join([]string{panel, logBox, inputLine}, "\n")
}

// ── entry point ───────────────────────────────────────────────────────────────

func main() {
	m := initialModel()
	m.logLines = []string{
		boldStyle.Render("Doubly Linked List Demo") + " — type " + cyanStyle.Render("/help") + " for commands.",
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
