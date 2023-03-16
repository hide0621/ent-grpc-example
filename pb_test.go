package main

import (
	"context"
	"testing"

	"ent-grpc-example/ent/category"
	"ent-grpc-example/ent/enttest"
	"ent-grpc-example/ent/proto/entpb"
	"ent-grpc-example/ent/user"
)

func TestUserProto(t *testing.T) {

	user := entpb.User{
		Name:         "rotemtam",
		EmailAddress: "rotemtam@example.com",
	}
	if user.GetName() != "rotemtam" {
		t.Fatal("expected user name to be rotemtam")
	}
	if user.GetEmailAddress() != "rotemtam@example.com" {
		t.Fatal("expected email address to be rotemtam@example.com")
	}
}

func TestServiceWithEdges(t *testing.T) {

	// start by initializing an ent client connected to an in memory sqlite instance
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	svc := entpb.NewUserService(client)

	cat := client.Category.Create().SetName("cat_1").SaveX(ctx)

	create, err := svc.Create(ctx, &entpb.CreateUserRequest{
		User: &entpb.User{
			Name:         "user",
			EmailAddress: "user@service.code",
			Administered: []*entpb.Category{
				{Id: int64(cat.ID)},
			},
		},
	})
	if err != nil {
		t.Fatal("failed creating user using UserService", err)
	}

	count, err := client.Category.
		Query().
		Where(
			category.HasAdminWith(
				user.ID(int(create.Id)),
			),
		).
		Count(ctx)
	if err != nil {
		t.Fatal("failed counting categories admin by created user", err)
	}
	if count != 1 {
		t.Fatal("expected exactly one group to managed by the created user")
	}

}

func TestGet(t *testing.T) {
	// インメモリのsqliteインスタンスに接続されたentクライアントの初期化から始めます
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	// 次に、Userサービスを初期化します。 ここでは、実際にポートを開いてgRPCサーバーを作成するのではなく
	// ライブラリのコードを直接呼び出していることに注目してください。
	svc := entpb.NewUserService(client)

	// 次に、ユーザーとカテゴリーを作成し、そのユーザーをカテゴリーの管理者に設定します。
	user := client.User.Create().
		SetName("rotemtam").
		SetEmailAddress("r@entgo.io").
		SaveX(ctx)

	client.Category.Create().
		SetName("category").
		SetAdmin(user).
		SaveX(ctx)

	// next, retrieve the user without edge information
	get, err := svc.Get(ctx, &entpb.GetUserRequest{
		Id: int64(user.ID),
	})
	if err != nil {
		t.Fatal("failed retrieving the created user", err)
	}
	if len(get.Administered) != 0 {
		t.Fatal("by default edge information is not supposed to be retrieved")
	}

	// next, retrieve the user *WITH* edge information
	get, err = svc.Get(ctx, &entpb.GetUserRequest{
		Id:   int64(user.ID),
		View: entpb.GetUserRequest_WITH_EDGE_IDS,
	})
	if err != nil {
		t.Fatal("failed retrieving the created user", err)
	}
	if len(get.Administered) != 1 {
		t.Fatal("using WITH_EDGE_IDS edges should be returned")
	}
}
