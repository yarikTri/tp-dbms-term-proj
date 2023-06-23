package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yarikTri/dbms-term-proj/configs"
	handler "github.com/yarikTri/dbms-term-proj/internal/app/delivery"
	repo "github.com/yarikTri/dbms-term-proj/internal/app/repository"
	usecase "github.com/yarikTri/dbms-term-proj/internal/app/usecase"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func applicationJSONMiddleware(_ *mux.Router) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	router := mux.NewRouter()

	connString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=%s",
		configs.PostgresConfig.User,
		configs.PostgresConfig.Password,
		configs.PostgresConfig.DB,
		configs.PostgresConfig.Port,
	)

	pgxConnConfig, err := pgx.ParseConnectionString(connString)
	if err != nil {
		log.Fatal(err.Error())
	}
	pgxConnConfig.PreferSimpleProtocol = true

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     pgxConnConfig,
		MaxConnections: 200,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	}

	pool, err := pgx.NewConnPool(poolConfig)
	if err != nil {
		log.Fatalf(err.Error())
	}

	repo := repo.NewPostgresAppRepository(pool)
	usecase := usecase.NewAppUseCase(repo)
	handler.NewAppHandler(router, usecase)

	router.Use(applicationJSONMiddleware(router))

	log.Fatal(fasthttp.ListenAndServe(":5000", fasthttpadaptor.NewFastHTTPHandler(router)))
}
