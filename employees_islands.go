package main

// Порт-employees (TASKS.md раздел 3): Vue-острова + SSR-таблица для раздела
// «Сотрудники» по образцу раздела «Адреса» (addresses_islands.go — эталон).
// Island-эндпоинты:
//   POST   /api/reference/employees/island          — создание сотрудника
//   PUT    /api/reference/employees/island/{id}     — обновление сотрудника
//   DELETE /api/reference/employees/island/{id}     — удаление (с проверкой ссылок из АРМ/хостов)
//   GET    /api/reference/employees/edit-form/{id}  — HTML-фрагмент модалки редактирования
//
// Единый контракт данных (все поля — строки, включая address_id):
// full_name, short_name, phone_city, phone_internal, address_id.
// Правило ключа (TASKS.md раздел 1): уникален только id — ФИО и телефоны
// могут дублироваться; предпроверка дублей отсутствует.

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// ==================== КОНТРАКТ ДАННЫХ ====================

// IslandEmployeeRequest — единый JSON-контракт острова сотрудников.
// ВАЖНО: все поля — строки (включая address_id), чтобы исключить ошибки
// unmarshal string→int на границе клиент/сервер.
type IslandEmployeeRequest struct {
	FullName      string `json:"full_name"`
	ShortName     string `json:"short_name"`
	PhoneCity     string `json:"phone_city"`
	PhoneInternal string `json:"phone_internal"`
	AddressID     string `json:"address_id"`
}

// employeeColumns — ЕДИНЫЙ список колонок контракта сотрудника.
const employeeColumns = "id, full_name, short_name, phone_city, phone_internal, address_id, created_at"

// validateIslandEmployeeRequest — общая серверная валидация для создания и
// обновления: full_name и address_id обязательны, адрес должен существовать.
func validateIslandEmployeeRequest(req *IslandEmployeeRequest) error {
	req.FullName = strings.TrimSpace(req.FullName)
	req.ShortName = strings.TrimSpace(req.ShortName)
	req.PhoneCity = strings.TrimSpace(req.PhoneCity)
	req.PhoneInternal = strings.TrimSpace(req.PhoneInternal)
	req.AddressID = strings.TrimSpace(req.AddressID)

	if req.FullName == "" {
		return fmt.Errorf("ФИО обязательно для заполнения")
	}
	if req.AddressID == "" {
		return fmt.Errorf("не выбран адрес")
	}
	id, err := strconv.Atoi(req.AddressID)
	if err != nil || !validateID(id) {
		return fmt.Errorf("некорректный ID адреса")
	}
	// Проверка существования адреса (FK-целостность на уровне приложения)
	if _, err := querySingleInt("SELECT id FROM reference_addresses WHERE id = ?", id); err != nil {
		return fmt.Errorf("выбранный адрес не найден")
	}
	return nil
}

// ==================== SQL =================================

// buildEmployeeSQL — ОДИН запрос по полному контракту колонок (образец:
// buildAddressSQL). Пустые значения передаются плейсхолдерами как пустые
// строки — bindArgs сам подставляет NULL (СТОЯЩЕЕ ПРАВИЛО: строка "NULL"
// аргументом не передаётся). operation: "insert" | "update".
func buildEmployeeSQL(operation string, req *IslandEmployeeRequest) []interface{} {
	addrArg := req.AddressID // '' → NULL через bindArgs
	switch operation {
	case "insert":
		return []interface{}{req.FullName, req.ShortName, req.PhoneCity, req.PhoneInternal, addrArg}
	default: // update
		return []interface{}{req.FullName, req.ShortName, req.PhoneCity, req.PhoneInternal, addrArg}
	}
}

// execEmployeeInsert выполняет INSERT по контракту (правило ключа: дубли
// бизнес-полей разрешены, уникален только id).
func execEmployeeInsert(req *IslandEmployeeRequest) error {
	sqlStr := "INSERT INTO reference_employees (full_name, short_name, phone_city, phone_internal, address_id) VALUES (?, ?, ?, ?, ?)"
	return execSafe(sqlStr, buildEmployeeSQL("insert", req)...)
}

// execEmployeeUpdate выполняет UPDATE по контракту.
func execEmployeeUpdate(req *IslandEmployeeRequest, id int) error {
	sqlStr := "UPDATE reference_employees SET full_name = ?, short_name = ?, phone_city = ?, phone_internal = ?, address_id = ? WHERE id = ?"
	all := append(buildEmployeeSQL("update", req), id)
	return execSafe(sqlStr, all...)
}

// checkEmployeeUsage: используется ли сотрудник в АРМ и хостах.
// Возвращает текст ошибки или "".
func checkEmployeeUsage(id int) string {
	type usage struct {
		table string
		label string
	}
	uses := []usage{
		{table: "reference_workstations", label: "АРМ"},
		{table: "reference_hosts", label: "хостах"},
	}
	var found []string
	for _, u := range uses {
		n, err := querySingleInt("SELECT id FROM "+u.table+" WHERE employee_id = ? LIMIT 1", id)
		if err != nil && err.Error() != "sql: no rows in result set" {
			log.Printf("Ошибка проверки использования сотрудника %d в %s: %v", id, u.table, err)
		}
		if n > 0 {
			found = append(found, u.label)
		}
	}
	if len(found) > 0 {
		return "Сотрудник используется в " + strings.Join(found, ", ")
	}
	return ""
}

// decodeIslandEmployeeRequest декодирует и валидирует JSON-тело запроса.
func decodeIslandEmployeeRequest(w http.ResponseWriter, r *http.Request) (IslandEmployeeRequest, bool) {
	var req IslandEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, "Некорректный формат JSON: "+err.Error())
		return req, false
	}
	if err := validateIslandEmployeeRequest(&req); err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, err.Error())
		return req, false
	}
	return req, true
}

// ==================== ENDPOINT: POST /api/reference/employees/island ======

func apiEmployeeIslandCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeIslandJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}
	if !isIslandAuthed(r) {
		writeIslandJSONError(w, http.StatusUnauthorized, "Требуется аутентификация")
		return
	}

	req, ok := decodeIslandEmployeeRequest(w, r)
	if !ok {
		return
	}

	if err := execEmployeeInsert(&req); err != nil {
		log.Printf("Ошибка создания сотрудника (island): %v", err)
		writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при создании сотрудника")
		return
	}

	logPanel("Добавлен сотрудник (Vue-остров): " + req.FullName)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// ==================== ENDPOINT: /api/reference/employees/island/{id} ======

func apiEmployeeIslandItemHandler(w http.ResponseWriter, r *http.Request) {
	if !isIslandAuthed(r) {
		writeIslandJSONError(w, http.StatusUnauthorized, "Требуется аутентификация")
		return
	}

	id, err := idFromPath(r.URL.Path, "/api/reference/employees/island/")
	if err != nil {
		writeIslandJSONError(w, http.StatusBadRequest, "Некорректный ID сотрудника")
		return
	}

	switch r.Method {
	case http.MethodPut:
		req, ok := decodeIslandEmployeeRequest(w, r)
		if !ok {
			return
		}
		// Существование записи (404 до попытки UPDATE)
		if _, err := querySingleInt("SELECT id FROM reference_employees WHERE id = ?", id); err != nil {
			writeIslandJSONError(w, http.StatusNotFound, "Запись не найдена")
			return
		}
		if err := execEmployeeUpdate(&req, id); err != nil {
			log.Printf("Ошибка обновления сотрудника %d (island): %v", id, err)
			writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при обновлении сотрудника")
			return
		}
		logPanel("Обновлён сотрудник с ID " + strconv.Itoa(id) + " (Vue-остров)")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case http.MethodDelete:
		if _, err := querySingleInt("SELECT id FROM reference_employees WHERE id = ?", id); err != nil {
			writeIslandJSONError(w, http.StatusNotFound, "Запись не найдена")
			return
		}
		if usageErr := checkEmployeeUsage(id); usageErr != "" {
			writeIslandJSONError(w, http.StatusBadRequest, usageErr)
			return
		}
		if err := execSafe("DELETE FROM reference_employees WHERE id = ?", id); err != nil {
			log.Printf("Ошибка удаления сотрудника %d (island): %v", id, err)
			writeIslandJSONError(w, http.StatusInternalServerError, "Ошибка базы данных при удалении сотрудника")
			return
		}
		logPanel("Удалён сотрудник с ID " + strconv.Itoa(id) + " (Vue-остров)")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		writeIslandJSONError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	}
}

// ==================== ENDPOINT: GET /api/reference/employees/edit-form/{id}

// Возвращает HTML-фрагмент для Vue-острова EmployeeEditModal: div с
// data-island="employee-edit" и ЭКРАНИРОВАННЫМ data-initial-data (по образцу
// адресов).
func apiEmployeeEditFormHTMLHandler(w http.ResponseWriter, r *http.Request) {
	if !isIslandAuthed(r) {
		http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
		return
	}

	id, err := idFromPath(r.URL.Path, "/api/reference/employees/edit-form/")
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	rows, err := QueryDB("SELECT "+employeeColumns+" FROM reference_employees WHERE id = ?", strconv.Itoa(id))
	if err != nil || len(rows) == 0 {
		http.Error(w, "Сотрудник не найден", http.StatusNotFound)
		return
	}
	row := rows[0]

	payload := map[string]string{
		"id":             row["id"],
		"full_name":      row["full_name"],
		"short_name":     row["short_name"],
		"phone_city":     row["phone_city"],
		"phone_internal": row["phone_internal"],
		"address_id":     row["address_id"],
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Ошибка сериализации данных сотрудника", http.StatusInternalServerError)
		return
	}

	escapedData := html.EscapeString(string(jsonBytes))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div data-island="employee-edit" data-initial-data='%s'></div>`, escapedData)
}

// ==================== РЕГИСТРАЦИЯ МАРШРУТОВ =================================

// registerEmployeeIslandRoutes — island-маршруты раздела «Сотрудники».
// Вызывается из setupRoutes после registerAddressIslandRoutes. Старый JSON API
// (/api/reference/employees) остаётся зарегистрированным до полного демонтажа
// Alpine (Фазы C-D, TASKS.md раздел 5).
func registerEmployeeIslandRoutes(authMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	http.HandleFunc("/api/reference/employees/island", recoverMiddleware(authMiddleware(apiEmployeeIslandCreateHandler)))
	http.HandleFunc("/api/reference/employees/island/", recoverMiddleware(authMiddleware(apiEmployeeIslandItemHandler)))
	http.HandleFunc("/api/reference/employees/edit-form/", recoverMiddleware(authMiddleware(apiEmployeeEditFormHTMLHandler)))
}
