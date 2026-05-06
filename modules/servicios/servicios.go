package servicios

import (
	"encoding/json"
	"salon-app/db"

	"github.com/valyala/fasthttp"
)

type Servicio struct {
	ID                    int     `json:"id"`
	Nombre                string  `json:"nombre"`
	Descripcion           string  `json:"descripcion"`
	DuracionMinutos       int     `json:"duracion_minutos"`
	DuracionActivaMinutos int     `json:"duracion_activa_minutos"`
	PermiteParalelo       bool    `json:"permite_paralelo"`
	Precio                float64 `json:"precio"`
	Activo                bool    `json:"activo"`
}

func ListarServicios(ctx *fasthttp.RequestCtx) {
	rows, err := db.DB.Query("SELECT id, nombre, descripcion, duracion_minutos, duracion_activa_minutos, permite_paralelo, precio, activo FROM servicios WHERE activo = 1")
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al obtener servicios")
		return
	}
	defer rows.Close()

	var lista []Servicio
	for rows.Next() {
		var s Servicio
		rows.Scan(&s.ID, &s.Nombre, &s.Descripcion, &s.DuracionMinutos, &s.DuracionActivaMinutos, &s.PermiteParalelo, &s.Precio, &s.Activo)
		lista = append(lista, s)
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(lista)
}

func CrearServicio(ctx *fasthttp.RequestCtx) {
	var s Servicio
	if err := json.Unmarshal(ctx.PostBody(), &s); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos: " + err.Error())
		return
	}

	_, err := db.DB.Exec(
		"INSERT INTO servicios (nombre, descripcion, duracion_minutos, duracion_activa_minutos, permite_paralelo, precio) VALUES (?, ?, ?, ?, ?, ?)",
		s.Nombre, s.Descripcion, s.DuracionMinutos, s.DuracionActivaMinutos, s.PermiteParalelo, s.Precio,
	)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al crear servicio")
		return
	}

	ctx.SetStatusCode(201)
	ctx.SetBodyString("Servicio creado")
}

func EliminarServicio(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	_, err := db.DB.Exec("UPDATE servicios SET activo = 0 WHERE id = ?", id)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al eliminar servicio")
		return
	}
	ctx.SetBodyString("Servicio eliminado")
}

func ActualizarServicio(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	var s Servicio
	if err := json.Unmarshal(ctx.PostBody(), &s); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}
	_, err := db.DB.Exec(
		"UPDATE servicios SET nombre=?, descripcion=?, duracion_minutos=?, precio=? WHERE id=?",
		s.Nombre, s.Descripcion, s.DuracionMinutos, s.Precio, id,
	)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al actualizar servicio")
		return
	}
	ctx.SetBodyString("Servicio actualizado")
}
