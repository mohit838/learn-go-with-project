package response

import (
	"net/http"

	starterresponse "github.com/mohit838/go-project-stater/response"
)

type Body = starterresponse.Body

func JSON(w http.ResponseWriter, statusCode int, body Body) {
	starterresponse.JSON(w, statusCode, body)
}

func Success(w http.ResponseWriter, statusCode int, message string, data any) {
	starterresponse.Success(w, statusCode, message, data)
}

func Error(w http.ResponseWriter, statusCode int, message string, err any) {
	starterresponse.Error(w, statusCode, message, err)
}
