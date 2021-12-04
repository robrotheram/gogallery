package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/robrotheram/gogallery/config"
)

type Prediction struct {
	Status      string `json:"status"`
	Predictions []struct {
		Index       string  `json:"index"`
		Caption     string  `json:"caption"`
		Probability float64 `json:"probability"`
	} `json:"predictions"`
}

func GetCaptions(picture *Picture) (prediction Prediction, err error) {
	// Prepare a form that you will submit to that URL.
	url := config.Config.Server.CaptionURL
	if url == "" {
		return prediction, fmt.Errorf("caption server is disabled unable to generate captions")
	}

	var b bytes.Buffer
	var fw io.Writer

	file, err := os.Open(picture.Path)
	if err != nil {
		return prediction, fmt.Errorf("unable to open image path")
	}

	w := multipart.NewWriter(&b)
	client := &http.Client{
		Timeout: time.Second * 60,
	}

	if fw, err = CreateFormImageFile(w, "image", file.Name()); err != nil {
		return
	}
	if _, err = io.Copy(fw, file); err != nil {
		return
	}
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}
	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	bodyBytes, err := io.ReadAll(res.Body)
	json.Unmarshal(bodyBytes, &prediction)

	// if len(prediction.Predictions) > 0 {
	// 	fmt.Println("Caption: " + prediction.Predictions[0].Caption)
	// 	picture.Caption = prediction.Predictions[0].Caption
	// 	picture.Save()
	// }
	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// CreateFormFile is a convenience wrapper around CreatePart. It creates
// a new form-data header with the provided field name and file name.
func CreateFormImageFile(w *multipart.Writer, fieldname, filename string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", "image/jpeg")
	return w.CreatePart(h)
}

const remoteURL = "http://172.17.0.2:5000/model/predict"
