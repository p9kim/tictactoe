package main

import (
	"context"
	"log"

	gravi "github.com/p9kim/ticcytac/proto"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	client := gravi.NewGameServerClient(conn)

	// Ping
	message := gravi.RpcRequest{
		HelloWorldRequest: &gravi.HelloWorldRequest{
			GreetingMessage: "Hello There~!",
		},
		RpcType: gravi.RpcType_HellowWorld,
	}

	res, err := client.SayHello(context.Background(), &message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from server: %s\n", res.HellowWorldResponse.ReturnedMessage)

	// Create Game
	createGameReq := gravi.RpcRequest{
		CreateGameRequest: &gravi.CreateGameRequest{
			UserId: "Paul",
		},
	}

	res2, err := client.InitiateGame(context.Background(), &createGameReq)
	if err != nil {
		log.Fatal(err)
	}
	gameid := res2.CreateGameResponse.GameId
	log.Printf("Game ID: %s", gameid)

	// Make a move
	occupyReq := gravi.RpcRequest{
		OccupyPositionRequest: &gravi.OccupyPositionRequest{
			UserId: "Paul",
			GameId: gameid,
			X:      1,
			Y:      1,
		},
	}

	occupyReq2 := gravi.RpcRequest{
		OccupyPositionRequest: &gravi.OccupyPositionRequest{
			UserId: "Paul",
			GameId: gameid,
			X:      1,
			Y:      1,
		},
	}

	occupyReq3 := gravi.RpcRequest{
		OccupyPositionRequest: &gravi.OccupyPositionRequest{
			UserId: "Paul",
			GameId: gameid,
			X:      4,
			Y:      4,
		},
	}

	res3, err := client.MakeAMove(context.Background(), &occupyReq)
	if err != nil {
		log.Fatal(err)
	}
	res4, err := client.MakeAMove(context.Background(), &occupyReq2)
	if err != nil {
		log.Fatal(err)
	}
	res5, err := client.MakeAMove(context.Background(), &occupyReq3)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res3.OccupyPositionResponse.OccupyResult)
	log.Println(res4.OccupyPositionResponse.OccupyResult)
	log.Println(res5.OccupyPositionResponse.OccupyResult)

}
