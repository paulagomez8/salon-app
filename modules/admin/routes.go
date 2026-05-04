package admin

import "github.com/fasthttp/router"

func RegisterRoutes(r *router.Router) {
	r.GET("/admin", PanelAdmin)
	r.GET("/admin/login", PaginaLogin)
	r.POST("/admin/login", Login)
	r.GET("/admin/logout", Logout)
	r.GET("/admin/precios", ListarPrecios)
	r.PUT("/admin/precios/{id}", ActualizarPrecio)
	r.GET("/precios", ListarPreciosPublicos)
	r.GET("/admin/config", ObtenerConfig)
	r.PUT("/admin/config", ActualizarConfig)
}
