# Learning guide: ascii-art-web

This file explains the whole project in simple words: the idea, what you should learn,
the job of every file and every function, and every package we use.
Read it from top to bottom once, then keep it open while you read the code.

---

## 1. The idea in one minute

You already built `ascii-art`: a program that takes text and prints it with big letters
made of characters. **ascii-art-web** does the same thing, but in a web page.

1. A **Go server** runs on your computer (`go run .`).
2. You open `http://localhost:8080` in the browser. The server sends back an HTML page with a form.
3. You type text, choose a banner (`standard`, `shadow` or `thinkertoy`) and click **Generate**.
4. The browser sends the form to the server (`POST /ascii-art`).
5. The server draws the ASCII art and sends the page back, with the art inside.
6. You click **Download .txt**: a small JavaScript file saves the art as `ascii-art.txt`.

That is all. The rest is details.

---

## 2. The journey of one request

```
  BROWSER                                        GO SERVER (localhost:8080)
  -------                                        --------------------------
  GET /  ------------------------------------->  homeHandler
                                                   └─ render()  fills templates/index.html
         <------------------------------------  200 OK + HTML page (empty form)

  (you type "Hi", pick "standard", click Generate)

  POST /ascii-art  text=Hi&banner=standard --->  asciiArtHandler
                                                   ├─ checkInput()   wrong form?  -> 400
                                                   ├─ ascii.Render() draws the art
                                                   │    └─ loadFont() reads banners/standard.txt
                                                   └─ render()       puts the art in the page
         <------------------------------------  200 OK + HTML page with the art in <pre>

  (you click "Download .txt")
  static/script.js saves the text of <pre> as ascii-art.txt
  -> no message is sent to the server for this part
```

The browser also asks for `/static/design.css` and `/static/script.js`.
`main.go` tells the server to send those files from the `static` folder.

---

## 3. What you should learn from this project

Use this as a checklist. You understand the project when you can explain each line.

**The web**
- [ ] What a **server** and a **client** (the browser) are.
- [ ] What an **HTTP request** is: a method (`GET`, `POST`), a path (`/ascii-art`) and sometimes a body.
- [ ] The difference between **GET** (ask for a page) and **POST** (send data, like a form).
- [ ] What an **HTTP response** is: a status code, headers and a body.
- [ ] The **status codes** we use: `200`, `400`, `404`, `405`, `500` (see section 7).
- [ ] How an **HTML form** sends data: `method="POST"`, `action="/ascii-art"`, and the `name` of each input.

**Go**
- [ ] How a Go **module** and **packages** work (`go.mod`, `package main`, `package ascii`).
- [ ] Why names that start with a **capital letter** are exported (`Render`) and small ones are not (`loadFont`).
- [ ] **Handlers**: functions with `(w http.ResponseWriter, r *http.Request)`.
- [ ] **Templates**: how `{{.Text}}` puts Go data inside HTML.
- [ ] **Errors**: returning `error`, wrapping with `%w`, and checking with `errors.Is`.
- [ ] **Runes**: a `rune` is one character. `' '` is the number 32, `'~'` is 126.
- [ ] **Maps** (`map[rune][]string`) and **slices** (`lines[1:9]`).
- [ ] **Reading a file** line by line with `bufio.Scanner`, and `defer file.Close()`.
- [ ] Building a long string with `strings.Builder`.
- [ ] **Tests** with the `testing` package (`go test ./...`).

**The browser side**
- [ ] How CSS styles the page.
- [ ] How JavaScript finds an element (`getElementById`) and reacts to a click (`addEventListener`).
- [ ] How a `Blob` turns text into a file the browser can download.

**Safety**
- [ ] Why we only accept 3 banner names (so nobody can read other files on the server).
- [ ] Why `html/template` is safer than `text/template` (it escapes HTML, see section 6).

---

## 4. The banner files and the formula

A banner file (for example `banners/standard.txt`) holds the drawing of the 95 printable
characters, from space (32) to `~` (126), in order.

- The file **starts with 1 empty line**.
- Each character is **8 lines** tall.
- After each character there is **1 empty line**.
- So each character takes **9 lines**.

```
line 0      (empty)
line 1-8    ' '  (space, 32)
line 9      (empty)
line 10-17  '!'  (33)
line 18     (empty)
line 19-26  '"'  (34)
...
```

So character `c` starts at line:

```
start = 1 + (c - 32) * 9
```

**Example:** `'A'` is 65. `1 + (65 - 32) * 9 = 1 + 33 * 9 = 298`.
Lines 298 to 305 draw the `A` (your editor counts from 1, so it shows them as lines 299 to 306).

In the code, `loadFont` does this walk for you: it starts at line 1, takes 8 lines for `' '`,
jumps 9 lines, takes 8 lines for `'!'`, and so on. It stores them in a map, so later
`font['A']` gives the 8 lines of `A`, and `font['A'][0]` gives its first line.

**How a word is drawn.** Your screen prints line by line, so we cannot draw the letters one
after the other. We draw **row by row**: row 0 of every letter, then a newline, then row 1 of
every letter, and so on until row 7.

```
"Hi"   row 0:  font['H'][0] + font['i'][0]   ->   " _    _   _  "
       row 1:  font['H'][1] + font['i'][1]   ->   "| |  | | (_) "
       ...
       row 7:  font['H'][7] + font['i'][7]
```

If the text has several lines (you pressed Enter), each line is drawn like this, one under
the other. An empty line becomes one empty line in the art.

---

## 5. Every file and every function

### Project structure

```
ascii-art-web/
├── go.mod               # the module name and the Go version, no outside packages
├── main.go              # starts the server and connects paths to handlers
├── handlers.go          # what happens on GET / and POST /ascii-art
├── ascii/
│   ├── ascii.go         # draws the ASCII art (no web code here)
│   └── ascii_test.go    # tests for ascii.go
├── banners/             # the 3 fonts: standard, shadow, thinkertoy
├── templates/
│   └── index.html       # the HTML page, with {{ }} holes Go fills in
└── static/
    ├── design.css       # the look of the page
    └── script.js        # the Download button
```

Why is `ascii` its own package? Because drawing art has nothing to do with the web.
You could reuse `ascii.Render` in a terminal program without changing it.
Keeping things separate makes each file small and easy to test.

---

### `main.go`

#### `func main()`
The program starts here. It does 4 things:

| Line | What it does |
|------|--------------|
| `http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))` | Serves files from the `static` folder. For `/static/script.js`, `StripPrefix` removes `/static/`, and `FileServer` sends the file `static/script.js`. |
| `http.HandleFunc("/", homeHandler)` | Every request whose path is not matched by another route goes to `homeHandler`. |
| `http.HandleFunc("/ascii-art", asciiArtHandler)` | Requests to `/ascii-art` go to `asciiArtHandler`. |
| `http.ListenAndServe(port, nil)` | Starts the server on port 8080 and waits for requests forever. `nil` means "use the routes I registered above". If it fails (for example the port is busy), `log.Fatal` prints the error and stops the program. |

---

### `handlers.go`

#### `const maxInputLength = 1000`
The longest text we accept. It stops people from sending a huge text that would take a lot of work to draw.

#### `var banners = map[string]bool{...}`
The **allow-list** of banner names. `banners["shadow"]` is `true`, and any other name
(for example `"../go.mod"`) gives `false`. This makes sure the server can only open our 3 banner files.

#### `type pageData struct { Text, Banner, Result, Error string }`
The data we give to the template. `{{.Text}}` in the HTML means "the `Text` field of this struct".
- `Text`: what the user typed, so the textarea keeps it after Generate.
- `Banner`: the chosen banner, so the right radio button stays checked.
- `Result`: the ASCII art. Empty means "no result to show".
- `Error`: a message for the user. Empty means "no error".

#### `func homeHandler(w, r)`
Answers `GET /`.
1. If the path is not exactly `/` (for example `/hello`), it answers **404**.
   This check is needed because the route `"/"` catches every unknown path.
2. If the method is not `GET`, it answers **405**.
3. Otherwise it sends the empty page with **200**, with `standard` chosen by default.

#### `func asciiArtHandler(w, r)`
Answers `POST /ascii-art`. This is the heart of the server.
1. If the method is not `POST`, it answers **405**.
2. `r.ParseForm()` reads the form data from the request body. If it is broken: **400**.
3. It reads the 2 form fields with `r.FormValue("text")` and `r.FormValue("banner")`.
   These names come from `name="text"` and `name="banner"` in `index.html`.
4. It calls `checkInput`. If there is a message, it shows the page again with the error and **400**.
5. It calls `ascii.Render` with the text and the path of the banner file, for example `banners/shadow.txt`.
6. It checks the error from `Render`:
   - `ErrInvalidChar` (like `é` or an emoji): the user's mistake, so **400** with a message.
   - `ErrBannerNotFound` (the file was deleted): **404**.
   - any other error: **500**, our fault.
7. If all went well, it puts the art in `data.Result` and sends the page with **200**.

#### `func checkInput(text, banner string) string`
Checks the form and returns a message for the user, or `""` if everything is fine:
empty text, text longer than 1000, or a banner that is not in the allow-list.
It is a separate function so the handler stays short and easy to read.

#### `func render(w, status, data)`
Fills the HTML template with `data` and sends it to the browser.
1. `template.ParseFiles` reads `templates/index.html`. We read it on **every request**, so if
   someone deletes the file, the server answers **404** instead of crashing.
   (Bonus: you can change the HTML and just refresh the browser, without restarting the server.)
2. `tmpl.Execute(&buf, data)` fills the template into a **buffer** (a box in memory), not straight
   into `w`. Why? Once we start writing the answer, the status code is sent and cannot be changed.
   If the template failed half way, the user would get half a page with code 200.
   With the buffer, we can still answer a clean **500**.
3. It sets the `Content-Type` header, writes the status code, then sends the buffer.

---

### `ascii/ascii.go`

#### Constants
- `charHeight = 8`: every character is 8 lines tall.
- `firstChar = ' '` (32) and `lastChar = '~'` (126): the range of characters we can draw.

#### Errors: `ErrBannerNotFound` and `ErrInvalidChar`
Named errors that the handler can recognize with `errors.Is`. This is how the `ascii`
package tells the web part *what kind* of problem happened, so the web part can choose the
right status code (400 or 404).

#### `func Render(text, bannerFile string) (string, error)`
The only exported function of the package (capital `R`). It turns text into art.
1. Replaces `\r\n` with `\n`. Browsers send a new line from a textarea as `\r\n` (2 characters).
   Without this, `\r` would be an invalid character.
2. Checks every character. If one is outside 32 to 126 (and is not a new line), it returns
   `ErrInvalidChar`. `%w` wraps the error, and `%q` adds the bad character to the message.
3. Calls `loadFont` to read the banner file.
4. Splits the text into lines with `strings.Split(text, "\n")`.
5. For each line: if it is empty, it writes one `"\n"`. Otherwise it writes the 8 rows
   (see section 4, "How a word is drawn").
6. Returns the finished string.

#### `func loadFont(bannerFile string) (map[rune][]string, error)`
Reads a banner file and returns a map: character -> its 8 lines.
1. `os.Open` opens the file. If it fails, it returns `ErrBannerNotFound`.
2. `defer file.Close()` closes the file when the function ends, whatever happens.
3. `bufio.Scanner` reads the file line by line into the slice `lines`.
   `scanner.Text()` gives each line without its `\n` (and without `\r` for Windows files).
4. It walks from `' '` to `'~'`: `font[char] = lines[start : start+8]`, then `start += 9`.
5. If the file is too short (a broken banner), it returns an error instead of crashing.

It starts with a small letter, so only code inside the `ascii` package can call it.

---

### `ascii/ascii_test.go`
Run all tests with:

```bash
go test ./...
```

| Test | What it checks |
|------|----------------|
| `TestRenderHi` | `"Hi"` with `standard` gives exactly the right art. |
| `TestRenderNewlines` | `\r\n` works and an empty line gives one empty row (8 + 1 + 8 = 17 rows). |
| `TestRenderInvalidChar` | `"héllo"` returns `ErrInvalidChar`. |
| `TestRenderMissingBanner` | a file that does not exist returns `ErrBannerNotFound`. |

A test is a function `TestXxx(t *testing.T)`. `t.Errorf` marks it as failed but keeps going;
`t.Fatalf` stops it right away.

---

### `templates/index.html`
Normal HTML, plus Go **template actions** between `{{ }}`:

| Action | Meaning |
|--------|---------|
| `{{.Text}}` | print the `Text` field of `pageData` |
| `{{if .Error}} ... {{end}}` | show this part only if `Error` is not empty |
| `{{if eq .Banner "shadow"}}checked{{end}}` | write `checked` only if the banner is `"shadow"` (`eq` means "equal") |

Important parts of the form:
- `method="POST" action="/ascii-art"`: where and how the form is sent.
- `name="text"` and `name="banner"`: the names the Go code reads with `r.FormValue`.
- `required` on the textarea: the browser refuses to send an empty text (the server checks it too,
  because someone could send a request without a browser).
- `<pre>` keeps all spaces and new lines, which ASCII art needs.
- `<script src="/static/script.js" defer>`: `defer` runs the script after the page is loaded,
  so the button already exists when the script looks for it.

---

### `static/script.js`
Makes the **Download .txt** button work. It runs in the browser, not on the server.
1. Finds the button (`id="download"`) and the art (`id="result"`).
   If there is no result yet, they do not exist, so the script does nothing.
2. On click:
   - `new Blob([text], { type: "text/plain" })` makes a file in memory with the art.
   - `URL.createObjectURL(file)` gives this file a temporary address.
   - It creates a link `<a>` with `download = "ascii-art.txt"` and clicks it: the browser saves the file.
   - `URL.revokeObjectURL(url)` frees the memory.

Why in JavaScript? The art is already on the page, so there is no need to ask the server again.
(Later, the Gritlab **ascii-art-web-export** project asks you to export from the **server**,
with HTTP headers like `Content-Disposition: attachment`. That would be a new Go handler.)

---

### `static/design.css`
The look of the page. Some ideas used:
- `display: flex` with `flex-direction: column` puts the form items under each other.
- `.output-header` uses flex with `justify-content: space-between` to put "Result" on the left
  and the button on the right.
- `.result` has `overflow-x: auto`: if the art is wider than the screen, you can scroll it
  instead of breaking the page. `font-family: monospace` makes every character the same width,
  which ASCII art needs.

---

## 6. Every package we use

The project uses **only the Go standard library**, as the subject asks. `go.mod` has no
`require` lines, which proves it.

### Go packages

| Package | Used in | What we use from it |
|---------|---------|---------------------|
| `net/http` | `main.go`, `handlers.go` | `HandleFunc` / `Handle` (connect a path to a handler), `ListenAndServe` (start the server), `FileServer` + `StripPrefix` + `Dir` (send static files), `ResponseWriter` (where we write the answer), `Request` (the incoming request: `Method`, `URL.Path`, `ParseForm`, `FormValue`), `Error` (send a short text error with a status code), `StatusOK` and the other status constants, `MethodGet` / `MethodPost`. |
| `html/template` | `handlers.go` | `ParseFiles` reads the HTML file, `Execute` fills it with data. It **escapes** HTML: if someone types `<script>`, the page shows the text `<script>` and does not run it. This protects against XSS attacks. That is why we use it and not `text/template`. |
| `bytes` | `handlers.go` | `bytes.Buffer`, a growing box of bytes in memory. We fill the template into it, then send it with `buf.WriteTo(w)`. |
| `errors` | `handlers.go`, `ascii.go`, test | `errors.New` creates a named error. `errors.Is(err, target)` checks if `err` is (or wraps) `target`. |
| `fmt` | `ascii.go` | `fmt.Errorf` builds an error message. `%w` wraps another error inside (so `errors.Is` still finds it), `%v` prints a value, `%q` prints a character in quotes, `%d` prints a number. |
| `log` | `main.go`, `handlers.go` | `log.Println` / `log.Printf` print a message with the time, in the terminal (for you, not for the user). `log.Fatal` prints and stops the program. |
| `os` | `ascii.go` | `os.Open` opens a file for reading. |
| `bufio` | `ascii.go` | `bufio.NewScanner` reads a file line by line: `Scan()` goes to the next line, `Text()` gives it, `Err()` tells if reading failed. |
| `strings` | `ascii.go`, test | `ReplaceAll` (change `\r\n` into `\n`), `Split` (cut the text into lines), `Builder` (build a long string fast with `WriteString`), `Join` and `Count` (in the test). |
| `path/filepath` | `handlers.go` | `filepath.Join("banners", "shadow.txt")` builds a file path with the right separator for your system (`/` on Linux, `\` on Windows). |
| `testing` | `ascii_test.go` | `testing.T`, `t.Errorf`, `t.Fatalf`. |
| `ascii-art-web/ascii` | `handlers.go` | **Our own** package: `ascii.Render`, `ascii.ErrInvalidChar`, `ascii.ErrBannerNotFound`. The name starts with the module name from `go.mod`. |

### Browser (JavaScript) tools used in `script.js`

| Tool | What it does |
|------|--------------|
| `document.getElementById(id)` | finds the HTML element with this `id`, or `null` |
| `element.addEventListener("click", fn)` | runs `fn` every time the element is clicked |
| `element.textContent` | the plain text inside an element (here: the art) |
| `new Blob([text], { type })` | a file made in memory |
| `URL.createObjectURL(blob)` | a temporary address for that file |
| `document.createElement("a")` | creates a new link element |
| `link.download = "name.txt"` | tells the browser to save the link target as a file with this name |
| `link.click()` | clicks the link from code |
| `URL.revokeObjectURL(url)` | frees the temporary address |

---

## 7. Status codes, and how to test them yourself

| Code | When |
|------|------|
| `200 OK` | the page or the result is shown |
| `400 Bad Request` | empty text, text too long, unknown banner, or a character that is not printable ASCII |
| `404 Not Found` | unknown path, or the template / banner file is missing |
| `405 Method Not Allowed` | wrong method, like `GET /ascii-art` or `POST /` |
| `500 Internal Server Error` | any other unexpected problem |

Start the server in one terminal:

```bash
go run .
```

Then try these in another terminal (Git Bash). `-w "%{http_code}"` prints the status code.

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/
```

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/nothing
```

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/ascii-art
```

```bash
curl -s -o /dev/null -w "%{http_code}\n" -d text=Hi -d banner=standard http://localhost:8080/ascii-art
```

```bash
curl -s -o /dev/null -w "%{http_code}\n" -d text=Hi -d banner=../go.mod http://localhost:8080/ascii-art
```

Expected: `200`, `404`, `405`, `200`, `400`.

To see a 404 for a missing template: rename `templates/index.html`, refresh the page, then rename it back.

---

## 8. Questions an auditor may ask you

**Why does `homeHandler` check `r.URL.Path != "/"`?**
The route `"/"` catches every path no other route matches. Without the check, `/anything` would show the home page instead of a 404.

**What happens with `banner=../../secret.txt`?**
It is not in the `banners` map, so `checkInput` answers 400. The server never opens the file.

**What happens if I type `<script>alert(1)</script>`?**
`<`, `>`, `(` and so on are printable ASCII, so they are drawn as art. In the textarea,
`html/template` escapes the text, so it is shown as text and never runs.

**Why do you replace `\r\n`?**
Browsers send new lines from a textarea as `\r\n`. `\r` is not printable, so without this, every text with a new line would be rejected.

**Why write the template into a buffer first?**
After the first byte is written to `w`, the status code is sent and cannot change. With the buffer, a template error can still become a clean 500.

**Why is the template read on every request?**
So that a missing template gives a 404 (as the subject asks) and not a crash at start-up.

**Is it only standard packages?**
Yes. `go.mod` has no `require` line, and every import is from the standard library or our own `ascii` package.

**Where is the formula?**
In `loadFont`: `start` begins at 1 and grows by 9 for each character, which is the same as `1 + (c - 32) * 9`.

---

## 9. Small exercises to test yourself

Try them one by one. If you can do them, you understand the project.

1. Change the port to `:3000`. Which file and which line?
2. Change the downloaded file name to include the banner, like `ascii-art-shadow.txt`.
   Hint: put the banner in the HTML with `data-banner="{{.Banner}}"` and read it in `script.js`.
3. Add a **Copy** button next to Download. Hint: `navigator.clipboard.writeText(text)`.
4. Make the max length 200 and check that a long text gives the error message.
5. Write a new test: `Render("\n", standard)` should return `"\n\n"`. Why two?
   (Hint: what does `strings.Split("\n", "\n")` return?)
6. Draw on paper how `"Hi"` becomes 8 rows, using `font['H'][row]` and `font['i'][row]`.
