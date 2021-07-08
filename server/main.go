package main

import (
	"context"
	"fmt"
	"log"
	"net"

	tictac "github.com/p9kim/ticcytac/proto"
	"google.golang.org/grpc"

	"github.com/teris-io/shortid"
)

type Server struct{}

type board struct {
	spaces    [][]string
	boardSize int
	players   map[string]bool //whoever's turn it is will be set to true
	player1id string
	player2id string
	gameID    string
}

const maxSize = 3

const maxPlayers = 2

func newBoard() *board {
	b := board{}
	b.boardSize = maxSize
	b.spaces = make([][]string, maxSize)
	for i := 0; i < len(b.spaces); i++ {
		b.spaces[i] = make([]string, maxSize)
	}
	b.players = make(map[string]bool)
	return &b
}

//var gameBoard *board

var ongoingGames = make(map[string]*board)

func (s *Server) SayHello(ctx context.Context, req *tictac.RpcRequest) (*tictac.RpcResponse, error) {
	log.Printf("Message received from client: %s", req.HelloWorldRequest.GreetingMessage)
	res := tictac.RpcResponse{
		HellowWorldResponse: &tictac.HellowWorldResponse{
			ReturnedMessage: "Kenobi~!!! You are a BOLD ONE!",
		},
	}
	return &res, nil
}

func (s *Server) InitiateGame(ctx context.Context, req *tictac.RpcRequest) (*tictac.RpcResponse, error) {
	log.Printf("Create Game Request ID: %s", req.CreateGameRequest.UserId)
	gameBoard := newBoard()
	gameBoard.gameID, _ = shortid.Generate()
	gameBoard.player1id = req.CreateGameRequest.UserId
	gameBoard.players[gameBoard.player1id] = true
	ongoingGames[gameBoard.gameID] = gameBoard

	res := tictac.RpcResponse{
		CreateGameResponse: &tictac.CreateGameResponse{
			GameId: gameBoard.gameID,
		},
	}

	return &res, nil
}

func (s *Server) JoinGame(ctx context.Context, req *tictac.RpcRequest) (*tictac.RpcResponse, error) {
	log.Printf("Player %s joining game %s", req.JoinGameRequest.UserId, req.JoinGameRequest.GameId)

	userid := req.JoinGameRequest.UserId
	gameid := req.JoinGameRequest.GameId

	var res tictac.RpcResponse

	if gameBoard, ok := ongoingGames[gameid]; ok {
		if len(gameBoard.players) > maxPlayers {
			res = tictac.RpcResponse{
				JoinGameResponse: &tictac.JoinGameResponse{
					Result: tictac.JoinResult_GameIsFull,
				},
			}
		} else {
			gameBoard.player2id = userid
			gameBoard.players[gameBoard.player2id] = false
			res = tictac.RpcResponse{
				JoinGameResponse: &tictac.JoinGameResponse{
					Result: tictac.JoinResult_JoinSuccess,
				},
			}
		}
	} else {
		res = tictac.RpcResponse{
			JoinGameResponse: &tictac.JoinGameResponse{
				Result: tictac.JoinResult_NoGame,
			},
		}
	}

	return &res, nil
}

func (s *Server) MakeAMove(ctx context.Context, req *tictac.RpcRequest) (*tictac.RpcResponse, error) {
	log.Printf("Player %s occupies space [%d, %d]", req.OccupyPositionRequest.UserId,
		req.OccupyPositionRequest.X, req.OccupyPositionRequest.Y)

	x := req.OccupyPositionRequest.X
	y := req.OccupyPositionRequest.Y
	userid := req.OccupyPositionRequest.UserId
	gameid := req.OccupyPositionRequest.GameId

	var res tictac.RpcResponse

	// Ugly
	if gameBoard, ok := ongoingGames[gameid]; ok {
		if player, ok := gameBoard.players[userid]; ok {
			if !player {
				res = tictac.RpcResponse{
					OccupyPositionResponse: &tictac.OccupyPositionResponse{
						OccupyResult: tictac.OccupyResult_NotYourTurn,
					},
				}
			} else if x > 2 || y > 2 {
				res = tictac.RpcResponse{
					OccupyPositionResponse: &tictac.OccupyPositionResponse{
						OccupyResult: tictac.OccupyResult_InvalidPosition,
					},
				}
			} else if gameBoard.spaces[x][y] != "" {
				res = tictac.RpcResponse{
					OccupyPositionResponse: &tictac.OccupyPositionResponse{
						OccupyResult: tictac.OccupyResult_HasBeenTaken,
					},
				}
			} else {
				gameBoard.spaces[x][y] = userid
				if gameBoard.players[gameBoard.player1id] {
					gameBoard.players[gameBoard.player1id] = false
					gameBoard.players[gameBoard.player2id] = true
				} else {
					gameBoard.players[gameBoard.player1id] = true
					gameBoard.players[gameBoard.player2id] = false
				}
				res = tictac.RpcResponse{
					OccupyPositionResponse: &tictac.OccupyPositionResponse{
						OccupyResult: tictac.OccupyResult_OccupySuccess,
					},
				}
			}
		} else {
			res = tictac.RpcResponse{
				OccupyPositionResponse: &tictac.OccupyPositionResponse{
					OccupyResult: tictac.OccupyResult_NotAPlayer,
				},
			}
		}
	} else {
		res = tictac.RpcResponse{
			OccupyPositionResponse: &tictac.OccupyPositionResponse{
				OccupyResult: tictac.OccupyResult_InvalidGame,
			},
		}
	}

	return &res, nil
}

func (s *Server) CheckGameStatus(ctx context.Context, req *tictac.RpcRequest) (*tictac.RpcResponse, error) {
	log.Printf("Checking Game Result: %s", req.CheckGameResultRequest.GameId)
	gameid := req.CheckGameResultRequest.GameId
	gameBoard := ongoingGames[gameid]

	var res tictac.RpcResponse

	if len(gameBoard.players) < 2 {
		res = tictac.RpcResponse{
			CheckGameResultResponse: &tictac.CheckGameResultResponse{
				GameResult: tictac.GameResult_WaitMoreJoin,
			},
		}
	} else {
		if isFull(gameBoard.spaces) {
			res = tictac.RpcResponse{
				CheckGameResultResponse: &tictac.CheckGameResultResponse{
					GameResult: tictac.GameResult_Draw,
				},
			}
			delete(ongoingGames, gameid)
		} else {
			res = tictac.RpcResponse{
				CheckGameResultResponse: &tictac.CheckGameResultResponse{
					GameResult: tictac.GameResult_Ongoing,
				},
			}
		}
	}

	return &res, nil
}

func isFull(spaces [][]string) bool {
	for i := 0; i < len(spaces); i++ {
		for j := 0; j < len(spaces[i]); j++ {
			if spaces[i][j] == "" {
				return false
			}
		}
	}

	return true
}

func main() {
	fmt.Println("Starting gRPC server!!")

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}

	//s := proto.Server{}

	grpcServer := grpc.NewServer()

	tictac.RegisterGameServerServer(grpcServer, &Server{})

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
