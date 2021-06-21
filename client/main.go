package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	gravi "github.com/p9kim/ticcytac/proto"

	"google.golang.org/grpc"
)

func main() {
	done := make(chan int)
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	client := gravi.NewGameServerClient(conn)

	// Create Game
	userid := ""

	gameid := ""
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Are you joining a game or creating a new one? gameid = join, C = create")
		for scanner.Scan() {
			option := scanner.Text()
			if option == "C" {
				userid = "X"
				createGameReq := gravi.RpcRequest{
					CreateGameRequest: &gravi.CreateGameRequest{
						UserId: userid,
					},
				}
				game, err := client.InitiateGame(context.Background(), &createGameReq)
				if err != nil {
					log.Fatal(err)
				}
				gameid = game.CreateGameResponse.GameId
				fmt.Printf("Here's the game ID to share: %s\n", gameid)
				break
			} else {
				userid = "O"
				gameid = option
				joinGameReq := gravi.RpcRequest{
					JoinGameRequest: &gravi.JoinGameRequest{
						UserId: userid,
						GameId: gameid,
					},
				}
				joinResult, err := client.JoinGame(context.Background(), &joinGameReq)
				if err != nil {
					log.Fatal(err)
				}
				if joinResult.JoinGameResponse.Result != gravi.JoinResult_JoinSuccess {
					fmt.Println("No game by this id")
					break
				} else {
					fmt.Println("Join game success, let's win!")
					break
				}
			}
		}
		fmt.Println("Let's play!")
		for scanner.Scan() {
			checkGameReq := gravi.RpcRequest{
				CheckGameResultRequest: &gravi.CheckGameResultRequest{
					GameId: gameid,
				},
			}
			gameResult, err := client.CheckGameStatus(context.Background(), &checkGameReq)
			if err != nil {
				log.Fatal(err)
			}
			if gameResult.CheckGameResultResponse.GameResult == gravi.GameResult_WaitMoreJoin {
				fmt.Println("Wait for more 1 more player")
				continue
			} else if gameResult.CheckGameResultResponse.GameResult != gravi.GameResult_Ongoing {
				fmt.Println("Game is over")
			}

			position := strings.Split(scanner.Text(), ",")
			x, _ := strconv.Atoi(position[0])
			y, _ := strconv.Atoi(position[1])
			move := gravi.RpcRequest{
				OccupyPositionRequest: &gravi.OccupyPositionRequest{
					GameId: gameid,
					UserId: userid,
					X:      int32(x),
					Y:      int32(y),
				},
			}

			res, err := client.MakeAMove(context.Background(), &move)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Your move status: %s\n", res.OccupyPositionResponse.OccupyResult)
		}
	}()

	go func() {
		wg.Wait()
		close(done)
	}()

	<-done
}
