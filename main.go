package main

import (
	"context"
	"flag"
	"fmt"
	"go-grpc-client/domain"
	"go-grpc-client/shared/proto"
	"io"
	"log"
	"net/http"
	"time"

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
	e.GET("/villagers/serverstream", FindAllStreamServerSide)
	e.POST("/villagers/clientstream", FindStreamClientSide)
	e.POST("/villagers/bidirectionalstream", FindStreamBidirectional)

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

func FindAllStreamServerSide(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	stream, err := VillagersGRPCClient.FindAllStreamServerSide(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("client.FindAllStreamServerSide failed: %v", err)
	}

	villagerList := make([]domain.Villager, 0)
	for {
		villager, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("client.FindAllStreamServerSide failed: %v", err)
		}
		villagerList = append(villagerList, domain.Villager{
			Name:        villager.Name,
			Personality: villager.Personality,
		})
	}

	return c.JSON(http.StatusOK, domain.Response{
		Code:     http.StatusOK,
		GRPCCode: 0,
		Data:     villagerList,
	})

}

func FindStreamClientSide(c echo.Context) error {

	req := new(domain.FindStreamClientSideRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	stream, err := VillagersGRPCClient.FindStreamClientSide(ctx)
	if err != nil {
		log.Fatalf("client.FindStreamClientSide failed: %v", err)
	}

	for _, name := range req.Name {
		reqProto := proto.FindStreamClientSideRequest{
			Name: name,
		}
		if err := stream.Send(&reqProto); err != nil {
			log.Fatalf("client.FindStreamClientSide: stream.Send(%v) failed: %v", &reqProto, err)
		}

	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("client.FindStreamClientSide failed: %v", err)
	}

	return c.JSON(http.StatusOK, domain.Response{
		Code:     http.StatusOK,
		GRPCCode: 0,
		Data:     reply,
	})

}

func FindStreamBidirectional(c echo.Context) error {

	req := new(domain.FindStreamClientSideRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	stream, err := VillagersGRPCClient.FindStreamBidirecitonal(ctx)
	if err != nil {
		log.Fatalf("client.FindStreamBidirecitonal failed: %v", err)
	}

	villagerList := make([]domain.Villager, 0)

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("client.FindStreamBidirecitonal failed: %v", err)
			}
			newVillager := domain.Villager{
				Name:        in.Name,
				Personality: in.Personality,
			}
			fmt.Println("Received : ")
			fmt.Println(newVillager)
			villagerList = append(villagerList, newVillager)
		}
	}()

	for _, name := range req.Name {
		reqProto := proto.FindStreamClientSideRequest{
			Name: name,
		}
		if err := stream.Send(&reqProto); err != nil {
			log.Fatalf("client.FindStreamBidirecitonal: stream.Send(%v) failed: %v", &reqProto, err)
		}

	}
	stream.CloseSend()
	<-waitc
	return c.JSON(http.StatusOK, domain.Response{
		Code:     http.StatusOK,
		GRPCCode: 0,
		Data:     villagerList,
	})

}
