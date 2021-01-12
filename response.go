package main

import (
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"



	"google.golang.org/genproto/googleapis/rpc/code"
)

var ErrorKey = label.Key("error")

var microCodeToStatusCode = map[int32]code.Code{
	400: code.Code_INVALID_ARGUMENT,
	401: code.Code_UNAUTHENTICATED,
	403: code.Code_PERMISSION_DENIED,
	404: code.Code_NOT_FOUND,
	409: code.Code_ABORTED,
	500: code.Code_INTERNAL,
}

//func getResponseStatus(err error) trace.Status {
//	if err != nil {
//		microErr, ok := err.(*microerr.Error)
//		if ok {
//			statusCode := microErr.Code
//			code, ok := microCodeToStatusCode[microErr.Code]
//			if ok {
//				statusCode = int32(code)
//			}
//
//			return trace.Status{
//				Code:    statusCode,
//				Message: fmt.Sprintf("%s: %s", microErr.Id, microErr.Detail),
//			}
//		}
//
//		return trace.Status{
//			Code:    int32(code.Code_UNKNOWN),
//			Message: err.Error(),
//		}
//	}
//
//	return trace.Status{}
//}

func setResponseStatus(span trace.Span, err error, msg string) {
	if err == nil {
		span.SetStatus(codes.Ok, msg)
		return
	}

	span.SetStatus(codes.Error, msg)
	span.SetAttributes(ErrorKey.String(err.Error()))
}
