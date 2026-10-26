// Package ascii turns normal text into big ASCII art letters.
package ascii

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	charHeight = 8   // every character is drawn with 8 lines
	firstChar  = ' ' // ASCII 32, the first character in a banner file
	lastChar   = '~' // ASCII 126, the last character in a banner file
)

var (
	// ErrBannerNotFound is returned when the banner file cannot be opened.
	ErrBannerNotFound = errors.New("banner not found")
	// ErrInvalidChar is returned when the text has a character the banners cannot draw.
	ErrInvalidChar = errors.New("invalid character")
)

// Render draws text as ASCII art, using the font in bannerFile.
func Render(text string, bannerFile string) (string, error) {
	// Browsers send textarea newlines as \r\n
	text = strings.ReplaceAll(text, "\r\n", "\n")

	for _, char := range text {
		if char != '\n' && (char < firstChar || char > lastChar) {
			return "", fmt.Errorf("%w: %q", ErrInvalidChar, char)
		}
	}

	font, err := loadFont(bannerFile)
	if err != nil {
		return "", err
	}

	var art strings.Builder
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			art.WriteString("\n")
			continue
		}

		// Build the 8 rows one by one: row 0 of every character, then row 1...
		for row := 0; row < charHeight; row++ {
			for _, char := range line {
				art.WriteString(font[char][row])
			}
			art.WriteString("\n")
		}
	}

	return art.String(), nil
}

// loadFont reads a banner file and returns the 8 lines that draw each character.
// The file starts with one empty line, then every character takes 9 lines:
// its 8 drawing lines followed by 1 empty separator line.
func loadFont(bannerFile string) (map[rune][]string, error) {
	file, err := os.Open(bannerFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBannerNotFound, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading banner file: %w", err)
	}

	font := make(map[rune][]string)
	start := 1 // skip the empty first line
	for char := firstChar; char <= lastChar; char++ {
		if start+charHeight > len(lines) {
			return nil, fmt.Errorf("banner file %s is incomplete: %d lines", bannerFile, len(lines))
		}
		font[char] = lines[start : start+charHeight]
		start += charHeight + 1 // jump over the 8 lines and the separator
	}

	return font, nil
}
