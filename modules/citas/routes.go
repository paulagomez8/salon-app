package citas

import "github.com/fasthttp/router"

func RegisterRoutes(r *router.Router) {
	r.GET("/citas", ListarCitas)
	r.POST("/citas", CrearCita)
	r.POST("/reservar", ReservarCita)
	r.PUT("/citas/{id}/estado", ActualizarEstado)
	r.GET("/disponibilidad", ConsultarDisponibilidad)
}
