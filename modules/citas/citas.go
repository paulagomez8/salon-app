package citas

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"salon-app/db"
	"time"

	"github.com/valyala/fasthttp"
)

type Cita struct {
	ID         int    `json:"id"`
	ClienteID  int    `json:"cliente_id"`
	EmpleadaID int    `json:"empleada_id"`
	ServicioID int    `json:"servicio_id"`
	FechaHora  string `json:"fecha_hora"`
	Estado     string `json:"estado"`
	Notas      string `json:"notas"`
	// Datos relacionados
	ClienteNombre         string `json:"cliente_nombre"`
	EmpleadaNombre        string `json:"empleada_nombre"`
	ClienteTelefono       string `json:"cliente_telefono"`
	ServicioNombre        string `json:"servicio_nombre"`
	DuracionMinutos       int    `json:"duracion_minutos"`
	DuracionActivaMinutos int    `json:"duracion_activa_minutos"`
	PermiteParalelo       bool   `json:"permite_paralelo"`
}

// citaExistente representa una cita ya agendada con los datos necesarios para validar conflictos
type citaExistente struct {
	inicio          int // en minutos desde medianoche
	duracionTotal   int
	duracionActiva  int
	permiteParalelo bool
}

// hayConflicto determina si un slot propuesto [slotInicio, slotInicio+duracionNueva]
// entra en conflicto con una cita existente, considerando paralelismo.
func hayConflicto(slotInicio, duracionNueva int, existente citaExistente) bool {
	slotFin := slotInicio + duracionNueva
	exFin := existente.inicio + existente.duracionTotal

	// Si no se solapan en absoluto, no hay conflicto
	if slotInicio >= exFin || slotFin <= existente.inicio {
		return false
	}

	// Se solapan. Si la cita existente no permite paralelo, hay conflicto directo.
	if !existente.permiteParalelo {
		return true
	}

	// La cita existente permite paralelo: solo hay conflicto si el slot nuevo
	// se solapa con la ventana activa de la cita existente [inicio, inicio+duracionActiva].
	ventanaActivaFin := existente.inicio + existente.duracionActiva
	if slotInicio < ventanaActivaFin && slotFin > existente.inicio {
		return true
	}

	// El solapamiento cae solo en la ventana de espera química → sin conflicto
	return false
}

// obtenerCitasExistentes devuelve las citas del día para la empleada dada,
// excluyendo canceladas y opcionalmente una cita por ID (útil para ediciones futuras).
func obtenerCitasExistentes(fecha string, empleadaID int, excluirID int) ([]citaExistente, error) {
	rows, err := db.DB.Query(`
		SELECT TIME_FORMAT(c.fecha_hora, '%H:%i'),
		       se.duracion_minutos,
		       se.duracion_activa_minutos,
		       se.permite_paralelo
		FROM citas c
		JOIN servicios se ON c.servicio_id = se.id
		WHERE DATE(c.fecha_hora) = ?
		  AND c.empleada_id = ?
		  AND c.estado != 'cancelada'
		  AND c.id != ?
	`, fecha, empleadaID, excluirID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resultado []citaExistente
	for rows.Next() {
		var horaStr string
		var durTotal, durActiva int
		var paralelo bool
		rows.Scan(&horaStr, &durTotal, &durActiva, &paralelo)
		resultado = append(resultado, citaExistente{
			inicio:          horaAMinutos(horaStr),
			duracionTotal:   durTotal,
			duracionActiva:  durActiva,
			permiteParalelo: paralelo,
		})
	}
	return resultado, nil
}

// obtenerBloqueos devuelve los bloqueos activos del día como rangos en minutos.
func obtenerBloqueos(fecha string) ([][2]int, error) {
	diaSemana := 0
	if t, err := time.Parse("2006-01-02", fecha); err == nil {
		d := int(t.Weekday())
		if d == 0 {
			d = 7
		}
		diaSemana = d
	}

	rows, err := db.DB.Query(`
		SELECT TIME_FORMAT(hora_inicio, '%H:%i'), TIME_FORMAT(hora_fin, '%H:%i')
		FROM bloqueos
		WHERE activo = 1 AND (fecha = ? OR dia_semana = ?)
	`, fecha, diaSemana)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bloqueos [][2]int
	for rows.Next() {
		var ini, fin string
		rows.Scan(&ini, &fin)
		bloqueos = append(bloqueos, [2]int{horaAMinutos(ini), horaAMinutos(fin)})
	}
	return bloqueos, nil
}

func ListarCitas(ctx *fasthttp.RequestCtx) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.cliente_id, c.empleada_id, c.servicio_id,
		       c.fecha_hora, c.estado, c.notas,
		       cl.nombre, cl.telefono, em.nombre, se.nombre,
		       se.duracion_minutos, se.duracion_activa_minutos, se.permite_paralelo
		FROM citas c
		JOIN clientes cl ON c.cliente_id = cl.id
		JOIN empleadas em ON c.empleada_id = em.id
		JOIN servicios se ON c.servicio_id = se.id
		ORDER BY c.fecha_hora ASC
	`)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al obtener citas")
		return
	}
	defer rows.Close()

	var lista []Cita
	for rows.Next() {
		var c Cita
		rows.Scan(
			&c.ID, &c.ClienteID, &c.EmpleadaID, &c.ServicioID,
			&c.FechaHora, &c.Estado, &c.Notas,
			&c.ClienteNombre, &c.ClienteTelefono, &c.EmpleadaNombre, &c.ServicioNombre,
			&c.DuracionMinutos, &c.DuracionActivaMinutos, &c.PermiteParalelo,
		)
		lista = append(lista, c)
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(lista)
}

func CrearCita(ctx *fasthttp.RequestCtx) {
	var c Cita
	if err := json.Unmarshal(ctx.PostBody(), &c); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}

	// Obtener duración del servicio nuevo
	var durNueva, durActivaNueva int
	err := db.DB.QueryRow(
		"SELECT duracion_minutos, duracion_activa_minutos FROM servicios WHERE id = ?",
		c.ServicioID,
	).Scan(&durNueva, &durActivaNueva)
	if err != nil {
		ctx.SetStatusCode(404)
		ctx.SetBodyString("Servicio no encontrado")
		return
	}

	// Extraer fecha y hora del campo fecha_hora (formato "2006-01-02T15:04" o "2006-01-02 15:04")
	fechaHoraStr := c.FechaHora
	var fecha, horaStr string
	if len(fechaHoraStr) >= 16 {
		fecha = fechaHoraStr[:10]
		horaStr = fechaHoraStr[11:16]
	} else {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Formato de fecha_hora inválido")
		return
	}

	slotInicio := horaAMinutos(horaStr)

	// Obtener citas existentes del día
	existentes, err := obtenerCitasExistentes(fecha, c.EmpleadaID, 0)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al verificar disponibilidad")
		return
	}

	// Validar conflictos con citas existentes
	for _, ex := range existentes {
		if hayConflicto(slotInicio, durNueva, ex) {
			ctx.SetStatusCode(409)
			ctx.SetBodyString("El horario no está disponible")
			return
		}
	}

	// Validar contra bloqueos
	bloqueos, err := obtenerBloqueos(fecha)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al verificar bloqueos")
		return
	}
	slotFin := slotInicio + durNueva
	for _, b := range bloqueos {
		if slotInicio < b[1] && slotFin > b[0] {
			ctx.SetStatusCode(409)
			ctx.SetBodyString("El horario está bloqueado")
			return
		}
	}

	_, err = db.DB.Exec(
		`INSERT INTO citas (cliente_id, empleada_id, servicio_id, fecha_hora, estado, notas)
		 VALUES (?, ?, ?, ?, 'pendiente', ?)`,
		c.ClienteID, c.EmpleadaID, c.ServicioID, c.FechaHora, c.Notas,
	)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al crear cita")
		return
	}

	ctx.SetStatusCode(201)
	ctx.SetBodyString("Cita creada")
}

func ReservarCita(ctx *fasthttp.RequestCtx) {
	var body struct {
		Nombre     string `json:"nombre"`
		Telefono   string `json:"telefono"`
		Email      string `json:"email"`
		ServicioID int    `json:"servicio_id"`
		FechaHora  string `json:"fecha_hora"`
		Notas      string `json:"notas"`
	}

	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		log.Println("ERROR JSON:", err)
		ctx.SetStatusCode(400)
		ctx.SetBodyString(err.Error())
		return
	}

	// Obtener duración del servicio
	var durNueva int
	err := db.DB.QueryRow(
		"SELECT duracion_minutos FROM servicios WHERE id = ?",
		body.ServicioID,
	).Scan(&durNueva)

	if err != nil {
		log.Println("ERROR SERVICIO:", err)
		ctx.SetStatusCode(404)
		ctx.SetBodyString(err.Error())
		return
	}

	// Extraer fecha y hora
	var fecha, horaStr string
	if len(body.FechaHora) >= 16 {
		fecha = body.FechaHora[:10]
		horaStr = body.FechaHora[11:16]
	} else {
		log.Println("ERROR FECHA FORMATO:", body.FechaHora)
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Formato de fecha_hora inválido")
		return
	}

	slotInicio := horaAMinutos(horaStr)
	slotFin := slotInicio + durNueva

	const empleadaID = 1

	// Validar conflictos
	existentes, err := obtenerCitasExistentes(fecha, empleadaID, 0)
	if err != nil {
		log.Println("ERROR CITAS EXISTENTES:", err)
		ctx.SetStatusCode(500)
		ctx.SetBodyString(err.Error())
		return
	}

	for _, ex := range existentes {
		if hayConflicto(slotInicio, durNueva, ex) {
			ctx.SetStatusCode(409)
			ctx.SetBodyString("El horario no está disponible")
			return
		}
	}

	// Validar bloqueos
	bloqueos, err := obtenerBloqueos(fecha)
	if err != nil {
		log.Println("ERROR BLOQUEOS:", err)
		ctx.SetStatusCode(500)
		ctx.SetBodyString(err.Error())
		return
	}

	for _, b := range bloqueos {
		if slotInicio < b[1] && slotFin > b[0] {
			ctx.SetStatusCode(409)
			ctx.SetBodyString("El horario está bloqueado")
			return
		}
	}

	// Buscar o crear cliente
	var clienteID int
	err = db.DB.QueryRow(
		"SELECT id FROM clientes WHERE telefono = ?",
		body.Telefono,
	).Scan(&clienteID)

	if err != nil {
		if err == sql.ErrNoRows {
			res, err := db.DB.Exec(
				"INSERT INTO clientes (nombre, telefono, email) VALUES (?, ?, ?)",
				body.Nombre, body.Telefono, body.Email,
			)
			if err != nil {
				log.Println("ERROR INSERT CLIENTE:", err)
				ctx.SetStatusCode(500)
				ctx.SetBodyString(err.Error())
				return
			}
			id, _ := res.LastInsertId()
			clienteID = int(id)
		} else {
			log.Println("ERROR BUSCAR CLIENTE:", err)
			ctx.SetStatusCode(500)
			ctx.SetBodyString(err.Error())
			return
		}
	}

	log.Println("DEBUG INSERT:", clienteID, body.ServicioID, body.FechaHora)

	_, err = db.DB.Exec(
		`INSERT INTO citas (cliente_id, empleada_id, servicio_id, fecha_hora, estado, notas)
	 VALUES (?, ?, ?, ?, ?, ?)`,
		clienteID,
		empleadaID,
		body.ServicioID,
		body.FechaHora,
		"confirmada",
		body.Notas,
	)

	if err != nil {
		log.Println("🔥 ERROR MYSQL REAL:", err)
		ctx.SetStatusCode(500)
		ctx.SetBodyString(err.Error())
		return
	}
	ctx.SetStatusCode(201)
	ctx.SetBodyString("Cita creada")
}

func ActualizarEstado(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)

	var body struct {
		Estado string `json:"estado"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("Datos inválidos")
		return
	}

	_, err := db.DB.Exec("UPDATE citas SET estado = ? WHERE id = ?", body.Estado, id)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al actualizar estado")
		return
	}

	ctx.SetBodyString("Estado actualizado")
}

func ConsultarDisponibilidad(ctx *fasthttp.RequestCtx) {
	fecha := string(ctx.QueryArgs().Peek("fecha"))
	servicioID := string(ctx.QueryArgs().Peek("servicio_id"))

	// Obtener duración del servicio solicitado
	var duracion int
	err := db.DB.QueryRow(
		"SELECT duracion_minutos FROM servicios WHERE id = ?", servicioID,
	).Scan(&duracion)
	if err != nil {
		ctx.SetStatusCode(404)
		ctx.SetBodyString("Servicio no encontrado")
		return
	}

	// Empleada fija = 1 (Clémence) — escalar cuando haya más empleadas
	const empleadaID = 1

	existentes, err := obtenerCitasExistentes(fecha, empleadaID, 0)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al consultar citas")
		return
	}

	bloqueos, err := obtenerBloqueos(fecha)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("Error al consultar bloqueos")
		return
	}

	// Generar slots disponibles cada 15 min entre 08:00 y 19:00
	inicio := 8 * 60
	cierre := 19 * 60
	var disponibles []string

	for slot := inicio; slot+duracion <= cierre; slot += 15 {
		libre := true

		// Verificar contra citas existentes (con lógica paralela)
		for _, ex := range existentes {
			if hayConflicto(slot, duracion, ex) {
				libre = false
				break
			}
		}

		// Verificar contra bloqueos
		if libre {
			slotFin := slot + duracion
			for _, b := range bloqueos {
				if slot < b[1] && slotFin > b[0] {
					libre = false
					break
				}
			}
		}

		if libre {
			disponibles = append(disponibles, minutosAHora(slot))
		}
	}

	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(disponibles)
}

func horaAMinutos(h string) int {
	var hh, mm int
	fmt.Sscanf(h, "%d:%d", &hh, &mm)
	return hh*60 + mm
}

func minutosAHora(m int) string {
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}
