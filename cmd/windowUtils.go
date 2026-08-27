package main

import (
	"os"

	// fyne dependencies
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func windowShortcuts(w fyne.Window, p *os.File) {
	// Register Crtl + Key shortcuts and write them to pty
	// TODO: Implement Crtl + A-Z shortcuts
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyA,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		_, _ = p.Write([]byte{'\x01'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyB, 
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		_, _ = p.Write([]byte{'\x02'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyC, // Ctrl+C
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		_, _ = p.Write([]byte{'\x03'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyD, // Crtl+D (EOF)
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+D is byte value 4
		_, _ = p.Write([]byte{'\x04'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyE, // Crtl+E
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+D is byte value 4
		_, _ = p.Write([]byte{'\x05'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyF,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+E is byte value 6
		_, _ = p.Write([]byte{'\x06'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyG,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+G is byte value 7
		_, _ = p.Write([]byte{'\x07'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyH,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+H is byte value 8
		_, _ = p.Write([]byte{'\x08'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyI,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+I is byte value 9
		_, _ = p.Write([]byte{'\x09'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyJ,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+J is byte value A
		_, _ = p.Write([]byte{'\x0A'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyK,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+K is byte value B
		_, _ = p.Write([]byte{'\x0B'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyL,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+L is byte value C
		_, _ = p.Write([]byte{'\x0C'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyM,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+M is byte value D
		_, _ = p.Write([]byte{'\x0D'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyN,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+N is byte value E
		_, _ = p.Write([]byte{'\x0E'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyO,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+O is byte value F
		_, _ = p.Write([]byte{'\x0F'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyP,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+P is byte value 10
		_, _ = p.Write([]byte{'\x10'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyQ,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+Q is byte value 11
		_, _ = p.Write([]byte{'\x11'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyR,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+R is byte value 12
		_, _ = p.Write([]byte{'\x12'})
	})


	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyS,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+S is byte value 13
		_, _ = p.Write([]byte{'\x13'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyT,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+T is byte value 14
		_, _ = p.Write([]byte{'\x14'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyU,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+U is byte value 15
		_, _ = p.Write([]byte{'\x15'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyV,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+V is byte value 16
		_, _ = p.Write([]byte{'\x16'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyW,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+W is byte value 17
		_, _ = p.Write([]byte{'\x17'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyX,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+X is byte value 18
		_, _ = p.Write([]byte{'\x18'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyY,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+Y is byte value 19
		_, _ = p.Write([]byte{'\x19'})
	})

	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyZ,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		// Ctrl+Z is byte value 1A
		_, _ = p.Write([]byte{'\x1A'})
	})

}
