package bloqueos

import (
	"encoding/json"
	"salon-app/db"

	"github.com/valyala/fasthttp"
)

type Bloqueo struct {
	ID         int    `json:"id"`
	Fecha      string `json:"fecha"`       // "2026-04-10" o ""
	DiaSemana  int    `json:"dia_semana"`  // 1=Lunes..6=Sábado, 0 = no aplica
	HoraInicio string `json:"hora_inicio"` // "13:00"
	HoraFin    string `json:"hora_fin"`    // "14:00"
	Nota       string `json:"nota"`
	Activo     bool   `json:"activo"`
}

func ListarBloqueos(ctx *fasthttp.RequestCtx) {
	rows, err := db.DB.Query(`
		SELECT id, COALESCE(fecha, ''), COALESCE(dia_semana, 0),
		       hora_inicio, hora_fin, nota, activo
		FROM bloqueos WHERE activo = 1
		ORDER BY dia_semana, fecha, hora_inicio
	`)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al obtener bloqueos")
		return
	}
	defer rows.Close()

	var lista []Bloqueo
	for rows.Next() {
		var b Bloqueo
		rows.Scan(&b.ID, &b.Fecha, &b.DiaSemana, &b.HoraInicio, &b.HoraFin, &b.Nota, &b.Activo)
		lista = append(lista, b)
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(lista)
}

func CrearBloqueo(ctx *fasthttp.RequestCtx) {
	var b Bloqueo
	if err := json.Unmarshal(ctx.PostBody(), &b); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}

	var fechaVal interface{}
	var diaVal interface{}

	if b.Fecha != "" {
		fechaVal = b.Fecha
	}
	if b.DiaSemana != 0 {
		diaVal = b.DiaSemana
	}

	_, err := db.DB.Exec(`
		INSERT INTO bloqueos (fecha, dia_semana, hora_inicio, hora_fin, nota)
		VALUES (?, ?, ?, ?, ?)`,
		fechaVal, diaVal, b.HoraInicio, b.HoraFin, b.Nota,
	)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al crear bloqueo")
		return
	}

	ctx.SetStatusCode(201)
	ctx.SetBodyString("Bloqueo creado")
}

func EliminarBloqueo(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	_, err := db.DB.Exec("UPDATE bloqueos SET activo = 0 WHERE id = ?", id)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al eliminar bloqueo")
		return
	}
	ctx.SetBodyString("Bloqueo eliminado")
}
