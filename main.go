package main

import (
	"log"
	"salon-app/db"
	"salon-app/modules/admin"
	"salon-app/modules/bloqueos"
	"salon-app/modules/citas"
	"salon-app/modules/clientes"
	"salon-app/modules/empleadas"
	"salon-app/modules/servicios"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {
	db.Connect()

	r := router.New()

	r.ServeFiles("/static/{filepath:*}", "static")

	admin.RegisterRoutes(r)
	bloqueos.RegisterRoutes(r)
	citas.RegisterRoutes(r)
	clientes.RegisterRoutes(r)
	servicios.RegisterRoutes(r)
	empleadas.RegisterRoutes(r)

	r.GET("/agenda", func(ctx *fasthttp.RequestCtx) {
		ctx.SendFile("templates/agenda.html")
	})

	r.GET("/", func(ctx *fasthttp.RequestCtx) {
		ctx.SendFile("templates/index.html")
	})

	r.GET("/reservar", func(ctx *fasthttp.RequestCtx) {
		ctx.SendFile("templates/reservas.html")
	})

	log.Println("🚀 Servidor corriendo en http://localhost:8081")
	log.Fatal(fasthttp.ListenAndServe(":8081", r.Handler))
}
