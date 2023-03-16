package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"ent-grpc-example/ent/proto/entpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	// サーバーとのコネクションを開く
	conn, err := grpc.Dial(":5000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed connecting to server: %s", err)
	}
	defer conn.Close()

	// Create a User service Client on the connection.
	client := entpb.NewUserServiceClient(conn)

	// サーバーにランダムなUserの作成を依頼する
	ctx := context.Background()
	user := randomUser()
	created, err := client.Create(ctx, &entpb.CreateUserRequest{
		User: user,
	})
	if err != nil {
		se, _ := status.FromError(err)
		log.Fatalf("failed creating user: status=%s message=%s", se.Code(), se.Message())
	}
	log.Printf("user created with id: %d", created.Id)

	// 別のRPC呼び出しで、先に作成したUserを取得
	get, err := client.Get(ctx, &entpb.GetUserRequest{
		Id: created.Id,
	})
	if err != nil {
		se, _ := status.FromError(err)
		log.Fatalf("failed retrieving user: status=%s message=%s", se.Code(), se.Message())
	}
	log.Printf("retrieved user with id=%d: %v", get.Id, get)

}

func randomUser() *entpb.User {
	return &entpb.User{
		Name:         fmt.Sprintf("user_%d", rand.Int()),
		EmailAddress: fmt.Sprintf("user_%d@example.com", rand.Int()),
	}
}
