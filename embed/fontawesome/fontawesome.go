package fontawesome

import (
	"io"
	"strings"
)

const fontawesomejs string = `/*!
* Font Awesome Free 5.11.2 by @fontawesome - https://fontawesome.com
* License - https://fontawesome.com/license/free (Icons: CC BY 4.0, Fonts: SIL OFL 1.1, Code: MIT License)
*/

const fontawesomecss string = `/*!
* Font Awesome Free 5.11.2 by @fontawesome - https://fontawesome.com
* License - https://fontawesome.com/license/free (Icons: CC BY 4.0, Fonts: SIL OFL 1.1, Code: MIT License)
*/

func FontawesomeCSS() io.Reader {
	return strings.NewReader(fontawesomecss)
}

func FontawesomeJS() io.Reader {
	return strings.NewReader(fontawesomejs)
}

func FontawesomeFindJSCSS(fontawesomecssjs string) io.Reader {
	if strings.LastIndex(fontawesomecssjs, "/") >= 0 {
		fontawesomecssjs = fontawesomecssjs[strings.LastIndex(fontawesomecssjs, "/")+1:]
	}
	if fontawesomecssjs == "fontawesome.css" {
		return FontawesomeCSS()
	} else if fontawesomecssjs == "fontawesome.js" {
		return FontawesomeJS()
	}
	return nil
}