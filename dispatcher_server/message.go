package dispatcher

import "github.com/gofrs/uuid"

type ResponseMessage struct {
	RequestId uuid.UUID         `bson:"request_id"`
	Token     string            `bson:"token"`
	Body      []byte            `bson:"body"`
	Status    int               `bson:"status"`
	Header    map[string]string `bson:"header"`
}

type RequestMessage struct {
	ID           uuid.UUID            `bson:"ID"`
	Method       string               `bson:"method"`
	URL          string               `bson:"url"`
	Body         []byte               `bson:"body"`
	Header       map[string]string    `bson:"header"`
	ResponseChan chan ResponseMessage `bson:"-"`
}
