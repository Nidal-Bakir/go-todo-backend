package resutils

import (
	"context"
	"net/http"
	"strings"
)

func webWriteError(ctx context.Context, w http.ResponseWriter, code int, errs ...error) {
	sb := strings.Builder{}
	for _, v := range errs {
		sb.WriteString("<h1>")
		sb.WriteString(v.Error())
		sb.WriteString("</h1>")
		sb.WriteString("\n")
	}

	w.WriteHeader(code)
	w.Write([]byte(sb.String()))
}
