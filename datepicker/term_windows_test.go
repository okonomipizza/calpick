//go:build windows

package datepicker

import (
	"testing"
	"unsafe"
)

func TestWindowsStructLayout(t *testing.T) {
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"inputRecord", unsafe.Sizeof(inputRecord{}), 20},
		{"keyEventRecord", unsafe.Sizeof(keyEventRecord{}), 16},
		{"mouseEventRecord", unsafe.Sizeof(mouseEventRecord{}), 16},
		{"consoleScreenBufferInfo", unsafe.Sizeof(consoleScreenBufferInfo{}), 22},
		{"keyEventRecord.unicodeChar", unsafe.Offsetof(keyEventRecord{}.unicodeChar), 10},
		{"keyEventRecord.controlKeyState", unsafe.Offsetof(keyEventRecord{}.controlKeyState), 12},
		{"mouseEventRecord.eventFlags", unsafe.Offsetof(mouseEventRecord{}.eventFlags), 12},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

func TestParseKeyMappings(t *testing.T) {
	term := &winTerminal{}
	tests := []struct {
		name string
		key  keyEventRecord
		want keyKind
		ok   bool
	}{
		{"left", keyEventRecord{virtualKeyCode: vkLeft}, keyLeft, true},
		{"right", keyEventRecord{virtualKeyCode: vkRight}, keyRight, true},
		{"up", keyEventRecord{virtualKeyCode: vkUp}, keyUp, true},
		{"down", keyEventRecord{virtualKeyCode: vkDown}, keyDown, true},
		{"page up", keyEventRecord{virtualKeyCode: vkPrior}, keyPageUp, true},
		{"page down", keyEventRecord{virtualKeyCode: vkNext}, keyPageDown, true},
		{"enter", keyEventRecord{virtualKeyCode: vkReturn}, keyEnter, true},
		{"escape", keyEventRecord{virtualKeyCode: vkEscape}, keyCancel, true},
		{"q", keyEventRecord{unicodeChar: 'q'}, keyCancel, true},
		{"ctrl-c", keyEventRecord{unicodeChar: 0x03}, keyCancel, true},
		{"j", keyEventRecord{unicodeChar: 'j'}, keyDown, true},
		{"unknown", keyEventRecord{unicodeChar: 'x'}, keyNone, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, ok := term.parseKey(&tt.key)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if !ok {
				return
			}
			if ev.kind != evKey || ev.key != tt.want {
				t.Fatalf("event = %+v, want key %v", ev, tt.want)
			}
		})
	}
}
