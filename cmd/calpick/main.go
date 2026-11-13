// Command calpick is an interactive terminal date picker.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/okonomipizza/calpick/datepicker"
)

const (
	dateLayout = "2006-01-02"
	usage      = `usage: calpick [--range] [YYYY-MM-DD] [YYYY-MM-DD]

Interactive terminal date picker.

  --range                 select a From/Until range instead of one date
  YYYY-MM-DD              initial date (single mode)
  YYYY-MM-DD YYYY-MM-DD   initial From and Until (range mode)

Single mode prints the selected date as YYYY-MM-DD. Range mode prints
"from until" on one line.

  Click a date to select it. In range mode the first click sets From, the
  second sets Until, and a third click starts a new range.
  Click < or > to change months.
  Keys: arrows move the cursor, PgUp/PgDn change months,
        Space selects a range endpoint, Enter selects or confirms,
        q/Esc/Ctrl-C cancels.
`
)

type cliArgs struct {
	rangeMode bool
	help      bool
	from      time.Time
	until     time.Time
	hasFrom   bool
}

func main() {
	args, err := parseArguments(os.Args[1:])
	if err != nil {
		fail("%v", err)
	}
	if args.help {
		fmt.Print(usage)
		return
	}
	if args.rangeMode {
		runRange(args)
		return
	}
	runSingle(args)
}

func runSingle(args cliArgs) {
	var opts []datepicker.Option
	if args.hasFrom {
		opts = append(opts, datepicker.WithInitial(args.from))
	}

	picked, err := datepicker.Pick(opts...)
	handleErr(err)
	fmt.Println(picked.Format(dateLayout))
}

func runRange(args cliArgs) {
	var opts []datepicker.Option
	if args.hasFrom {
		opts = append(opts, datepicker.WithInitialRange(args.from, args.until))
	}

	from, until, err := datepicker.PickRange(opts...)
	handleErr(err)
	fmt.Printf("%s %s\n", from.Format(dateLayout), until.Format(dateLayout))
}

// parseArguments is kept free of I/O so the CLI contract can be unit tested.
func parseArguments(args []string) (cliArgs, error) {
	var parsed cliArgs
	var dates []string

	for _, arg := range args {
		switch arg {
		case "--range":
			parsed.rangeMode = true
		case "-h", "--help":
			parsed.help = true
		default:
			if strings.HasPrefix(arg, "-") && arg != "-" {
				return cliArgs{}, fmt.Errorf("unknown option %q", arg)
			}
			dates = append(dates, arg)
		}
	}
	if parsed.help {
		return parsed, nil
	}

	if parsed.rangeMode {
		switch len(dates) {
		case 0:
			return parsed, nil
		case 2:
			from, err := parseDate(dates[0])
			if err != nil {
				return cliArgs{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", dates[0])
			}
			until, err := parseDate(dates[1])
			if err != nil {
				return cliArgs{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", dates[1])
			}
			parsed.from, parsed.until = from, until
			parsed.hasFrom = true
			return parsed, nil
		default:
			return cliArgs{}, errors.New("range mode takes zero or two dates")
		}
	}

	switch len(dates) {
	case 0:
		return parsed, nil
	case 1:
		from, err := parseDate(dates[0])
		if err != nil {
			return cliArgs{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", dates[0])
		}
		parsed.from = from
		parsed.hasFrom = true
		return parsed, nil
	default:
		return cliArgs{}, errors.New("too many arguments")
	}
}

func parseDate(value string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, value, time.Local)
}

func handleErr(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, datepicker.ErrCanceled) {
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "calpick:", err)
	os.Exit(1)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "calpick: "+format+"\n", args...)
	os.Exit(1)
}
