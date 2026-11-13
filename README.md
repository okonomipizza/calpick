# calpick

An interactive, dependency-free terminal date picker for Go CLI tools.

`calpick` replaces commands such as `mytool 2026-01-01` with a clickable
calendar. It has two modes:

- single date: click a date, or highlight it with the keyboard and press Enter
- date range: click From, then Until, then press Enter

Click `<` or `>` to change months. Keyboard navigation is available as a
fallback.

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
go run ./cmd/calpick 2026-01-01          # start at a specific date
go run ./cmd/calpick --range             # From = Until = today
go run ./cmd/calpick --range 2026-11-01 2026-11-30
```

The single-date mode prints the selected date as `YYYY-MM-DD`. Range mode
prints `from until` on one line, for example `2026-11-01 2026-11-30`.

The interface is drawn on the controlling terminal, so redirecting stdout
does not mix escape sequences into the result.

## Range mode

Range mode uses one calendar and shows the selected range below it.

```text
<        2026 Nov         >
Sun Mon Tue Wed Thu Fri Sat
  1   2   3   4   5   6   7
  ...
2026-11-15 -> 2026-11-20
```

- The first click sets From, the second sets Until, and a third click starts a
  new range. Space does the same from the keyboard.
- If Until is before From, the two dates are swapped.
- The status line shows the selected range, or `YYYY-MM-DD -> --` while waiting for Until.
- Enter confirms. If Until has not been selected, the cursor date becomes Until.
- From and Until use a blue background with white text; dates between them use
  a lighter background. Today is underlined, and the cursor is reverse video
  (bold when it sits on a selected day).
- Clicking `<` or `>` changes the month without changing the selected dates.

## Keys

| Key | Action |
| --- | --- |
| Left / `h` | Previous day |
| Right / `l` | Next day |
| Up / `k` | Previous week |
| Down / `j` | Next week |
| PageUp / PageDown | Previous / next month |
| Space | Select a range endpoint (range mode) |
| Enter | Select or confirm |
| `q` / Esc / Ctrl-C | Cancel |

## Library

Single date:

```go
import "github.com/okonomipizza/calpick/datepicker"

picked, err := datepicker.Pick(datepicker.WithInitial(time.Now()))
if err != nil {
    // datepicker.ErrCanceled means the user cancelled.
}
fmt.Println(picked.Format("2006-01-02"))
```

Date range:

```go
from, until, err := datepicker.PickRange(
    datepicker.WithInitialRange(time.Now(), time.Time{}),
)
if err != nil {
    // datepicker.ErrCanceled means the user cancelled.
}
```

`Pick` and `PickRange` block until the user confirms, cancels, or an error
occurs. They draw on the controlling terminal and restore it before returning.
`WithInitialRange` accepts a zero Until (defaults to the same day as From) and
sorts a reversed range. Neither function is safe for concurrent use.

The package uses only the Go standard library. It contains no cgo code, so
`GOOS`/`GOARCH` cross-compilation works without a C toolchain.

The module path is `github.com/okonomipizza/calpick`. Install the command with:

```sh
go install github.com/okonomipizza/calpick/cmd/calpick@latest
```

## Platforms

- Linux and macOS: raw mode and terminal size use `syscall` ioctls.
- Windows: console modes and mouse input use `ReadConsoleInputW`; Windows
  Terminal or conhost with virtual terminal processing is required. Ctrl-Break
  and console close events are handled as cancellation where the OS allows it.
- tmux and screen: enable mouse reporting (`set -g mouse on`) for clicks.

Keyboard navigation works without mouse support. A single calendar needs at
least 28 columns and 8 rows. Range mode adds one status row, so it needs 28
columns and 7-9 rows depending on the displayed month (9 is the safe minimum).
If a month grows too tall, the picker shows the required size; PageUp/PageDown
still work so a shorter month can be selected.
