# Kindle Higlights

## How it works

Providing the correct ammount of params it will get all notes from `.txt` file.
Setup to be stored the output of the note parsing on `mac.env` or `linux.env` file destination env variable - for now.

### Setup
1. Build it with make `go build .` or via `make build-solo`
2. Run it with `<build-name> $(ARGS)`
   - **One Arg**:
     - `help`: helper arg - output the possible run args
   - **One Arg**:
     - `<title-name>`: title to search FileLocation path default
   - **Two Args**:
     - `test`: to setup test FileLocation path
     - `<title-name>`: title to search
   - **Two Args**:
     - `<title-name>`: title to search
     - `<file-location>`: absolute path to the file
