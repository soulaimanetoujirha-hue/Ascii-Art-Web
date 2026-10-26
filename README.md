# ascii-art-web

## Description

ascii-art-web is a web version of the ascii-art project. A Go HTTP server shows a page where you type some text, pick a banner (`standard`, `shadow` or `thinkertoy`) and get the text drawn as ASCII art. A **Download .txt** button saves the result as a text file.

New to the code? Read [LEARN.md](LEARN.md): it explains the idea, every function and every package.

## Authors

- Soulaimane Toujirha

## Usage: how to run

Requires Go 1.27 or newer.

```bash
go run .
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

Run the tests:

```bash
go test ./...
```

### Endpoints

| Method | Path         | Description                                   |
|--------|--------------|-----------------------------------------------|
| GET    | `/`          | Main page with the form                       |
| POST   | `/ascii-art` | Takes `text` and `banner`, shows the result   |

### Status codes

- `200 OK`: the page or the result was rendered
- `400 Bad Request`: empty text, text too long, unknown banner or non-ASCII characters
- `404 Not Found`: unknown page, or a missing template or banner file
- `405 Method Not Allowed`: wrong HTTP method on an endpoint
- `500 Internal Server Error`: any other unexpected error

## Implementation details: algorithm

1. **Banner files**: each file in `banners/` holds the 95 printable ASCII characters, from space (32) to `~` (126). Every character is 8 lines tall and followed by one empty line, so a character takes 9 lines. The file starts with one empty line.
2. **Loading a font**: `loadFont` walks the file and stores each character's 8 lines in a map, so `font['A'][r]` is row `r` of `A`. Character `c` starts at line `1 + (c - 32) * 9`.
3. **Drawing**: the input is split on newlines. For each input line, the program builds 8 output rows. For each row it appends that row of every character in the line, side by side. An empty input line becomes one empty output line.
4. **Validation**: browser `\r\n` newlines are turned into `\n`, and any character outside 32–126 is rejected with `400 Bad Request`. The banner name must be one of the three allowed names, so no other file can be read.
5. **Server**: `net/http` serves the routes and `html/template` renders `templates/index.html`. The page is rendered into a buffer first, so that a template error can still return a clean `500`.
6. **Download**: `static/script.js` puts the text of the result `<pre>` in a `Blob` and saves it as `ascii-art.txt`, all in the browser.

### Project structure

```
ascii-art-web/
├── main.go          # starts the server and registers the routes
├── handlers.go      # GET / and POST /ascii-art handlers
├── ascii/
│   ├── ascii.go       # ASCII art rendering
│   └── ascii_test.go  # tests
├── banners/         # standard, shadow and thinkertoy fonts
├── templates/       # HTML templates
├── static/          # design.css and script.js (Download button)
└── LEARN.md         # learning guide
```
