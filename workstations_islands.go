package main

// Фаза 2 рефакторинга раздела «АРМ» (AK-2.2.0, портовая миграция на образец Фазы 1):
// Vue-острова + SSR-таблица. Этот файл содержит island-эндпоинты для АРМ:
//   POST   /api/reference/workstations/island          — создание АРМ
//   PUT    /api/reference/workstations/island/{id}     — обновление АРМ
//   DELETE /api/reference/workstations/island/{id}     — удаление АРМ (только существующий id)
//   GET    /api/reference/workstations/edit-form/{id}  — HTML-фрагмент модалки редактирования
//
// Единый контракт данных (все поля — строки):
// inventory_number, serial_number, seal_numbers, monitor_count ("1".."5"),
// employee_id ("" или число), address_id (число, обязателен), is_vacant ("1"/"0"),
// replacement_done ("1"/"0"), replacement_date, replacement_letter.
//
// Валидация и приведение типов выполняются только на сервере — общей функцией
// validateIslandWorkstationRequest для создания и обновления (Шаг 1.3).

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
)

// ==================== КОНТРАКТ ДАННЫХ ====================

// IslandWorkstationRequest — единый JSON-контракт острова АРМ.
// ВАЖНО: все поля — строки, чтобы исключить ошибки unmarshal на границе
// клиент/сервер (employee_id допускает "" — «не назначен»).
type IslandWorkstationRequest struct {
	InventoryNumber   string `json:"inventory_number"`
	SerialNumber      string `json:"serial_number"`
	SealNumbers       string `json:"seal_numbers"`
	MonitorCount      string `json:"monitor_count"`
	EmployeeID        string `json:"employee_id"`
	AddressID         string `json:"address_id"`
	IsVacant          string `json:"is_vacant"`
	ReplacementDone   string `json:"replacement_done"`
	ReplacementDate   string `json:"replacement_date"`
	ReplacementLetter string `json:"replacement_letter"`
}

// workstationColumns — ЕДИНЫЙ список колонок контракта АРМ.
// Используется во всех SQL-запросах (SELECT списка, SELECT одной записи,
// INSERT, UPDATE), чтобы слои никогда не расходились по составу полей.
const workstationColumns = "inventory_number, serial_number, seal_numbers, monitor_count, employee_id, address_id, is_vacant, replacement_done, replacement_date, replacement_letter"

// workstationColumnsPrefixed — тот же ЕДИНЫЙ список колонок с алиасом таблицы w
// (для SELECT списка/фильтров и SELECT одной записи в edit-form).
var workstationColumnsPrefixed = "w.id, w." + strings.ReplaceAll(workstationColumns, ", ", ", w.") + ", w.created_at"

// workstationJoins — единые JOIN для списка, фильтров и одной записи.
const workstationJoins = `FROM reference_workstations w
	LEFT JOIN reference_employees e ON w.employee_id = e.id
	LEFT JOIN reference_addresses a ON w.address_id = a.id`

// workstationListSelect — единый SELECT для списка и одной записи:
// ВСЕ колонки контракта плюс employee full_name, street, building (JOIN уже есть).
var workstationListSelect = "SELECT " + workstationColumnsPrefixed + ",\n" +
	"	e.full_name AS employee_name, a.street, a.building, a.cabinet, a.corridor, a.service_room, a.floor, a.garage_number\n" +
	"	" + workstationJoins

// ==================== СЕРВЕРНАЯ ВАЛИДАЦИЯ (единая для POST и PUT) ==========

// validateIslandWorkstationRequest — общая серверная валидация создания и
// обновления. Возвращает человекочитаемые тексты ошибок; нормализует
// monitor_count (целое 1..5), employee_id, булевы поля; при replacement_done="0"
// сбрасывает replacement_date и replacement_letter.
func validateIslandWorkstationRequest(req *IslandWorkstationRequest) error {
	req.InventoryNumber = strings.TrimSpace(req.InventoryNumber)
	req.SerialNumber = strings.TrimSpace(req.SerialNumber)
	req.SealNumbers = strings.TrimSpace(req.SealNumbers)
	req.MonitorCount = strings.TrimSpace(req.MonitorCount)
	req.EmployeeID = strings.TrimSpace(req.EmployeeID)
	req.AddressID = strings.TrimSpace(req.AddressID)
	req.IsVacant = strings.TrimSpace(req.IsVacant)
	req.ReplacementDone = strings.TrimSpace(req.ReplacementDone)
	req.ReplacementDate = strings.TrimSpace(req.ReplacementDate)
	req.ReplacementLetter = strings.TrimSpace(req.ReplacementLetter)

	if req.InventoryNumber == "" {
		return fmt.Errorf("инвентарный номер обязателен для заполнения")
	}

	// address_id — число, обязательно, должно существовать в БД
	if req.AddressID == "" {
		return fmt.Errorf("адрес обязателен для заполнения: выберите адрес из справочника")
	}
	addressID, err := strconv.Atoi(req.AddressID)
	if err != nil || !validateID(addressID) {
		return fmt.Errorf("некорректный адрес: значение должно быть положительным числом")
	}
	if _, err := querySingleInt("SELECT id FROM reference_addresses WHERE id = ? LIMIT 1", addressID); err != nil {
		return fmt.Errorf("выбранный адрес не найден в справочнике адресов")
	}

	// employee_id — "" (не назначен) либо существующее число
	if req.EmployeeID != "" {
		employeeID, err := strconv.Atoi(req.EmployeeID)
		if err != nil || !validateID(employeeID) {
			return fmt.Errorf("некорректный сотрудник: значение должно быть положительным числом или пустым")
		}
		if _, err := querySingleInt("SELECT id FROM reference_employees WHERE id = ? LIMIT 1", employeeID); err != nil {
			return fmt.Errorf("выбранный сотрудник не найден в справочнике сотрудников")
		}
	} else {
		// "" означает «не назначен» — нормализуем явственный ноль
		req.EmployeeID = ""
	}

	// monitor_count — целое 1..5 с нормализацией
	monitorCount, err := strconv.Atoi(req.MonitorCount)
	if err != nil || monitorCount < 1 || monitorCount > 5 {
		return fmt.Errorf("количество мониторов должно быть целым числом от 1 до 5")
	}
	req.MonitorCount = strconv.Itoa(monitorCount)

	// Булевые поля: нормализация "1"/"0"
	if req.IsVacant == "" {
		req.IsVacant = "0"
	}
	if req.IsVacant != "0" && req.IsVacant != "1" {
		return fmt.Errorf("некорректное значение вакантности: ожидается «да» или «нет»")
	}
	if req.ReplacementDone == "" {
		req.ReplacementDone = "0"
	}
	if req.ReplacementDone != "0" && req.ReplacementDone != "1" {
		return fmt.Errorf("некорректное значение признака замены: ожидается «да» или «нет»")
	}

	// При replacement_done="0" сбрасываем дату и букву замены
	if req.ReplacementDone == "0" {
		req.ReplacementDate = ""
		req.ReplacementLetter = ""
	}

	return nil
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

// sqlNullIfEmpty возвращает SQL-литерал NULL для пустой строки, иначе плейсхолдер "?".
func sqlNullIfEmpty(value string) string {
	if value == "" {
		return "NULL"
	}
	return "?"
}

// execWorkstationInsert выполняет INSERT по ЕДИНОМУ списку колонок контракта
// (workstationColumns). Пустые employee_id/replacement_date/replacement_letter
// записываются как литерал NULL, остальные значения — только плейсхолдеры "?".
func execWorkstationInsert(req *IslandWorkstationRequest) error {
	cols := workstationColumns
	placeholders, args := buildValueTuples(req)
	return execSafe(
		fmt.Sprintf("INSERT INTO reference_workstations (%s) VALUES (%s)", cols, placeholders),
		args...,
	)
}

// execWorkstationUpdate выполняет UPDATE по ЕДИНОМУ списку колонок контракта
// (workstationColumns): SET inventory_number = ?, ..., replacement_letter = ?
// либо NULL для пустых необязательных полей.
func execWorkstationUpdate(req *IslandWorkstationRequest, id int) error {
	setParts, args := buildSetClauses(req)
	args = append(args, id)
	return execSafe(
		"UPDATE reference_workstations SET "+strings.Join(setParts, ", ")+" WHERE id = ?",
		args...,
	)
}

// buildValueTuples возвращает кортеж значений для INSERT и плоский список
// аргументов в том же порядке, что и workstationColumns.
func buildValueTuples(req *IslandWorkstationRequest) (string, []interface{}) {
	values := workstationColumnValues(req)
	tuple := make([]string, len(values))
	for i, v := range values {
		tuple[i] = v.placeholder
	}
	return strings.Join(tuple, ", "), valuesArgs(values)
}

// buildSetClauses возвращает части "колонка = ?" (или "колонка = NULL") для
// UPDATE в порядке workstationColumns.
func buildSetClauses(req *IslandWorkstationRequest) ([]string, []interface{}) {
	values := workstationColumnValues(req)
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = v.column + " = " + v.placeholder
	}
	return parts, valuesArgs(values)
}

type columnValue struct {
	column      string
	placeholder string // "?" или "NULL"
	value       string
}

func valuesArgs(values []columnValue) []interface{} {
	var out []interface{}
	for _, v := range values {
		if v.placeholder == "?" {
			out = append(out, v.value)
		}
	}
	return out
}

// workstationColumnValues — ЕДИНАЯ таблица «колонка → значение» контракта АРМ,
// используемая и для INSERT, и для UPDATE (гарантия одного списка колонок).
func workstationColumnValues(req *IslandWorkstationRequest) []columnValue {
	return []columnValue{
		{"inventory_number", "?", req.InventoryNumber},
		{"serial_number", "?", req.SerialNumber},
		{"seal_numbers", "?", req.SealNumbers},
		{"monitor_count", "?", req.MonitorCount},
		{"employee_id", sqlNullIfEmpty(req.EmployeeID), req.EmployeeID},
		{"address_id", "?", req.AddressID},
		{"is_vacant", "?", req.IsVacant},
		{"replacement_done", "?", req.ReplacementDone},
		{"replacement_date", sqlNullIfEmpty(req.ReplacementDate), req.ReplacementDate},
		{"replacement_letter", sqlNullIfEmpty(req.ReplacementLetter), req.ReplacementLetter},
	}
}

// workstationExists проверяет существование записи АРМ по ID.
func workstationExists(id int) bool {
	_, err := querySingleInt("SELECT id FROM reference_workstations WHERE id = ? LIMIT 1", id)
	return err == nil
}

// decodeIslandWorkstationRequest декодирует JSON-тело запроса острова и
// прогоняет его через единую серверную валидацию.
func decodeIslandWorkstationRequest(w http.ResponseWriter, r *http.Request) (IslandWorkstationRequest, bool) {
	var req IslandWorkstationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, "Некорректный формат JSON: "+err.Error())
		return req, false
	}
	if err := validateIslandWorkstationRequest(&req); err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, err.Error())
		return req, false
	}
	return req, true
}

// ==================== ENDPOINT: POST /api/reference/workstations/island ============

func apiWorkstationIslandCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeIslandJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}
	if !isIslandAuthed(r) {
		writeIslandJSONError(w, http.StatusUnauthorized, "Требуется аутентификация")
		return
	}

	req, ok := decodeIslandWorkstationRequest(w, r)
	if !ok {
		return
	}

	if err := execWorkstationInsert(&req); err != nil {
		logDiagnostic("Ошибка создания АРМ (island): " + err.Error())
		writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при создании АРМ")
		return
	}

	logPanel("Добавлено АРМ (Vue-остров): инв. номер " + req.InventoryNumber)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// ==================== ENDPOINT: /api/reference/workstations/island/{id} (PUT, DELETE)

func apiWorkstationIslandItemHandler(w http.ResponseWriter, r *http.Request) {
	if !isIslandAuthed(r) {
		writeIslandJSONError(w, http.StatusUnauthorized, "Требуется аутентификация")
		return
	}

	id, err := idFromPath(r.URL.Path, "/api/reference/workstations/island/")
	if err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, "Некорректный ID АРМ")
		return
	}

	switch r.Method {
	case http.MethodPut:
		req, ok := decodeIslandWorkstationRequest(w, r)
		if !ok {
			return
		}
		if !workstationExists(id) {
			writeIslandJSONError(w, http.StatusNotFound, "АРМ с указанным ID не найден")
			return
		}
		if err := execWorkstationUpdate(&req, id); err != nil {
			logDiagnostic(fmt.Sprintf("Ошибка обновления АРМ %d (island): %v", id, err))
			writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при обновлении АРМ")
			return
		}
		logPanel("Обновлён АРМ с ID " + strconv.Itoa(id) + " (Vue-остров)")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case http.MethodDelete:
		// Проверки ссылочной целостности не требуются: таблица reference_workstations
		// не используется другими разделами. Но удаляем только существующий id.
		if !workstationExists(id) {
			writeIslandJSONError(w, http.StatusNotFound, "АРМ с указанным ID не найден")
			return
		}
		if err := execSafe("DELETE FROM reference_workstations WHERE id = ?", id); err != nil {
			logDiagnostic(fmt.Sprintf("Ошибка удаления АРМ %d (island): %v", id, err))
			writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при удалении АРМ")
			return
		}
		logPanel("Удалён АРМ с ID " + strconv.Itoa(id) + " (Vue-остров)")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		writeIslandJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	}
}

// ==================== ENDPOINT: GET /api/reference/workstations/edit-form/{id} =====

// Возвращает HTML-фрагмент для Vue-острова WorkstationEditModal: div с
// data-island="workstation-edit" и data-initial-data — JSON записи,
// ЭКРАНИРОВАННЫЙ для вставки в HTML-атрибут (кавычки внутри JSON иначе сломали бы атрибут).
func apiWorkstationEditFormHTMLHandler(w http.ResponseWriter, r *http.Request) {
	if !isIslandAuthed(r) {
		http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	id, err := idFromPath(r.URL.Path, "/api/reference/workstations/edit-form/")
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	rows, err := QueryDB(workstationListSelect+" WHERE w.id = ?", strconv.Itoa(id))
	if err != nil || len(rows) == 0 {
		http.Error(w, "АРМ не найден", http.StatusNotFound)
		return
	}
	row := rows[0]

	// Нормализация: SQLite возвращает NULL как "<nil>", boolean — как "true"/"false".
	// Приводим к контракту острова (все поля — строки, NULL → "", bool → "1"/"0").
	clean := func(v string) string {
		if v == "<nil>" {
			return ""
		}
		return v
	}
	toFlag := func(v string) string {
		v = clean(v)
		if v == "true" || v == "1" {
			return "1"
		}
		return "0"
	}
	payload := map[string]string{
		"id":                 clean(row["id"]),
		"inventory_number":   clean(row["inventory_number"]),
		"serial_number":      clean(row["serial_number"]),
		"seal_numbers":       clean(row["seal_numbers"]),
		"monitor_count":      clean(row["monitor_count"]),
		"employee_id":        clean(row["employee_id"]), // "" если не назначен
		"address_id":         clean(row["address_id"]),
		"is_vacant":          toFlag(row["is_vacant"]),
		"replacement_done":   toFlag(row["replacement_done"]),
		"replacement_date":   clean(row["replacement_date"]),
		"replacement_letter": clean(row["replacement_letter"]),
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Ошибка сериализации данных АРМ", http.StatusInternalServerError)
		return
	}

	// Экранирование для HTML-атрибута: & < > " '
	escapedData := html.EscapeString(string(jsonBytes))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div data-island="workstation-edit" data-initial-data='%s'></div>`, escapedData)
}

// ==================== РЕГИСТРАЦИЯ МАРШРУТОВ ========================================

// registerWorkstationIslandRoutes регистрирует island-маршруты раздела «АРМ».
// Вызывается из setupRoutes после authMiddleware. Go-мультиплексор выбирает более
// специфичный паттерн ("/api/reference/workstations/island" и ".../island/" раньше
// префикса ".../"), поэтому старый JSON API остаётся нетронутым.
func registerWorkstationIslandRoutes(authMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// Создание АРМ (Vue-остров)
	http.HandleFunc("/api/reference/workstations/island", recoverMiddleware(authMiddleware(apiWorkstationIslandCreateHandler)))
	// Обновление/удаление АРМ по ID (Vue-остров)
	http.HandleFunc("/api/reference/workstations/island/", recoverMiddleware(authMiddleware(apiWorkstationIslandItemHandler)))
	// HTML-фрагмент формы редактирования (Vue-остров)
	http.HandleFunc("/api/reference/workstations/edit-form/", recoverMiddleware(authMiddleware(apiWorkstationEditFormHTMLHandler)))
}
