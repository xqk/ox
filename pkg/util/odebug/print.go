package odebug

import (
	"fmt"

	"github.com/tidwall/pretty"
	"github.com/xqk/ox/pkg/util/ocolor"
	"github.com/xqk/ox/pkg/util/ostring"
)

// DebugObject ...
func PrintObject(message string, obj interface{}) {
	if !IsDevelopmentMode() {
		return
	}
	fmt.Printf("%s => %s\n",
		ocolor.Red(message),
		pretty.Color(
			pretty.Pretty([]byte(ostring.PrettyJson(obj))),
			pretty.TerminalStyle,
		),
	)
}

// DebugBytes ...
func DebugBytes(obj interface{}) string {
	return string(pretty.Color(pretty.Pretty([]byte(ostring.Json(obj))), pretty.TerminalStyle))
}

// PrintKV ...
func PrintKV(key string, val string) {
	if !IsDevelopmentMode() {
		return
	}
	fmt.Printf("%-50s => %s\n", ocolor.Red(key), ocolor.Green(val))
}

// PrettyKVWithPrefix ...
func PrintKVWithPrefix(prefix string, key string, val string) {
	if !IsDevelopmentMode() {
		return
	}
	fmt.Printf("%-8s]> %-30s => %s\n", prefix, ocolor.Red(key), ocolor.Blue(val))
}

// PrintMap ...
func PrintMap(data map[string]interface{}) {
	if !IsDevelopmentMode() {
		return
	}
	for key, val := range data {
		fmt.Printf("%-20s : %s\n", ocolor.Red(key), fmt.Sprintf("%+v", val))
	}
}
