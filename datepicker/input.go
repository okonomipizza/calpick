package datepicker

const escByte = 0x1b

type eventKind int

const (
	evIgnore eventKind = iota
	evKey
	evMouse
	evResize
)

type keyKind int

const (
	keyNone keyKind = iota
	keyUp
	keyDown
	keyLeft
	keyRight
	keyEnter
	keyCancel
	keyPageUp
	keyPageDown
	keySelect
)

type event struct {
	kind eventKind
	key  keyKind
	row  int // 1-based mouse row
	col  int // 1-based mouse column
}

func keyEvent(k keyKind) event {
	return event{kind: evKey, key: k}
}

// parseOne consumes one input event from b. The bool result reports that more
// bytes are needed to complete a sequence; in that case consumed is zero.
func parseOne(b []byte) (event, int, bool) {
	if len(b) == 0 {
		return event{}, 0, true
	}
	if b[0] != escByte {
		return parseControl(b)
	}
	return parseEscape(b)
}

func parseControl(b []byte) (event, int, bool) {
	switch b[0] {
	case '\r', '\n':
		return keyEvent(keyEnter), 1, false
	}
	if key, ok := runeKey(rune(b[0])); ok {
		return keyEvent(key), 1, false
	}
	return event{kind: evIgnore}, 1, false
}

// runeKey maps a printable or control character to an action. Both the Unix
// byte parser and the Windows key parser use it so bindings stay in sync.
func runeKey(r rune) (keyKind, bool) {
	switch r {
	case 0x03, 'q', 'Q':
		return keyCancel, true
	case 'h', 'H':
		return keyLeft, true
	case 'l', 'L':
		return keyRight, true
	case 'k', 'K':
		return keyUp, true
	case 'j', 'J':
		return keyDown, true
	case ' ':
		return keySelect, true
	}
	return keyNone, false
}

func parseEscape(b []byte) (event, int, bool) {
	if len(b) == 1 {
		return event{}, 0, true
	}

	switch b[1] {
	case '[':
		return parseCSI(b)
	case 'O':
		if len(b) < 3 {
			return event{}, 0, true
		}
		switch b[2] {
		case 'A':
			return keyEvent(keyUp), 3, false
		case 'B':
			return keyEvent(keyDown), 3, false
		case 'C':
			return keyEvent(keyRight), 3, false
		case 'D':
			return keyEvent(keyLeft), 3, false
		}
		return event{kind: evIgnore}, 3, false
	default:
		// Alt+key: consume both bytes so it is not mistaken for plain key.
		return event{kind: evIgnore}, 2, false
	}
}

func parseCSI(b []byte) (event, int, bool) {
	if len(b) < 3 {
		return event{}, 0, true
	}
	if b[2] == '<' {
		return parseSGRMouse(b)
	}
	if b[2] == 'M' {
		// Legacy X10 mouse report: ESC [ M Cb Cx Cy. Accept only a plain left
		// button press (Cb == 32); motion, wheel, release, and modified clicks
		// are not selections. Consume the coordinate bytes so they cannot be
		// mistaken for keystrokes.
		if len(b) < 6 {
			return event{}, 0, true
		}
		if b[3] == 32 {
			x := int(b[4]) - 32
			y := int(b[5]) - 32
			if x >= 1 && y >= 1 {
				return event{kind: evMouse, row: y, col: x}, 6, false
			}
		}
		return event{kind: evIgnore}, 6, false
	}

	for i := 2; i < len(b); i++ {
		if b[i] >= 0x40 && b[i] <= 0x7e {
			return csiEvent(b, i)
		}
	}
	return event{}, 0, true
}

func csiEvent(b []byte, final int) (event, int, bool) {
	consumed := final + 1
	switch b[final] {
	case 'A':
		return keyEvent(keyUp), consumed, false
	case 'B':
		return keyEvent(keyDown), consumed, false
	case 'C':
		return keyEvent(keyRight), consumed, false
	case 'D':
		return keyEvent(keyLeft), consumed, false
	case '~':
		switch string(b[2:final]) {
		case "5":
			return keyEvent(keyPageUp), consumed, false
		case "6":
			return keyEvent(keyPageDown), consumed, false
		}
	}
	return event{kind: evIgnore}, consumed, false
}

// parseSGRMouse parses "ESC [ < button ; x ; y M" and "m". Only a left-button
// press (button 0, final M) produces an event; releases and other buttons are
// ignored.
func parseSGRMouse(b []byte) (event, int, bool) {
	var nums [3]int
	i := 3

	for n := 0; n < 3; n++ {
		if i >= len(b) {
			return event{}, 0, true
		}
		value := 0
		digits := 0
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			value = value*10 + int(b[i]-'0')
			i++
			digits++
		}
		if digits == 0 {
			// Malformed sequence: drop it.
			return event{kind: evIgnore}, min(i+1, len(b)), false
		}
		nums[n] = value
		if n < 2 {
			if i >= len(b) {
				return event{}, 0, true
			}
			if b[i] != ';' {
				return event{kind: evIgnore}, i + 1, false
			}
			i++
		}
	}

	if i >= len(b) {
		return event{}, 0, true
	}
	final := b[i]
	i++
	if final != 'M' || nums[0] != 0 || nums[1] < 1 || nums[2] < 1 {
		return event{kind: evIgnore}, i, false
	}
	return event{kind: evMouse, row: nums[2], col: nums[1]}, i, false
}
