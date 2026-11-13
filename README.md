# calpick

An interactive, dependency-free terminal date picker for Go CLI tools.

`calpick` replaces commands such as `mytool 2026-01-01` with a clickable
calendar. Click a date to select it; click `<` or `>` to change months.
Keyboard navigation is available as a fallback.

```text
<        2026 Nov         >
Sun Mon Tue Wed Thu Fri Sat
  1   2   3   4   5   6   7
  8   9  10  11  12  13  14
 15  16  17  18  19  20  21
 22  23  24  25  26  27  28
 29  30
```

## Run

```sh
nix develop
go run ./cmd/calpick
go run ./cmd/calpick 2026-01-01   # start at a specific date
```

The selected date is printed to stdout as `YYYY-MM-DD`. The interface is drawn
on the controlling terminal, so redirecting stdout does not mix escape
sequences into the result.

## Keys

| Key | Action |
| --- | --- |
| Left / `h` | Previous day |
| Right / `l` | Next day |
| Up / `k` | Previous week |
| Down / `j` | Next week |
| PageUp / PageDown | Previous / next month |
| Enter | Select the highlighted date |
| `q` / Esc / Ctrl-C | Cancel |

## Library

```go
import "calpick/datepicker"

picked, err := datepicker.Pick(datepicker.WithInitial(time.Now()))
if err != nil {
    // datepicker.ErrCanceled means the user cancelled.
}
fmt.Println(picked.Format("2006-01-02"))
```

`Pick` blocks until a date is selected, the user cancels, or an error occurs.
It draws on the controlling terminal and restores it before returning. It is
not safe for concurrent use.

The package uses only the Go standard library. It contains no cgo code, so
`GOOS`/`GOARCH` cross-compilation works without a C toolchain.

The module path is currently the local name `calpick`. Before importing the
library from another module, change it in `go.mod` to the repository path you
publish under (for example `github.com/<owner>/calpick`).

## Platforms

- Linux and macOS: raw mode and terminal size use `syscall` ioctls.
- Windows: console modes and mouse input use `ReadConsoleInputW`; Windows
  Terminal or conhost with virtual terminal processing is required. Ctrl-Break
  and console close events are handled as cancellation where the OS allows it.
- tmux and screen: enable mouse reporting (`set -g mouse on`) for clicks.

Keyboard navigation works without mouse support. The calendar needs at least
28 columns and 8 rows.
