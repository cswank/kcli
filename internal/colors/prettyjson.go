package colors

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

//much of this code is from https://github.com/hokaccha/go-prettyjson so I could make changes.
var (
	colors = map[string]color.Attribute{
		"black":   color.FgBlack,
		"blue":    color.FgBlue,
		"cyan":    color.FgCyan,
		"green":   color.FgGreen,
		"magenta": color.FgMagenta,
		"red":     color.FgRed,
		"white":   color.FgWhite,
		"yellow":  color.FgYellow,
	}
)

//Formatter is a struct to format JSON data. `color` is github.com/fatih/color: https://github.com/fatih/color
type Formatter struct {
	// JSON key color. Default is `color.New(color.FgBlue, color.Bold)`.
	KeyColor *color.Color

	// JSON string value color. Default is `color.New(color.FgGreen, color.Bold)`.
	StringColor *color.Color

	// JSON boolean value color. Default is `color.New(color.FgYellow, color.Bold)`.
	BoolColor *color.Color

	// JSON number value color. Default is `color.New(color.FgCyan, color.Bold)`.
	NumberColor *color.Color

	// JSON null value color. Default is `color.New(color.FgBlack, color.Bold)`.
	NullColor *color.Color

	// Max length of JSON string value. When the value is 1 and over, string is truncated to length of the value. Default is 0 (not truncated).
	StringMaxLength int

	// Boolean to disable color. Default is false.
	DisabledColor bool

	// Indent space number. Default is 2.
	Indent int
}

// NewFormatter returns a new formatter with following default values.
func NewFormatter() *Formatter {
	return &Formatter{
		KeyColor:        getColor("KEY", "blue"),
		StringColor:     getColor("STRING", "green"),
		BoolColor:       getColor("BOOL", "yellow"),
		NumberColor:     getColor("NUMBER", "cyan"),
		NullColor:       getColor("NULL", "black"),
		StringMaxLength: 0,
		DisabledColor:   false,
		Indent:          2,
	}
}

func getColor(key string, defaultColor string) *color.Color {
	c := os.Getenv(fmt.Sprintf("KCLI_%s_COLOR", key))
	if c == "" {
		c = defaultColor
	}
	return color.New(colors[c], color.Bold)
}

// Marshal marshals and formats JSON data.
func (f *Formatter) Marshal(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)

	if err != nil {
		return nil, err
	}

	return f.Format(data)
}

//Format creates a colorful JSON string.
func (f *Formatter) Format(data []byte) ([]byte, error) {
	var v interface{}
	err := json.Unmarshal(data, &v)

	if err != nil {
		return nil, err
	}

	s := f.pretty(v, 1)

	return []byte(s), nil
}

func (f *Formatter) sprintfColor(c *color.Color, format string, args ...interface{}) string {
	if f.DisabledColor || c == nil {
		return fmt.Sprintf(format, args...)
	}
	return c.SprintfFunc()(format, args...)
}

func (f *Formatter) sprintColor(c *color.Color, s string) string {
	if f.DisabledColor || c == nil {
		return fmt.Sprint(s)
	}
	return c.SprintFunc()(s)
}

func (f *Formatter) pretty(v interface{}, depth int) string {
	switch val := v.(type) {
	case string:
		return f.processString(val)
	case float64:
		return f.sprintColor(f.NumberColor, strconv.FormatFloat(val, 'f', -1, 64))
	case bool:
		return f.sprintColor(f.BoolColor, strconv.FormatBool(val))
	case nil:
		return f.sprintColor(f.NullColor, "null")
	case map[string]interface{}:
		return f.processMap(val, depth)
	case []interface{}:
		return f.processArray(val, depth)
	}

	return ""
}

func (f *Formatter) processString(s string) string {
	r := []rune(s)

	if f.StringMaxLength != 0 && len(r) >= f.StringMaxLength {
		s = string(r[0:f.StringMaxLength]) + "..."
	}

	b, _ := json.Marshal(s)

	return f.sprintColor(f.StringColor, string(b))
}

func (f *Formatter) processMap(m map[string]interface{}, depth int) string {
	currentIndent := f.generateIndent(depth - 1)
	nextIndent := f.generateIndent(depth)
	rows := []string{}
	keys := []string{}

	if len(m) == 0 {
		return "{}"
	}

	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		val := m[key]
		k := f.sprintfColor(f.KeyColor, `"%s"`, key)
		v := f.pretty(val, depth+1)
		row := fmt.Sprintf("%s%s: %s", nextIndent, k, v)
		rows = append(rows, row)
	}

	return fmt.Sprintf("{\n%s\n%s}", strings.Join(rows, ",\n"), currentIndent)
}

func (f *Formatter) processArray(a []interface{}, depth int) string {
	currentIndent := f.generateIndent(depth - 1)
	nextIndent := f.generateIndent(depth)
	rows := []string{}

	if len(a) == 0 {
		return "[]"
	}

	for _, val := range a {
		c := f.pretty(val, depth+1)
		row := nextIndent + c
		rows = append(rows, row)
	}

	return fmt.Sprintf("[\n%s\n%s]", strings.Join(rows, ",\n"), currentIndent)
}

func (f *Formatter) generateIndent(depth int) string {
	return strings.Join(make([]string, f.Indent*depth+1), " ")
}

// Marshal JSON data with default options.
func Marshal(v interface{}) ([]byte, error) {
	return NewFormatter().Marshal(v)
}

//Format JSON string with default options.
func Format(data []byte) ([]byte, error) {
	return NewFormatter().Format(data)
}
