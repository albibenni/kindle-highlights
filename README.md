# Kindle Highlights Parser

A CLI tool to parse and organize your Kindle highlights from `My Clippings.txt`.

## Features

- Extracts highlights by book title.
- Export clean notes to a specified directory (e.g., your Obsidian vault or Second Brain).
- Supports local and global configuration.

## Installation

### Prerequisites

- [Go](https://go.dev/dl/) installed (version 1.24+ recommended).
- `make` utility.

### Steps

1. **Clone the repository:**

   ```bash
   git clone <repo-url>
   cd kindle-highlights
   ```

2. **Setup Global Configuration:**
   This command creates the config directory at `~/.config/kindle-highlights/` and copies the default `mac.env` there.

   ```bash
   make setup-config
   ```

3. **Install Pre-commit Hooks (Optional but recommended):**
   This sets up Git hooks to run linting and tests automatically before each commit.

   ```bash
   make install-hooks
   ```

4. **Install the Binary:**
   Compiles the application as `kindle-parser` and installs it to `~/go/bin`.

   ```bash
   make install
   ```

5. **Update PATH (if necessary):**
   Ensure your Go binary directory is in your system PATH.

   ```bash
   export PATH=$HOME/go/bin:$PATH
   ```

## Configuration

The application looks for configuration in the following order:

1. **Local:** `mac.env` (or `linux.env`) in the current directory. This takes precedence and is useful for development.
2. **Global:** `~/.config/kindle-highlights/.env`. This is used when running `kindle-parser` from any other directory.

### Environment Variables

Edit your `.env` file (local or global) to set your paths:

```env
# Where your parsed notes will be saved (e.g., Obsidian vault)
NOTE_PATH="/Users/yourname/Documents/Notes/Books/"

# Path to your Kindle clippings file
CLIPPING_PATH="/Users/yourname/Downloads/My Clippings.txt"
```

## Usage

Once installed, you can run the tool globally using `kindle-parser`.

### Command Help

Show available commands and usage options.

```bash
kindle-parser help
```

### Parse a Book

Search for a specific book title in your configured `CLIPPING_PATH`.

```bash
kindle-parser "The clean coder"
```

### Test Mode

Use the internal test file (`test-file/My Clippings.txt`) for dry runs.

```bash
kindle-parser test "Book Title"
```

### Custom File Path

Override the default clippings path for a single run.

```bash
kindle-parser "Book Title" "/path/to/other/clippings.txt"
```

## Development

Use the included `Makefile` for common tasks:

- `make build`: Build the binary locally.
- `make build-solo`: Build without args (basic).
- `make test`: Run unit tests.
- `make lint`: Run golangci-lint.
- `make deps`: Download Go modules.
