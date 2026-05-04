package bloqueos

import "github.com/fasthttp/router"

func RegisterRoutes(r *router.Router) {
	r.GET("/bloqueos", ListarBloqueos)
	r.POST("/bloqueos", CrearBloqueo)
	r.DELETE("/bloqueos/{id}", EliminarBloqueo)
}
