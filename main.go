package main

import (
	"flag"
	"go-grpc-client/domain"
	"go-grpc-client/shared/proto"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var GreetingGRPCClient proto.GreetingServiceClient
var VillagersGRPCClient proto.VillagersServiceClient

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

func main() {

	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		// If using TLS
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
	GreetingGRPCClient = proto.NewGreetingServiceClient(conn)
	VillagersGRPCClient = proto.NewVillagersServiceClient(conn)

	e := echo.New()
	// Greeting
	e.GET("/greeting/generic", GetGenericGreeting)
	e.GET("/greeting/named", GetNamedGreeting)
	e.POST("/greeting/verbose", GetVerboseGreeting)

	e.GET("/villagers/:Name", GetVillagerByName)

	e.Logger.Fatal(e.Start(":1323"))
}

func GetGenericGreeting(c echo.Context) error {

	resp, err := GreetingGRPCClient.GetGenericGreeting(c.Request().Context(), &emptypb.Empty{})

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
	resp, err := GreetingGRPCClient.GetNamedGreeting(c.Request().Context(), &proto.GetNamedGreetingRequest{Name: name})

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

	resp, err := GreetingGRPCClient.GetVerboseGreeting(c.Request().Context(), reqProto)

	if err != nil {
		st, _ := status.FromError(err)

		return c.JSON(http.StatusInternalServerError, domain.Response{
			Code:     http.StatusInternalServerError,
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

func GetVillagerByName(c echo.Context) error {
	villagerName := c.Param("Name")

	reqProto := &proto.FindByNameRequest{
		Name: villagerName,
	}

	resp, err := VillagersGRPCClient.FindByName(c.Request().Context(), reqProto)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {

			return c.JSON(http.StatusNotFound, domain.Response{
				Code:     http.StatusNotFound,
				GRPCCode: int(st.Code()),
				Message:  st.Message(),
			})

		}
		return c.JSON(http.StatusInternalServerError, domain.Response{
			Code:     http.StatusInternalServerError,
			GRPCCode: int(st.Code()),
			Message:  st.Message(),
		})
	}

	return c.JSON(http.StatusOK, domain.Response{
		Code:     http.StatusOK,
		GRPCCode: 0,
		Data:     resp,
	})
}
