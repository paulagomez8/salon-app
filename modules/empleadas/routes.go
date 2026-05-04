package empleadas

import "github.com/fasthttp/router"

func RegisterRoutes(r *router.Router) {
	r.GET("/empleadas", ListarEmpleadas)
	r.POST("/empleadas", CrearEmpleada)
	r.DELETE("/empleadas/{id}", EliminarEmpleada)
}
