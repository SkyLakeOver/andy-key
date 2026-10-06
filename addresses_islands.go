package main

// Фаза 1 рефакторинга раздела «Адреса» (AK-2.2.0):
// Vue-острова + SSR-таблица. Этот файл содержит island-эндпоинты для адресов:
//   POST /api/reference/addresses/island          — создание адреса
//   PUT  /api/reference/addresses/island/{id}     — обновление адреса
//   DELETE /api/reference/addresses/island/{id}   — удаление адреса (с проверкой ссылочной целостности)
//   GET  /api/reference/addresses/edit-form/{id}  — HTML-фрагмент модалки редактирования для Vue-острова
//
// Единый контракт данных (все поля — строки, включая floor):
// street, building, type (cabinet|corridor|service_room), cabinet, corridor, floor,
// service_room, service_room_type (server|warehouse|office|garage|other), garage_number.
// Валидация и приведение типов выполняются только на сервере — общей функцией
// validateIslandAddressRequest для создания и обновления.

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// ==================== КОНТРАКТ ДАННЫХ ====================

// IslandAddressRequest — единый JSON-контракт острова адресов.
// ВАЖНО: все поля — строки (включая floor), чтобы исключить ошибки
// unmarshal string→int на границе клиент/сервер.
type IslandAddressRequest struct {
	Street          string `json:"street"`
	Building        string `json:"building"`
	Type            string `json:"type"`
	Cabinet         string `json:"cabinet"`
	Corridor        string `json:"corridor"`
	Floor           string `json:"floor"`
	ServiceRoom     string `json:"service_room"`
	ServiceRoomType string `json:"service_room_type"`
	GarageNumber    string `json:"garage_number"`
}

// addressColumns — ЕДИНЫЙ список колонок контракта адреса.
// Используется во всех SQL-запросах (SELECT списка, SELECT одного, INSERT, UPDATE),
// чтобы слои никогда не расходились по составу полей.
const addressColumns = "id, street, building, type, cabinet, corridor, floor, service_room, service_room_type, garage_number, created_at"

var validAddressTypes = map[string]bool{"cabinet": true, "corridor": true, "service_room": true}
var validServiceRoomTypes = map[string]bool{"server": true, "warehouse": true, "office": true, "garage": true, "other": true}

// validateIslandAddressRequest — общая серверная валидация для создания и обновления.
// Сбрасывает поля, не относящиеся к выбранному типу, нормализует этаж и возвращает
// человекочитаемые тексты ошибок.
func validateIslandAddressRequest(req *IslandAddressRequest) error {
	req.Street = strings.TrimSpace(req.Street)
	req.Building = strings.TrimSpace(req.Building)
	req.Cabinet = strings.TrimSpace(req.Cabinet)
	req.Corridor = strings.TrimSpace(req.Corridor)
	req.Floor = strings.TrimSpace(req.Floor)
	req.ServiceRoom = strings.TrimSpace(req.ServiceRoom)
	req.ServiceRoomType = strings.TrimSpace(req.ServiceRoomType)
	req.GarageNumber = strings.TrimSpace(req.GarageNumber)

	if req.Street == "" || req.Building == "" {
		return fmt.Errorf("улица и дом обязательны для заполнения")
	}
	if !validAddressTypes[req.Type] {
		return fmt.Errorf("некорректный тип адреса: выберите кабинет, коридор или служебное помещение")
	}

	// Сброс полей, не относящихся к выбранному типу
	switch req.Type {
	case "cabinet":
		req.Corridor = ""
		req.ServiceRoom = ""
		req.ServiceRoomType = ""
		req.GarageNumber = ""

		if req.Cabinet == "" {
			return fmt.Errorf("для типа «кабинет» необходимо указать номер кабинета")
		}
		if req.Floor != "" {
			floor, err := strconv.Atoi(req.Floor)
			if err != nil || floor < 0 {
				return fmt.Errorf("этаж должен быть целым неотрицательным числом")
			}
			req.Floor = strconv.Itoa(floor) // нормализация (убирает leading zeros/пробелы)
		}

	case "corridor":
		req.Cabinet = ""
		req.ServiceRoom = ""
		req.ServiceRoomType = ""
		req.GarageNumber = ""
		req.Floor = ""

		if req.Corridor == "" {
			return fmt.Errorf("для типа «коридор» необходимо указать наименование коридора")
		}

	case "service_room":
		req.Cabinet = ""
		req.Corridor = ""
		req.Floor = ""

		if !validServiceRoomTypes[req.ServiceRoomType] {
			return fmt.Errorf("выберите подтип служебного помещения: серверная, склад, офис, гараж или другое")
		}
		if req.ServiceRoomType == "garage" {
			req.ServiceRoom = ""
			if req.GarageNumber == "" {
				return fmt.Errorf("для гаража необходимо указать номер гаража")
			}
		} else {
			// server | warehouse | office | other
			req.GarageNumber = ""
			if req.ServiceRoomType == "other" && req.ServiceRoom == "" {
				return fmt.Errorf("для подтипа «другое» необходимо указать наименование помещения")
			}
		}
	}

	return nil
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

// writeIslandJSONError пишет JSON-ответ с ошибкой в человекочитаемом виде.
func writeIslandJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// isIslandAuthed проверяет сессионную куку (маршруты регистрируются под authMiddleware,
// но дополнительная проверка защищает от случайного вызова вне цепочки).
func isIslandAuthed(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	return err == nil && CheckSession(cookie.Value)
}

// idFromPath извлекает и проверяет числовой ID из префикса пути.
func idFromPath(path, prefix string) (int, error) {
	idStr := strings.TrimPrefix(path, prefix)
	id, err := strconv.Atoi(idStr)
	if err != nil || !validateID(id) {
		return 0, fmt.Errorf("некорректный ID адреса")
	}
	return id, nil
}

// addressFloorSQLArg возвращает значение для колонки floor: строку этажа либо NULL.
func addressFloorSQLArg(floor string) string {
	if floor == "" {
		return "NULL"
	}
	return floor
}

// addressUniqueColumns — ЕДИНЫЙ список UNIQUE-колонок контракта адреса
// (включая floor, который входит в ветку сборки SQL, но НЕ входит в индекс
// UNIQUE(street, building, cabinet, corridor, service_room)).
// Порядок задаётся здесь один раз и используется и для INSERT, и для UPDATE,
// чтобы списки колонок никогда не дублировались и не расходились.
var addressUniqueColumns = []string{"cabinet", "corridor", "service_room"}

// buildAddressSQL собирает SQL по одному канону колонок addressUniqueColumns + floor
// плюс полные колонки типа (service_room_type, garage_number).
// operation: "insert" | "update".
// Пустые значения UNIQUE-колонок (и пустой этаж) подставляются в SQL ГОЛЫМ
// SQL-литералом NULL (без кавычек), аргументы исключаются из списка; непустые —
// прежним путём через плейсхолдеры ?. Причина: SQLite считает NULL различными,
// а пустые строки — равными, поэтому два гаража (или безымянные служебные разных
// подтипов) на одном адресе коллидируют по UNIQUE(street, building, cabinet,
// corridor, service_room), если неиспользуемые колонки хранить как ''.
// ВАЖНО (стоящее правило): строка "NULL" никогда не передаётся аргументом —
// только как литерал внутри собранного SQL.
func buildAddressSQL(operation string, req *IslandAddressRequest) (string, []interface{}) {
	cols := make([]string, 0, len(addressUniqueColumns)+3)
	cols = append(cols, addressUniqueColumns...)
	cols = append(cols, "floor", "service_room_type", "garage_number")

	values := map[string]string{
		"cabinet":           req.Cabinet,
		"corridor":          req.Corridor,
		"service_room":      req.ServiceRoom,
		"floor":             addressFloorSQLArg(req.Floor),
		"service_room_type": req.ServiceRoomType,
		"garage_number":     req.GarageNumber,
	}

	args := make([]interface{}, 0, len(cols))
	exprs := make([]string, 0, len(cols)) // VALUES-выражения (insert) или col = expr (update)
	for _, c := range cols {
		v := values[c]
		if v == "" || v == "NULL" {
			// голодный SQL-литерал NULL, аргумент НЕ добавляется
			exprs = append(exprs, "NULL")
		} else {
			exprs = append(exprs, "?")
			args = append(args, v)
		}
	}

	var sb strings.Builder
	if operation == "insert" {
		sb.WriteString("INSERT INTO reference_addresses (street, building, type, ")
		sb.WriteString(strings.Join(cols, ", "))
		sb.WriteString(") VALUES (?, ?, ?, ")
		sb.WriteString(strings.Join(exprs, ", "))
		sb.WriteString(")")
		return sb.String(), args
	}

	// operation == "update": SET-пары вида col = NULL / col = ?
	sets := make([]string, 0, len(cols))
	for i, c := range cols {
		sets = append(sets, c+" = "+exprs[i])
	}
	sb.WriteString("UPDATE reference_addresses SET street = ?, building = ?, type = ?, ")
	sb.WriteString(strings.Join(sets, ", "))
	sb.WriteString(" WHERE id = ?")
	return sb.String(), args
}

// normalizeUniqueEmptyStringsToNULL — разовая нормализация легаси-строк: старые
// записи могли хранить неиспользуемые колонки как '' (пустые строки равны в
// UNIQUE, NULL — различны). При первом обращении к островам адресов '' в
// cabinet/corridor/service_room приводятся к NULL (семантика контракта Фазы 1).
// Схему (UNIQUE-индексы) НЕ трогаем — дефект схемы является отдельной задачей.
var addressNormalizeReady sync.Once

func normalizeUniqueEmptyStringsToNULL() {
	addressNormalizeReady.Do(func() {
		for _, col := range []string{"cabinet", "corridor", "service_room"} {
			if err := execSafe("UPDATE reference_addresses SET " + col + " = NULL WHERE " + col + " = ''"); err != nil {
				log.Printf("Нормализация '' → NULL в %s не выполнена: %v", col, err)
			}
		}
	})
}

// execAddressInsert выполняет INSERT по единому канону колонок контракта.
// Пустые cabinet/corridor/service_room и пустой этаж вставляются как голодные
// SQL-литералы NULL (см. buildAddressSQL); непустые — через плейсхолдеры.
// service_room_type/garage_number — обычные данные, всегда через плейсхолдеры.
func execAddressInsert(req *IslandAddressRequest) error {
	normalizeUniqueEmptyStringsToNULL()
	sqlStr, tail := buildAddressSQL("insert", req)
	all := []interface{}{req.Street, req.Building, req.Type}
	all = append(all, tail...)
	return execSafe(sqlStr, all...)
}

// execAddressUpdate выполняет UPDATE по единому канону колонок контракта.
// Пустые cabinet/corridor/service_room и пустой этаж обновляются на голодные
// SQL-литералы NULL (см. buildAddressSQL); непустые — через плейсхолдеры.
func execAddressUpdate(req *IslandAddressRequest, id int) error {
	normalizeUniqueEmptyStringsToNULL()
	sqlStr, tail := buildAddressSQL("update", req)
	all := []interface{}{req.Street, req.Building, req.Type}
	all = append(all, tail...)
	all = append(all, id)
	return execSafe(sqlStr, all...)
}

// checkAddressUsage проверяет ссылочную целостность: используется ли адрес
// в справочниках сотрудников, АРМ и хостов. Возвращает текст ошибки или "".
func checkAddressUsage(id int) string {
	type usage struct {
		table string
		label string
		used  int
	}
	uses := []usage{
		{table: "reference_employees", label: "сотрудниках"},
		{table: "reference_workstations", label: "АРМ"},
		{table: "reference_hosts", label: "хостах"},
	}
	for i := range uses {
		n, err := querySingleInt("SELECT id FROM "+uses[i].table+" WHERE address_id = ? LIMIT 1", id)
		if err != nil && err.Error() != "sql: no rows in result set" {
			log.Printf("Ошибка проверки использования адреса %d в %s: %v", id, uses[i].table, err)
		}
		uses[i].used = n
	}
	var found []string
	for _, u := range uses {
		if u.used > 0 {
			found = append(found, u.label)
		}
	}
	if len(found) > 0 {
		return "Нельзя удалить адрес: он используется в справочниках — " + strings.Join(found, ", ")
	}
	return ""
}

// isUniqueViolation — проверка, что ошибка БД вызвана нарушением UNIQUE-ограничения.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint")
}

// isAddressDuplicate — серверная предпроверка дублей адреса.
// Предпроверка дублей: UNIQUE-индекс не ловит дубли с NULL-колонками
// (SQLite считает NULL различными), поэтому повтор кабинета/коридора с теми же
// пустыми неиспользуемыми полями нужно отлавливать до вставки.
// Сравнение по всем колонкам контракта: непустые значения сопоставляются через
// плейсхолдеры ?, пустые — через IS NULL (голый SQL-литерал NULL, НЕ аргумент).
// excludeID > 0 (для UPDATE) исключает саму редактируемую запись.
func isAddressDuplicate(req *IslandAddressRequest, excludeID int) bool {
	conds := []string{"street = ?", "building = ?", "type = ?"}
	args := []interface{}{req.Street, req.Building, req.Type}

	pairs := []struct {
		col    string
		value  string
	}{
		{"cabinet", req.Cabinet},
		{"corridor", req.Corridor},
		{"service_room", req.ServiceRoom},
		{"service_room_type", req.ServiceRoomType},
		{"garage_number", req.GarageNumber},
	}
	for _, p := range pairs {
		if p.value != "" {
			conds = append(conds, p.col+" = ?")
			args = append(args, p.value)
		} else {
			conds = append(conds, p.col+" IS NULL")
		}
	}

	// этаж: CHECK-колонка, пустой хранится как NULL (см. addressFloorSQLArg)
	if f := addressFloorSQLArg(req.Floor); f == "NULL" {
		conds = append(conds, "floor IS NULL")
	} else {
		conds = append(conds, "floor = ?")
		args = append(args, f)
	}

	if excludeID > 0 {
		conds = append(conds, "id != ?")
		args = append(args, excludeID)
	}

	query := "SELECT id FROM reference_addresses WHERE " + strings.Join(conds, " AND ")
	rows, err := querySafe(query, args...)
	if err != nil {
		log.Printf("Предпроверка дублей адреса не выполнена: %v", err)
		return false
	}
	return len(rows) >= 1
}

// decodeIslandAddressRequest декодирует JSON-тело запроса острова.
func decodeIslandAddressRequest(w http.ResponseWriter, r *http.Request) (IslandAddressRequest, bool) {
	var req IslandAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, "Некорректный формат JSON: "+err.Error())
		return req, false
	}
	if err := validateIslandAddressRequest(&req); err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, err.Error())
		return req, false
	}
	return req, true
}

// ==================== ENDPOINT: POST /api/reference/addresses/island ====================

func apiAddressIslandCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeIslandJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}
	if !isIslandAuthed(r) {
		writeIslandJSONError(w, http.StatusUnauthorized, "Требуется аутентификация")
		return
	}

	req, ok := decodeIslandAddressRequest(w, r)
	if !ok {
		return
	}

	// Предпроверка дублей: UNIQUE-индекс не ловит дубли с NULL-колонками.
	if isAddressDuplicate(&req, 0) {
		writeIslandJSONError(w, http.StatusBadRequest, "Такой адрес уже существует")
		return
	}

	if err := execAddressInsert(&req); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			writeIslandJSONError(w, http.StatusBadRequest, "Такой адрес уже существует")
			return
		}
		log.Printf("Ошибка создания адреса (island): %v", err)
		writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при создании адреса")
		return
	}

	logPanel("Добавлен адрес (Vue-остров): " + FormatFullAddress(req.Street, req.Building, req.Cabinet, req.Corridor, req.ServiceRoom, req.GarageNumber, req.Floor))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// ==================== ENDPOINT: /api/reference/addresses/island/{id} (PUT, DELETE) =====

func apiAddressIslandItemHandler(w http.ResponseWriter, r *http.Request) {
	if !isIslandAuthed(r) {
		writeIslandJSONError(w, http.StatusUnauthorized, "Требуется аутентификация")
		return
	}

	id, err := idFromPath(r.URL.Path, "/api/reference/addresses/island/")
	if err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodPut:
		req, ok := decodeIslandAddressRequest(w, r)
		if !ok {
			return
		}
		// Предпроверка дублей: UNIQUE-индекс не ловит дубли с NULL-колонками.
		// excludeID = id — смена адреса на собственную запись не считается дублем.
		if isAddressDuplicate(&req, id) {
			writeIslandJSONError(w, http.StatusBadRequest, "Такой адрес уже существует")
			return
		}
		if err := execAddressUpdate(&req, id); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				writeIslandJSONError(w, http.StatusBadRequest, "Такой адрес уже существует")
				return
			}
			log.Printf("Ошибка обновления адреса %d (island): %v", id, err)
			writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при обновлении адреса")
			return
		}
		logPanel("Обновлён адрес с ID " + strconv.Itoa(id) + " (Vue-остров)")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case http.MethodDelete:
		if usageErr := checkAddressUsage(id); usageErr != "" {
			writeIslandJSONError(w, http.StatusBadRequest, usageErr)
			return
		}
		if err := execSafe("DELETE FROM reference_addresses WHERE id = ?", id); err != nil {
			log.Printf("Ошибка удаления адреса %d (island): %v", id, err)
			writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при удалении адреса")
			return
		}
		logPanel("Удалён адрес с ID " + strconv.Itoa(id) + " (Vue-остров)")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		writeIslandJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	}
}

// ==================== ENDPOINT: GET /api/reference/addresses/edit-form/{id} =============

// Возвращает HTML-фрагмент для Vue-острова AddressEditModal: div с
// data-island="address-edit" и data-initial-data — JSON записи, ЭКРАНИРОВАННЫЙ
// для вставки в HTML-атрибут (кавычки внутри JSON иначе сломали бы атрибут).
func apiAddressEditFormHTMLHandler(w http.ResponseWriter, r *http.Request) {
	if !isIslandAuthed(r) {
		http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	id, err := idFromPath(r.URL.Path, "/api/reference/addresses/edit-form/")
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	rows, err := QueryDB("SELECT "+addressColumns+" FROM reference_addresses WHERE id = ?", strconv.Itoa(id))
	if err != nil || len(rows) == 0 {
		http.Error(w, "Адрес не найден", http.StatusNotFound)
		return
	}
	row := rows[0]

	payload := map[string]string{
		"id":                row["id"],
		"street":            row["street"],
		"building":          row["building"],
		"type":              row["type"],
		"cabinet":           row["cabinet"],
		"corridor":          row["corridor"],
		"floor":             row["floor"], // строка в обе стороны контракта
		"service_room":      row["service_room"],
		"service_room_type": row["service_room_type"],
		"garage_number":     row["garage_number"],
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Ошибка сериализации данных адреса", http.StatusInternalServerError)
		return
	}

	// Экранирование для HTML-атрибута: & < > " '
	escapedData := html.EscapeString(string(jsonBytes))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div data-island="address-edit" data-initial-data='%s'></div>`, escapedData)
}

// ==================== РЕГИСТРАЦИЯ МАРШРУТОВ ============================================

// registerAddressIslandRoutes регистрирует island-маршруты раздела «Адреса».
// Вызывается из setupRoutes после authMiddleware. Go-мультиплексор выбирает более
// специфичный паттерн ("/api/reference/addresses/island" и ".../island/" раньше
// префикса ".../"), поэтому старый JSON API (Phase 3) остаётся нетронутым.
func registerAddressIslandRoutes(authMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// Создание адреса (Vue-остров)
	http.HandleFunc("/api/reference/addresses/island", recoverMiddleware(authMiddleware(apiAddressIslandCreateHandler)))
	// Обновление/удаление адреса по ID (Vue-остров)
	http.HandleFunc("/api/reference/addresses/island/", recoverMiddleware(authMiddleware(apiAddressIslandItemHandler)))
	// HTML-фрагмент формы редактирования (Vue-остров) — заменяет старый шаблонный фрагмент
	http.HandleFunc("/api/reference/addresses/edit-form/", recoverMiddleware(authMiddleware(apiAddressEditFormHTMLHandler)))
}
