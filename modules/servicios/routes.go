package servicios

import "github.com/fasthttp/router"

func RegisterRoutes(r *router.Router) {
	r.GET("/servicios", ListarServicios)
	r.POST("/servicios", CrearServicio)
	r.DELETE("/servicios/{id}", EliminarServicio)
	r.PUT("/servicios/{id}", ActualizarServicio)
}
