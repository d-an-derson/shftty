```

     ____   _   _  _____  _____  _____ __   __
    / ___| |_| |_||  ___||_   _||_   _|\ \ / /
    \___ \ |  _  || |_     | |    | |   \ V / 
     ___) || | | ||  _|    | |    | |    | |  
    |____/ |_| |_||_|      |_|    |_|    |_|  
      Sh*tty F*cking Terminal Emulator(tty)                                         

```                                          


![Shftty screenshot 1](./assets/shffty-main.png)


# What it do
Shffty is VT100 compliant* terminal emulator written in Go using the Fyne GUI library. 
Shftty aims to be a fast(it's not), functional(getting there) tty that works well enough for me to not feel dumb about it.

* It's not currently VT100 compliant

## Getting started

### Prerequisites
To build from source your system must have:
Go 1.20+
A C compiler(gcc/clang) and the Fyne dependencies for your build target. 
The [Fyne - Getting Started](https://docs.fyne.io/started/quick/) can walk you through the required dependencies.

```
# Clone the repository
git clone https://github.com/d-an-derson/shftty.git
cd shftty

# Install dependencies
go mod tidy

# Run from the terminal
go run .

# Build from source
Fyne allows you to build directly for target environments. Replace 'darwin' with your target as needed:

fyne package -src ./cmd/. -os target -icon ../assets/shfte-icon-1024x1024.png

```

### Supported build targets

| Platform | Command | Architectures | Packaging |
|---|---|---|---|
| Linux | linux | amd64, 386, arm, arm64 | .tar.xz |
| Windows | windows | amd64, 386, arm64 | .zip, .appx (release) |
| macOS | darwin | amd64, arm64 | .app bundle, .dmg |
| Android | android | arm64-v8a, armeabi-v7a, x86, x86_64 | .apk, .aab |
| iOS | ios | arm64 | .ipa |
| FreeBSD | freebsd | amd64, arm64 | .tar.xz |
| Web | web | wasm | Web directory |

Linux users:
*the wayland build tag compiles the program for the Wayland window protocol instead of X11.*
*eg:*
``` go run -tags wayland cmd/.```

### Releases
No builds are currently available for Shftty. As development progresses this may change, but currently if you want it, you gotta build it.

## What Works
- 256 color support (kinda)
- Essential VT100 escape sequences (kinda)
- Custom fonts (kinda)
- NeoVim (kinda)

## What doesn't work
- 256 color support
- Essential VT100 escape sequences
- NeoVim
- Mouse support
- Kitty Graphics Protocol
- Python REPL
- Claude/AGY (probably, not tested but they use React to render their TUI and I can't imagine that shit works here)
- Everything else...

### TODO 
- Implement the rest of the ANSI escape codes
- Scrollback buffer
- Kitty graphics (probably never gonna happen)
- Fix everything else
