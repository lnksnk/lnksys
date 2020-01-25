package typescript

import (
	"io"
	"strings"
)


func TypescriptJS() io.Reader {
	return strings.NewReader(strings.Replace(typescriptjs, "|'|", "`", -1))
}

func TypescriptFindJS(typescriptfindjs string) io.Reader {
	if strings.LastIndex(typescriptfindjs, "/") >= 0 {
		typescriptfindjs = typescriptfindjs[strings.LastIndex(typescriptfindjs, "/")+1:]
	}
	if typescriptfindjs == "typescript.js" {
		return TypescriptJS()
	}
	return nil
}