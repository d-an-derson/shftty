# Comprehensive ANSI/VT100 Escape Sequence Reference

This document outlines the control characters and escape sequences required for a modern, xterm-compatible terminal emulator.

## 1. C0 Control Characters
Single-byte non-printable characters used for basic terminal control.

| Hex / Code | Name | Description | Status |
| :--- | :--- | :--- | :---|
| `\x00` or `^@` | NUL (Null) | Ignored by the terminal; used for padding. | |
| `\x07` or `^G` | BEL (Bell) | Triggers a terminal bell or system alert. | |
| `\x08` or `^H` | BS (Backspace) | Moves the cursor left by one column. | |
| `\x09` or `^I` | HT (Horizontal Tab) | Moves the cursor to the next tab stop (usually every 8 columns). | |
| `\x0A` or `^J` | LF (Line Feed) | Moves the cursor down one row. | |
| `\x0D` or `^M` | CR (Carriage Return) | Moves the cursor to column 1 of the current row. | |
| `\x1B` or `^[` | ESC (Escape) | Initiates an ANSI escape sequence. | |

---

## 2. CSI: Cursor Positioning
Control Sequence Introducer (CSI) sequences begin with `ESC [` (or `\x9b`). Parameters (`<n>`, `<r>`, `<c>`) default to `1` if omitted.

| Sequence | Name | Description |
| :--- | :--- | :--- |
| `ESC [ <n> A` | CUU (Cursor Up) | Moves cursor up by `<n>` rows. |
| `ESC [ <n> B` | CUD (Cursor Down) | Moves cursor down by `<n>` rows. |
| `ESC [ <n> C` | CUF (Cursor Forward) | Moves cursor right by `<n>` columns. |
| `ESC [ <n> D` | CUB (Cursor Back) | Moves cursor left by `<n>` columns. |
| `ESC [ <n> E` | CNL (Cursor Next Line) | Moves cursor to the beginning of the line `<n>` rows down. |
| `ESC [ <n> F` | CPL (Cursor Previous Line) | Moves cursor to the beginning of the line `<n>` rows up. |
| `ESC [ <n> G` | CHA (Cursor Horizontal Absolute) | Moves cursor to column `<n>` on the current row. |
| `ESC [ <r> ; <c> H` | CUP (Cursor Position) | Moves cursor to absolute row `<r>` and column `<c>`. |
| `ESC [ <r> ; <c> f` | HVP (Horizontal/Vertical Position) | Same as CUP; moves cursor to row `<r>` and column `<c>`. |
| `ESC 7` | DECSC (Save Cursor) | Saves the current cursor position and attributes. |
| `ESC 8` | DECRC (Restore Cursor) | Restores the saved cursor position and attributes. |

---

## 3. CSI: Editing and Display Management
Commands for erasing text and inserting/deleting lines.

| Sequence | Name | Description |
| :--- | :--- | :--- |
| `ESC [ <n> J` | ED (Erase in Display) | `0` = cursor to end, `1` = start to cursor, `2` = entire screen, `3` = clear scrollback. |
| `ESC [ <n> K` | EL (Erase in Line) | `0` = cursor to end of line, `1` = start of line to cursor, `2` = entire line. |
| `ESC [ <n> L` | IL (Insert Lines) | Inserts `<n>` blank lines at cursor row, shifting text down. |
| `ESC [ <n> M` | DL (Delete Lines) | Deletes `<n>` lines at cursor row, shifting text up. |
| `ESC [ <n> P` | DCH (Delete Characters) | Deletes `<n>` characters at cursor position, shifting text left. |
| `ESC [ <n> X` | ECH (Erase Characters) | Erases `<n>` characters from the cursor rightward (replaces with spaces). |
| `ESC [ <t> ; <b> r` | DECSTBM (Margins) | Sets scrolling margins from top row `<t>` to bottom row `<b>`. |

---

## 4. CSI: Select Graphic Rendition (SGR)
Controls text styling and colors. SGR sequences end with `m`. Parameters can be chained with semicolons (e.g., `ESC [ 1;31m`).

| Sequence | Name | Description |
| :--- | :--- | :--- |
| `ESC [ 0 m` | Reset | Resets all text formatting and colors to default. |
| `ESC [ 1 m` | Bold | Applies bold or increased intensity text. |
| `ESC [ 2 m` | Dim | Decreases text intensity (faint). |
| `ESC [ 3 m` | Italic | Applies italicized text. |
| `ESC [ 4 m` | Underline | Applies a single underline. |
| `ESC [ 5 m` | Blink | Applies blinking text. |
| `ESC [ 7 m` | Inverse | Swaps foreground and background colors. |
| `ESC [ 8 m` | Hidden | Hides text (invisible). |
| `ESC [ 9 m` | Strikethrough | Draws a line through the text. |
| `ESC [ 22-29 m` | Reset Specific | `22` resets bold/dim, `24` resets underline, `27` resets inverse, etc. |
| `ESC [ 30-37 m` | Foreground (Standard) | Sets foreground color (Black, Red, Green, Yellow, Blue, Magenta, Cyan, White). |
| `ESC [ 40-47 m` | Background (Standard) | Sets background color (Black, Red, Green, Yellow, Blue, Magenta, Cyan, White). |
| `ESC [ 90-97 m` | Foreground (Bright) | Sets bright foreground color variants. |
| `ESC [ 100-107 m` | Background (Bright) | Sets bright background color variants. |
| `ESC [ 38 ; 5 ; <n> m` | Foreground (256) | Sets foreground to a 256-color palette index (`<n> = 0-255`). |
| `ESC [ 48 ; 5 ; <n> m` | Background (256) | Sets background to a 256-color palette index (`<n> = 0-255`). |
| `ESC [ 38 ; 2 ; <r>;<g>;<b> m`| Foreground (RGB) | Sets foreground to 24-bit truecolor (RGB). |
| `ESC [ 48 ; 2 ; <r>;<g>;<b> m`| Background (RGB) | Sets background to 24-bit truecolor (RGB). |

---

## 5. DEC Private Modes (Terminal State)
Toggles for terminal behaviors, using `h` (high/enable) and `l` (low/disable).

| Sequence | Name | Description |
| :--- | :--- | :--- |
| `ESC [ ? 7 h` / `l` | Auto-Wrap Mode | `h` enables wrapping at the right edge; `l` truncates text. |
| `ESC [ ? 25 h` / `l` | Cursor Visibility | `h` shows the cursor; `l` hides the cursor. |
| `ESC [ ? 1049 h` / `l`| Alternate Screen | `h` saves cursor and switches to alt screen; `l` restores cursor and main screen. |
| `ESC [ ? 2004 h` / `l`| Bracketed Paste | `h` wraps pasted text in `ESC [ 200 ~` (start) and `ESC [ 201 ~` (end). |

---

## 6. Mouse Tracking Modes
When enabled via DEC Private Modes, the terminal reports mouse events to the application as keyboard input. Modern applications typically use SGR 1006 encoding for coordinates beyond 223.

| Sequence (Enable/Disable) | Name | Description |
| :--- | :--- | :--- |
| `ESC [ ? 1000 h` / `l` | Normal Mouse Tracking | Reports on button press and release only. |
| `ESC [ ? 1002 h` / `l` | Button-Event Tracking | Reports button presses, releases, and drag events. |
| `ESC [ ? 1003 h` / `l` | Any-Event Tracking | Reports all mouse movements, presses, and releases. |
| `ESC [ ? 1006 h` / `l` | SGR Mouse Encoding | Formats reports as `ESC [ <button> ; <x> ; <y> M` (press) or `m` (release). |

---

## 7. Device Status and Reports
Bidirectional requests where the terminal must "type" the answer back to the application.

| Request Sequence | Name | Required Terminal Response |
| :--- | :--- | :--- |
| `ESC [ 5 n` | DSR (Device Status Report) | `ESC [ 0 n` (Terminal is OK). |
| `ESC [ 6 n` | CPR (Cursor Position Report) | `ESC [ <r> ; <c> R` (Where `<r>` and `<c>` are current coordinates). |
| `ESC [ c` | Primary DA (Device Attributes) | `ESC [ ? 6 c` (Mimics standard VT102 capabilities). |
| `ESC [ > c` | Secondary DA | `ESC [ > 0 ; 276 ; 0 c` (Mimics xterm version). |

---

## 8. Operating System Commands (OSC)
OSC sequences begin with `ESC ]` and are terminated by either BEL (`\x07`) or the String Terminator `ST` (`ESC \`).

| Sequence | Name | Description |
| :--- | :--- | :--- |
| `ESC ] 0 ; <string> BEL` | Window Title | Sets the terminal window and icon title to `<string>`. |
| `ESC ] 2 ; <string> BEL` | Window Title (Strict) | Sets the terminal window title (but not icon) to `<string>`. |
| `ESC ] 4 ; <c> ; ? BEL` | Color Palette Query | Asks for the RGB value of color index `<c>`. |
| `ESC ] 11 ; ? BEL` | Background Query | Asks for background color. Reply: `ESC ] 11 ; rgb:<r>/<g>/<b> BEL`. |
| `ESC ] 52 ; <c> ; <b> BEL`| Clipboard Integration | Allows host to read/write system clipboard (base64 encoded). |

---

## 9. Character Set Selection
Used heavily by older applications to draw UI borders before Unicode was standard.

| Sequence | Name | Description |
| :--- | :--- | :--- |
| `ESC ( B` | Designate G0 to US-ASCII | Switches terminal to standard text mode. |
| `ESC ( 0` | Designate G0 to Line Drawing | Switches terminal to the VT100 special graphics character set. |
| `\x0E` or `^N` | SO (Shift Out) | Invokes the G1 character set. |
| `\x0F` or `^O` | SI (Shift In) | Invokes the G0 character set (default). |
