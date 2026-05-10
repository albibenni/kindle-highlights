# Kindle Highlights Parser

[![CI](https://github.com/albibenni/kindle-highlights/actions/workflows/ci.yml/badge.svg)](https://github.com/albibenni/kindle-highlights/actions/workflows/ci.yml)

A interactive TUI tool to parse and organize your Kindle highlights from `My Clippings.txt`.

## Features

- **Interactive TUI:** Easily navigate your clipping sources and book list.
- **Smart Search:** Search for your clippings file directly within the app.
- **Clean Export:** Extracts highlights by book title and exports them as clean Markdown files to your specified directory (e.g., Obsidian vault).
- **Automated Organization:** Automatically creates folders and filename patterns based on Title and Author.
- **Local & Global Config:** Supports flexible environment-based configuration.

## Installation

### Prerequisites

- [Go](https://go.dev/dl/) installed (version 1.24+ recommended).
- `make` utility.
- [ripgrep](https://github.com/BurntSushi/ripgrep) (`rg`) for the system-wide clippings search feature.

### Steps

1. **Clone the repository:**

   ```bash
   git clone <repo-url>
   cd kindle-highlights
   ```

2. **Interactive Setup (Recommended):**
   This script checks dependencies, installs the app, and lets you set a custom alias.

   ```bash
   make setup
   ```

3. **Setup Global Configuration:**
   This command creates the config directory at `~/.config/kindle-highlights/` and copies a template `.env` there.

   ```bash
   make setup-config
   ```

4. **Install Pre-commit Hooks (Recommended for developers):**
   Sets up Git hooks to run linting and tests automatically before each commit.

   ```bash
   make install-hooks
   ```

5. **Install the Binary:**
   This uses `go install` to compile and place the binary in your Go bin directory (`$(go env GOPATH)/bin`).

   ```bash
   make install
   ```

6. **Update PATH (if necessary):**
   If you can't run `kindle-parser` immediately, ensure your Go bin directory is in your system PATH.

   ```bash
   export PATH=$(go env GOPATH)/bin:$PATH
   ```

## Configuration

The application looks for configuration in the following order:

1. **Local:** `mac.env` (or `linux.env`) in the current directory.
2. **Global:** `~/.config/kindle-highlights/.env`.

### Environment Variables

Edit your `.env` file to set your paths:

```env
# Where your parsed notes will be saved (e.g., Obsidian vault)
NOTE_PATH="/Users/yourname/Documents/Notes/Books/"

# Path to your Kindle clippings file
CLIPPING_PATH="/Users/yourname/Downloads/My Clippings.txt"
```

## Usage

Once installed, run the tool globally:

```bash
kindle-parser
```

### Navigation

- **Source Selection:** Choose between your configured `Default Path` or a `Custom Path`.
- **Custom Path Search:** If you choose Custom Path, you can start typing to search for `.txt` files on your system (requires `ripgrep`).
    - **Smart Case:** The search is case-insensitive if you type in all lowercase, and case-sensitive if you include any uppercase letters.
- **Book Selection:** Browse the list of books found in your clippings file.
- **Export:** Select a book and press `Enter` to parse and export all highlights to your `NOTE_PATH`.

## Development

Use the included `Makefile` for common tasks:

- `make run`: Run the application directly.
- `make test`: Run unit tests.
- `make lint`: Run `golangci-lint` (requires `golangci-lint` to be installed).
- `make coverage`: Run tests and show code coverage.
- `make deps`: Download Go modules.

---
Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) 🫧
