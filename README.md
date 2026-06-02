# Image Viewer

A lightweight, frameless image viewer built with Go and Ebiten. Supports common image formats with zoom, pan, rectangular selection, cropping, and clipboard operations.

## Features

- **Frameless window** — Clean, minimal interface without OS window decorations
- **Universal image support** — PNG, JPEG, GIF, BMP, WebP via Go's standard and extended image packages
- **Mouse wheel zoom** — Scroll up/down to scale the image smoothly
- **Pan** — Middle-click drag to move the image around
- **Rectangular selection** — Left-click drag to draw a selection rectangle
- **Context menu** — Right-click inside a selection for options:
  - *Crop to selection* — replaces the preview with the cropped area
  - *Copy to clipboard* — copies the selected region as PNG to the system clipboard
  - *Save image* — saves the current preview (`Ctrl+S` also works)
- **Window controls** — Hover the top-right corner to reveal minimize, maximize, and close buttons with crisp icon glyphs
- **Cross-platform** — Linux, macOS, Windows

## Installation

### Prebuilt binaries

Download the latest release for your platform from the [Releases](../../releases) page.

### Build from source

Requirements:
- Go 1.22 or later
- Platform-specific graphics/audio libraries (for Ebiten):
  - **Linux:** `libgl1-mesa-dev`, `xorg-dev`, `libasound2-dev`
  - **macOS:** Xcode Command Line Tools
  - **Windows:** MinGW-w64 (for CGO)

```bash
git clone https://github.com/yourusername/imageviewer.git
cd imageviewer
go build -o imageviewer .
```

## Usage

Open an image from the command line:

```bash
./imageviewer photo.png
```

If no file is provided, the window opens and displays a hint to drag an image or run with a file path.

### Controls

| Action | Input |
|--------|-------|
| Zoom | Mouse wheel |
| Pan image | Middle-click drag |
| Select region | Left-click drag on image |
| Open context menu | Right-click inside selection |
| Crop selection | Context menu → *Crop to selection* |
| Copy selection | Context menu → *Copy to clipboard* |
| Save preview | Context menu → *Save image* or `Ctrl+S` |
| Move window | Left-drag top bar |
| Window actions | Hover top-right corner, click icons |

## Development

### Running tests

```bash
go test ./...
```

### Pre-commit hooks

This project uses [prek](https://github.com/prek/prek) for managing Git hooks. After cloning, install the hooks:

```bash
prek install
```

Configured hooks:
- trailing-whitespace
- end-of-file-fixer
- check-yaml
- check-added-large-files
- go-fmt
- go-vet
- go-build
- go-mod-tidy

Run all hooks manually:

```bash
prek run --all-files
```

### CI & Releases

GitHub Actions automatically runs tests and builds on every push and pull request. Pushing a `v*` tag triggers multi-platform releases for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## License

MIT License — see [LICENSE](LICENSE) for details.
