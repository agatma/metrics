package rest_test

import (
	"github.com/go-resty/resty/v2"
)

func ExampleHandler_SetMetricValue() {
	client := resty.New()
	client.R().
		SetHeader("Content-Type", "text/plain").
		Post("http:/localhost:8080/update/counter/someMetric/999")
	/*
		HTTP Request example:
		POST /update/counter/someMetric/527 HTTP/1.1
		Host: localhost:8080
		Content-Length: 0
		Content-Type: text/plain


		HTTP Response example:
		HTTP/1.1 200 OK
		Date: Tue, 29 Oct 2024 09:11:35 GMT
		Content-Length: 11
		Content-Type: text/plain; charset=utf-8
	*/
}
