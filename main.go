package main

import (
	"flag"
	"go-grpc-client/domain"
	"go-grpc-client/shared/proto"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var GRPCClient proto.GreetingServiceClient

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

func main() {

	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		// TODO TLS example
		// opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Create connection to Go GRPC (GRPC Server)
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("Unable to connect to GRPC Server : %s ", err.Error())
	}

	// Create Service Client
	GRPCClient = proto.NewGreetingServiceClient(conn)

	e := echo.New()
	// Greeting
	e.GET("/greeting/generic", GetGenericGreeting)
	e.GET("/greeting/named", GetNamedGreeting)
	e.POST("/greeting/verbose", GetVerboseGreeting)

	e.Logger.Fatal(e.Start(":1323"))
}

func GetGenericGreeting(c echo.Context) error {

	resp, err := GRPCClient.GetGenericGreeting(c.Request().Context(), &emptypb.Empty{})

	if err != nil {
		st, _ := status.FromError(err)

		return c.JSON(http.StatusBadRequest, domain.Response{
			Code:     http.StatusBadRequest,
			GRPCCode: int(st.Code()),
			Message:  st.Message(),
		})
	}

	return c.JSON(http.StatusOK, domain.Response{
		Code:     http.StatusOK,
		GRPCCode: 0,
		Message:  resp.Message,
	})
}

func GetNamedGreeting(c echo.Context) error {

	name := c.QueryParam("name")
	resp, err := GRPCClient.GetNamedGreeting(c.Request().Context(), &proto.GetNamedGreetingRequest{Name: name})

	if err != nil {
		st, _ := status.FromError(err)

		return c.JSON(http.StatusBadRequest, domain.Response{
			Code:     http.StatusBadRequest,
			GRPCCode: int(st.Code()),
			Message:  st.Message(),
		})
	}

	return c.JSON(http.StatusOK, domain.Response{
		Code:     http.StatusOK,
		GRPCCode: 0,
		Message:  resp.Message,
	})
}

func GetVerboseGreeting(c echo.Context) error {

	req := new(domain.VerboseGreetingRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	reqProto := &proto.GetVerboseGreetingRequest{Name: req.Name,
		Age: int64(req.Age),
		FavoriteGame: &proto.Game{
			Name:    req.FavoriteGame.Name,
			Console: req.FavoriteGame.Console,
		}}

	resp, err := GRPCClient.GetVerboseGreeting(c.Request().Context(), reqProto)

	if err != nil {
		st, _ := status.FromError(err)

		return c.JSON(http.StatusBadRequest, domain.Response{
			Code:     http.StatusBadRequest,
			GRPCCode: int(st.Code()),
			Message:  st.Message(),
		})
	}

	return c.JSON(http.StatusOK, domain.Response{
		Code:     http.StatusOK,
		GRPCCode: 0,
		Message:  resp.Message,
	})
}
