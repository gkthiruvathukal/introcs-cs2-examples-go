package main

import (
	"fmt"
	"strings"
)

const displayValueLimit = 40

func formatDisplayValue(v any) string {
	s := fmt.Sprintf("%v", v)
	if len(s) > displayValueLimit {
		return s[:displayValueLimit] + "..."
	}
	return s
}

type cardSnapshot struct {
	kind        string // "node" or "gap"
	index       int
	nodeID      int
	nodeLabel   string
	data        any
	prevLabel   string
	nextLabel   string
	hiddenCount int
}

func buildDLLSnapshots(items []NodeSnapshot, visibleIndices []int, itemCount int) []cardSnapshot {
	var snaps []cardSnapshot
	prevIndex := -1
	for i, item := range items {
		idx := visibleIndices[i]
		if prevIndex >= 0 && idx-prevIndex > 1 {
			snaps = append(snaps, cardSnapshot{
				kind:        "gap",
				hiddenCount: idx - prevIndex - 1,
			})
		}
		snaps = append(snaps, cardSnapshot{
			kind:      "node",
			index:     idx,
			nodeID:    item.NodeID,
			nodeLabel: item.NodeLabel,
			data:      item.Data,
			prevLabel: item.PrevLabel,
			nextLabel: item.NextLabel,
		})
		prevIndex = idx
	}
	return snaps
}

func buildDLLCards(snaps []cardSnapshot) [][]string {
	var cards [][]string
	for _, s := range snaps {
		if s.kind == "gap" {
			hiddenText := fmt.Sprintf("%d hidden", s.hiddenCount)
			rows := []string{"...", hiddenText, "...", "..."}
			w := maxLen(rows)
			card := []string{fmt.Sprintf("┌%s┐", strings.Repeat("─", w+2))}
			for _, row := range rows {
				card = append(card, fmt.Sprintf("│ %s │", center(row, w)))
			}
			card = append(card, fmt.Sprintf("└%s┘", strings.Repeat("─", w+2)))
			cards = append(cards, card)
			continue
		}
		dataText := formatDisplayValue(s.data)
		rows := []string{
			fmt.Sprintf("id: %s", s.nodeLabel),
			fmt.Sprintf("data: %s", dataText),
			fmt.Sprintf("prev: %s", s.prevLabel),
			fmt.Sprintf("next: %s", s.nextLabel),
		}
		w := maxLen(rows)
		card := []string{fmt.Sprintf("┌%s┐", strings.Repeat("─", w+2))}
		for _, row := range rows {
			card = append(card, fmt.Sprintf("│ %-*s │", w, row))
		}
		card = append(card, fmt.Sprintf("└%s┘", strings.Repeat("─", w+2)))
		cards = append(cards, card)
	}
	return cards
}

func estimateHorizontalWidth(cards [][]string) int {
	if len(cards) == 0 {
		return 0
	}
	const connectorWidth = 5 // " <-> "
	total := 0
	for _, c := range cards {
		total += len([]rune(c[0]))
	}
	total += connectorWidth * (len(cards) - 1)
	return total
}

func renderHorizontalDLL(cards [][]string) []string {
	if len(cards) == 0 {
		return []string{"(empty)"}
	}
	const connector = " <-> "
	totalWidth := estimateHorizontalWidth(cards)

	// compute card start positions
	starts := make([]int, len(cards))
	cursor := 0
	for i, c := range cards {
		starts[i] = cursor
		cursor += len([]rune(c[0])) + len(connector)
	}

	// head/tail label line
	labelRunes := []rune(strings.Repeat(" ", totalWidth))
	placeLabel := func(label string, cardIdx int) {
		cardWidth := len([]rune(cards[cardIdx][0]))
		start := starts[cardIdx] + (cardWidth-len(label))/2
		for i, ch := range label {
			pos := start + i
			if pos >= 0 && pos < len(labelRunes) {
				labelRunes[pos] = ch
			}
		}
	}
	placeLabel("head", 0)
	placeLabel("tail", len(cards)-1)
	lines := []string{strings.TrimRight(string(labelRunes), " ")}

	// card rows
	connectorRow := len(cards[0]) / 2
	for row := range cards[0] {
		var parts []string
		for ci, card := range cards {
			parts = append(parts, card[row])
			if ci < len(cards)-1 {
				if row == connectorRow {
					parts = append(parts, connector)
				} else {
					parts = append(parts, strings.Repeat(" ", len(connector)))
				}
			}
		}
		lines = append(lines, strings.Join(parts, ""))
	}
	return lines
}

// viewport result
type dllViewResult struct {
	visibleItems    []NodeSnapshot
	visibleIndices  []int
	hiddenBefore    int
	hiddenAfter     int
	scrollOffset    int
	itemCount       int
	renderWidth     int
	middleWindowSize int
}

func buildHorizontalDLLView(items []NodeSnapshot, availableWidth int, scrollOffset int) dllViewResult {
	itemCount := len(items)
	if itemCount == 0 {
		return dllViewResult{}
	}

	// check if everything fits
	allIndices := make([]int, itemCount)
	for i := range allIndices {
		allIndices[i] = i
	}
	fullCards := buildDLLCards(buildDLLSnapshots(items, allIndices, itemCount))
	if estimateHorizontalWidth(fullCards) <= availableWidth {
		return dllViewResult{
			visibleItems:     items,
			visibleIndices:   allIndices,
			itemCount:        itemCount,
			renderWidth:      estimateHorizontalWidth(fullCards),
			middleWindowSize: max(itemCount-4, 0),
		}
	}

	for _, anchorCount := range []int{2, 1} {
		leftCount := min(anchorCount, itemCount)
		rightCount := min(anchorCount, max(itemCount-leftCount, 0))
		middleStartMin := leftCount
		middleEndMax := itemCount - rightCount
		middleCount := max(middleEndMax-middleStartMin, 0)

		if middleCount <= 0 {
			vis := items
			visIdx := allIndices
			rw := estimateHorizontalWidth(buildDLLCards(buildDLLSnapshots(vis, visIdx, itemCount)))
			if rw <= availableWidth {
				return dllViewResult{
					visibleItems:   vis,
					visibleIndices: visIdx,
					itemCount:      itemCount,
					renderWidth:    rw,
				}
			}
			continue
		}

		clamped := min(max(scrollOffset, 0), max(middleCount-1, 0))
		middleStart := middleStartMin + clamped

		for middleLen := middleCount; middleLen >= 0; middleLen-- {
			middleEnd := min(middleStart+middleLen, middleEndMax)
			leftHidden := max(middleStart-middleStartMin, 0)
			rightHidden := max(middleEndMax-middleEnd, 0)

			var visItems []NodeSnapshot
			var visIdx []int

			for i := 0; i < leftCount; i++ {
				visItems = append(visItems, items[i])
				visIdx = append(visIdx, i)
			}

			// gap placeholder (we mark with a zero NodeSnapshot and handle in snapshot builder)
			if leftHidden > 0 {
				visItems = append(visItems, NodeSnapshot{NodeLabel: "gap"})
				visIdx = append(visIdx, middleStart-1)
			}

			for i := middleStart; i < middleEnd; i++ {
				visItems = append(visItems, items[i])
				visIdx = append(visIdx, i)
			}

			if rightHidden > 0 {
				visItems = append(visItems, NodeSnapshot{NodeLabel: "gap"})
				visIdx = append(visIdx, middleEnd)
			}

			for i := itemCount - rightCount; i < itemCount; i++ {
				visItems = append(visItems, items[i])
				visIdx = append(visIdx, i)
			}

			// build snapshots honouring gaps
			var snaps []cardSnapshot
			prevIdx := -1
			for i, item := range visItems {
				idx := visIdx[i]
				if item.NodeLabel == "gap" {
					count := 0
					if prevIdx >= 0 {
						count = idx - prevIdx
					}
					snaps = append(snaps, cardSnapshot{kind: "gap", hiddenCount: count})
					prevIdx = idx
					continue
				}
				if prevIdx >= 0 && idx-prevIdx > 1 {
					snaps = append(snaps, cardSnapshot{kind: "gap", hiddenCount: idx - prevIdx - 1})
				}
				snaps = append(snaps, cardSnapshot{
					kind:      "node",
					index:     idx,
					nodeID:    item.NodeID,
					nodeLabel: item.NodeLabel,
					data:      item.Data,
					prevLabel: item.PrevLabel,
					nextLabel: item.NextLabel,
				})
				prevIdx = idx
			}

			cards := buildDLLCards(snaps)
			rw := estimateHorizontalWidth(cards)
			if rw <= availableWidth {
				// collect only real items
				var realItems []NodeSnapshot
				var realIdx []int
				for i, item := range visItems {
					if item.NodeLabel != "gap" {
						realItems = append(realItems, item)
						realIdx = append(realIdx, visIdx[i])
					}
				}
				return dllViewResult{
					visibleItems:     realItems,
					visibleIndices:   realIdx,
					hiddenBefore:     leftHidden,
					hiddenAfter:      rightHidden,
					scrollOffset:     clamped,
					itemCount:        itemCount,
					renderWidth:      rw,
					middleWindowSize: max(middleEnd-middleStart, 1),
				}
			}
		}
	}

	// fallback: show just head and tail
	vis := []NodeSnapshot{items[0]}
	visIdx := []int{0}
	if itemCount > 1 {
		vis = append(vis, items[itemCount-1])
		visIdx = append(visIdx, itemCount-1)
	}
	return dllViewResult{
		visibleItems:   vis,
		visibleIndices: visIdx,
		hiddenAfter:    max(itemCount-2, 0),
		itemCount:      itemCount,
		renderWidth:    availableWidth,
		middleWindowSize: 1,
	}
}

// helpers

func maxLen(rows []string) int {
	m := 0
	for _, r := range rows {
		if len(r) > m {
			m = len(r)
		}
	}
	return m
}

func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	total := width - len(s)
	left := total / 2
	right := total - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
