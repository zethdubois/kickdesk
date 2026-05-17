package status

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

var ansiStrip = regexp.MustCompile(`\033\[[0-9;]*m`)

type ansiStyle struct {
	bold, dim, green, red, yellow, cyan, reset string
}

func stdoutStyle() ansiStyle {
	s := ansiStyle{}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		s.bold = "\033[1m"
		s.dim = "\033[2m"
		s.green = "\033[32m"
		s.red = "\033[31m"
		s.yellow = "\033[33m"
		s.cyan = "\033[36m"
		s.reset = "\033[0m"
	}
	return s
}

func visibleLen(s string) int {
	return runewidth.StringWidth(ansiStrip.ReplaceAllString(s, ""))
}

func padVisible(s string, width int) string {
	n := visibleLen(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

type tableCell struct {
	text string // may contain ANSI
}

type tableCol struct {
	title string
	width int
}

type tableRow struct {
	cells []tableCell
}

func buildStatusRows(apps []AppStatus, hub bool) []tableRow {
	var rows []tableRow
	for i, a := range apps {
		num := ""
		if hub {
			num = fmt.Sprintf("[%d]", i+1)
		}
		fam := a.PortFamily
		if fam == "" {
			fam = "—"
		}
		git := formatGit(a.Branch, a.Dirty)
		url := a.URL
		if url == "" {
			url = "—"
		}

		migrate := a.Migrate
		if migrate == "" {
			migrate = "n/a"
		}
		addRow := func(first bool, p PortLine) {
			rows = append(rows, tableRow{cells: statusRowCells(hub, first, num, fam, a.Name, migrate, git, url, p)})
		}

		if len(a.Ports) == 0 {
			addRow(true, PortLine{})
			continue
		}
		for j, p := range a.Ports {
			addRow(j == 0, p)
		}
	}
	return rows
}

func statusRowCells(hub, first bool, num, fam, app, migrate, git, url string, p PortLine) []tableCell {
	var cells []tableCell
	if hub {
		key := ""
		if first {
			key = num
		}
		cells = append(cells, tableCell{text: key})
	}
	f, n, m, g, u := "", "", "", "", ""
	if first {
		f, n, m, g, u = fam, app, migrate, git, url
	}
	role, portStr, state := "—", "—", "—"
	if p.Port > 0 || p.Role != "" {
		role = portRole(p)
		portStr = fmt.Sprintf("%d", p.Port)
		state = p.State
	}
	cells = append(cells,
		tableCell{text: f},
		tableCell{text: n},
		tableCell{text: role},
		tableCell{text: portStr},
		tableCell{text: state},
		tableCell{text: m},
		tableCell{text: g},
		tableCell{text: u},
	)
	return cells
}

func portRole(p PortLine) string {
	if p.Role != "" {
		return p.Role
	}
	return "port"
}

func renderTable(apps []AppStatus, hub bool) {
	st := stdoutStyle()
	rows := buildStatusRows(apps, hub)

	cols := []tableCol{
		{title: "FAM", width: 4},
		{title: "APP", width: 10},
		{title: "SERVICE", width: 10},
		{title: "PORT", width: 5},
		{title: "STATUS", width: 6},
		{title: "MIGRATE", width: 10},
		{title: "GIT", width: 14},
		{title: "URL", width: 28},
	}
	if hub {
		cols = append([]tableCol{{title: "#", width: 4}}, cols...)
	}

	for _, r := range rows {
		for i, c := range r.cells {
			if i >= len(cols) {
				break
			}
			w := visibleLen(c.text)
			if c.text == "" {
				w = 0
			}
			if w > cols[i].width {
				cols[i].width = w
			}
		}
	}
	for i := range cols {
		if w := runewidth.StringWidth(cols[i].title); w > cols[i].width {
			cols[i].width = w
		}
	}

	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = st.bold + st.cyan + c.title + st.reset
	}

	hr := func(left, mid, right, fill string) {
		var b strings.Builder
		b.WriteString(left)
		for i, c := range cols {
			if i > 0 {
				b.WriteString(mid)
			}
			b.WriteString(strings.Repeat(fill, c.width+2))
		}
		b.WriteString(right)
		fmt.Println(b.String())
	}

	hr("┌", "┬", "┐", "─")
	printTableLine(cols, header, st, "│")
	hr("├", "┼", "┤", "─")

	rowIdx := 0
	for ai, a := range apps {
		portCount := len(a.Ports)
		if portCount == 0 {
			portCount = 1
		}
		for pi := 0; pi < portCount; pi++ {
			r := rows[rowIdx]
			rowIdx++
			line := make([]string, len(cols))
			for i, c := range cols {
				raw := ""
				if i < len(r.cells) {
					raw = r.cells[i].text
				}
				switch c.title {
				case "STATUS":
					line[i] = styleState(st, raw)
				case "MIGRATE":
					line[i] = styleMigrate(st, raw)
				case "GIT":
					line[i] = styleGit(st, raw)
				default:
					line[i] = raw
				}
			}
			printTableLine(cols, line, st, "│")
			if pi == portCount-1 && ai < len(apps)-1 {
				hr("├", "┼", "┤", "─")
			}
		}
	}
	hr("└", "┴", "┘", "─")
}

func printTableLine(cols []tableCol, cells []string, st ansiStyle, border string) {
	var b strings.Builder
	b.WriteString(border)
	for i, c := range cols {
		if i > 0 {
			b.WriteString(border)
		}
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteString(" ")
		b.WriteString(padVisible(cell, c.width))
		b.WriteString(" ")
	}
	b.WriteString(border)
	fmt.Println(b.String())
}

func styleState(st ansiStyle, state string) string {
	switch state {
	case "up":
		return st.bold + st.green + state + st.reset
	case "down":
		return st.red + state + st.reset
	default:
		return state
	}
}

func styleMigrate(st ansiStyle, migrate string) string {
	switch {
	case migrate == "ok":
		return st.green + migrate + st.reset
	case strings.HasPrefix(migrate, "pending:"):
		return st.yellow + migrate + st.reset
	case migrate == "unavailable", migrate == "n/a", migrate == "—":
		return st.dim + migrate + st.reset
	case migrate == "error":
		return st.red + migrate + st.reset
	default:
		return migrate
	}
}

func styleGit(st ansiStyle, git string) string {
	if strings.Contains(git, "dirty") {
		return st.yellow + git + st.reset
	}
	if git == "—" {
		return st.dim + git + st.reset
	}
	return git
}

func renderAppPortTable(ports []PortLine) {
	st := stdoutStyle()
	cols := []tableCol{
		{title: "SERVICE", width: 10},
		{title: "PORT", width: 5},
		{title: "STATUS", width: 6},
	}
	var rows []tableRow
	for _, p := range ports {
		rows = append(rows, tableRow{cells: []tableCell{
			{text: portRole(p)},
			{text: fmt.Sprintf("%d", p.Port)},
			{text: p.State},
		}})
	}
	if len(rows) == 0 {
		return
	}
	for _, r := range rows {
		for i, c := range r.cells {
			w := visibleLen(c.text)
			if w > cols[i].width {
				cols[i].width = w
			}
		}
	}
	for i := range cols {
		if w := runewidth.StringWidth(cols[i].title); w > cols[i].width {
			cols[i].width = w
		}
	}

	hr := func(left, mid, right, fill string) {
		var b strings.Builder
		b.WriteString(left)
		for i, c := range cols {
			if i > 0 {
				b.WriteString(mid)
			}
			b.WriteString(strings.Repeat(fill, c.width+2))
		}
		b.WriteString(right)
		fmt.Println(b.String())
	}

	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = st.bold + st.cyan + c.title + st.reset
	}
	hr("┌", "┬", "┐", "─")
	printTableLine(cols, header, st, "│")
	hr("├", "┼", "┤", "─")
	for _, r := range rows {
		line := []string{
			r.cells[0].text,
			r.cells[1].text,
			styleState(st, r.cells[2].text),
		}
		printTableLine(cols, line, st, "│")
	}
	hr("└", "┴", "┘", "─")
}
