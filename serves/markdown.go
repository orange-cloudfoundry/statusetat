package serves

import (
	"io"
	"net/http"

	"github.com/orange-cloudfoundry/statusetat/v2/markdown"
	"github.com/orange-cloudfoundry/statusetat/v2/models"
)

func (a *Serve) convertMessageToHtml(messages []models.Message) []models.Message {
	for i, msg := range messages {
		content := markdown.Convert([]byte(msg.Content))
		msg.Content = string(content)
		messages[i] = msg
	}
	return messages
}

func (a *Serve) preview(w http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck
	w.Write(markdown.Convert(b))
}
