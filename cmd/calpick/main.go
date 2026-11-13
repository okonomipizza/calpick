// Command calpick is an interactive terminal date picker.
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"calpick/datepicker"
)

const usage = `usage: calpick [YYYY-MM-DD]

Interactive terminal date picker. The initial month is the current month, or
the month containing the optional date argument.

  Click a date to select it. Click < or > to change months.
  Keys: arrows move the cursor, PgUp/PgDn change months,
        Enter selects, q/Esc/Ctrl-C cancels.

The selected date is printed to stdout as YYYY-MM-DD.
`

func main() {
	args := os.Args[1:]
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(usage)
		return
	}
	if len(args) > 1 {
		fail("too many arguments")
	}

	var opts []datepicker.Option
	if len(args) == 1 {
		initial, err := time.ParseInLocation("2006-01-02", args[0], time.Local)
		if err != nil {
			fail("invalid date %q: expected YYYY-MM-DD", args[0])
		}
		opts = append(opts, datepicker.WithInitial(initial))
	}

	picked, err := datepicker.Pick(opts...)
	if err != nil {
		if errors.Is(err, datepicker.ErrCanceled) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "calpick:", err)
		os.Exit(1)
	}
	fmt.Println(picked.Format("2006-01-02"))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "calpick: "+format+"\n", args...)
	os.Exit(1)
}
