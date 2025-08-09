package user_handler

import (
	"bufio"
	"errors"
	"os"

	pb "github.com/anarakinson/go_stonks/stonks_pb/gen/order/v1"
)

var (
	ErrFinish = errors.New("program finished")
	ErrStdin  = errors.New("stdin error")
	ErrNoData = errors.New("no data")
)

type UserHandler struct {
	reader *bufio.Reader
	client pb.OrderServiceClient
}

func NewUserHandler(client pb.OrderServiceClient) *UserHandler {
	return &UserHandler{
		reader: bufio.NewReader(os.Stdin),
		client: client,
	}
}
