package graphql

import (
	"time"

	"ozon-posts/internal/services"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/sirupsen/logrus"
)

func InitGraphQLServer(
	userService *services.UserService,
	postService *services.PostService,
	commentService *services.CommentService,
	logger *logrus.Logger,
) *handler.Server {
	resolver := NewResolver(userService, postService, commentService, logger)

	srv := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.Use(extension.Introspection{})

	logger.Info("GraphQL сервер инициализирован")
	return srv
}
