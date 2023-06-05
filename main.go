package main

import (
	"context"
	"fmt"
	"go-grpc-client/shared/proto"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var GRPCClient proto.GreetingServiceClient

type Response struct {
	Code     int
	GRPCCode int
	Message  string
	Data     interface{}
}

func OpenServiceClient() (*grpc.ClientConn, error) {
	host := "localhost"
	port := 50052
	timeout := 30000
	ctx, _ := context.WithTimeout(
		context.Background(),
		time.Duration(timeout)*time.Millisecond,
	)
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithInsecure())

	// opts = append(opts, grpc.WithUnaryInterceptor(clientInterceptor))

	opts = append(opts, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:%d", host, port),
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func main() {

	grpcClient, err := OpenServiceClient()
	if err != nil {
		log.Fatalf("Unable to connect to GRPC Server : %s ", err.Error())
	}
	GRPCClient = proto.NewGreetingServiceClient(grpcClient)

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/generic", GetGenericGreeting)

	e.Logger.Fatal(e.Start(":1323"))
}

func GetGenericGreeting(c echo.Context) error {

	resp, err := GRPCClient.GetGenericGreeting(c.Request().Context(), &emptypb.Empty{})

	if err != nil {
		st, _ := status.FromError(err)

		return c.JSON(http.StatusBadRequest, Response{
			Code:     http.StatusBadRequest,
			GRPCCode: int(st.Code()),
			Message:  st.Message(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Code:     http.StatusOK,
		GRPCCode: 0,
		Message:  resp.Message,
	})
}
