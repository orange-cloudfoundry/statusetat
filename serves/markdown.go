package serves

import (
	"io/ioutil"
	"net/http"

	"github.com/ArthurHlt/statusetat/markdown"
	"github.com/ArthurHlt/statusetat/models"
)

func (a Serve) convertMessageToHtml(messages []models.Message) []models.Message {
	for i, msg := range messages {
		content := markdown.Convert([]byte(msg.Content))
		msg.Content = string(content)
		messages[i] = msg
	}
	return messages
}

func (a Serve) preview(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		JSONError(w, err, http.StatusPreconditionRequired)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(markdown.Convert(b))
}
