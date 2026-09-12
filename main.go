package main

import (
	
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"andy-key/frontend"
)

const DEFAULT_PORT = 9000 // AK-v.2.1.1: Fixed static routing and MIME types

var (
	taskMutex    sync.Mutex
	runningTasks = make(map[int]chan struct{})
	footerText   string // Глобальная переменная для хранения расшифрованного футера
)

// ==================== ИНИЦИАЛИЗАЦИЯ ШАБЛОНОВ ====================

var (
	templates *template.Template
)

func initTemplates() error {
	// Создаём парсер шаблонов с функциями dict и default
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if key, ok := values[i].(string); ok {
					dict[key] = values[i+1]
				}
			}
			return dict
		},
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" {
				return def
			}
			return val
		},
	}
	templates = template.New("").Funcs(funcMap)
	
	// Проходим по всем HTML файлам в embedded FS (основные шаблоны)
	pattern := "templates/*.html"
	matches, err := fs.Glob(frontend.TemplatesFS, pattern)
	if err != nil {
		return fmt.Errorf("ошибка поиска шаблонов: %w", err)
	}
	
	if len(matches) == 0 {
		return fmt.Errorf("шаблоны не найдены")
	}
	
	// Парсим каждый основной шаблон
	for _, match := range matches {
		content, err := frontend.TemplatesFS.ReadFile(match)
		if err != nil {
			return fmt.Errorf("ошибка чтения шаблона %s: %w", match, err)
		}
		
		name := strings.TrimSuffix(strings.TrimPrefix(match, "templates/"), ".html")
		tmpl, err := templates.New(name).Parse(string(content))
		if err != nil {
			return fmt.Errorf("ошибка парсинга шаблона %s: %w", match, err)
		}
		templates = tmpl
	}
	
	// Дополнительно загружаем компоненты из подпапки components
	componentPattern := "templates/components/*.html"
	componentMatches, err := fs.Glob(frontend.TemplatesFS, componentPattern)
	if err != nil {
		return fmt.Errorf("ошибка поиска компонентов: %w", err)
	}
	
	for _, match := range componentMatches {
		content, err := frontend.TemplatesFS.ReadFile(match)
		if err != nil {
			return fmt.Errorf("ошибка чтения компонента %s: %w", match, err)
		}
		
		_, err = templates.Parse(string(content))
		if err != nil {
			return fmt.Errorf("ошибка парсинга компонента %s: %w", match, err)
		}
	}
	
	return nil
}

// ==================== ОБРАБОТЧИКИ ДЛЯ АВТОРИЗАЦИИ ====================

func indexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	// Рендерим главный шаблон index
	data := map[string]string{
		"FooterText": footerText,
		"AppVersion": CURRENT_SCHEMA_VERSION,
	}
	
	if err := templates.ExecuteTemplate(w, "index", data); err != nil {
		http.Error(w, "Ошибка рендеринга шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// contentSectionHandler обрабатывает запросы контента для различных разделов
func contentSectionHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем имя раздела из пути /api/content/{section}
	section := strings.TrimPrefix(r.URL.Path, "/api/content/")
	if section == "" {
		section = "tasks" // по умолчанию
	}

	// Маппинг имен разделов на шаблоны
	templateMap := map[string]string{
		"tasks":             "tasks",
		"credentials":       "credentials",
		"scripts":           "scripts",
		"bash-constructor":  "bash_constructor",
		"logs":              "logs",
		"addresses":         "addresses",
		"employees":         "employees",
		"workstations":      "workstations",
		"hosts":             "hosts",
		"network-equipment": "network_equipment",
		"network-mfps":      "network_mfps",
		"ip-phones":         "ip_phones",
	}

	templateName, exists := templateMap[section]
	if !exists {
		http.Error(w, "Раздел не найден: "+section, http.StatusNotFound)
		return
	}

	// Проверяем, это HTMX-запрос или обычный
	isHtmx := r.Header.Get("HX-Request") == "true"

	// Загружаем данные из БД в зависимости от раздела
	data := make(map[string]interface{})
	var dbError error
	
	switch section {
	case "addresses":
		rows, err := QueryDB(`SELECT id, street, building, cabinet, corridor, floor, service_room, created_at FROM reference_addresses ORDER BY street, building`)
		if err != nil {
			dbError = err
		} else {
			data["addresses"] = rows
		}
	case "employees":
		rows, err := QueryDB(`SELECT id, full_name, short_name, phone_city, phone_internal, address_id, created_at FROM reference_employees ORDER BY full_name`)
		if err != nil {
			dbError = err
		} else {
			data["employees"] = rows
		}
	case "workstations":
		rows, err := QueryDB(`SELECT w.id, w.inventory_number, w.serial_number, w.seal_numbers, w.monitor_count, w.employee_id, w.replacement_done, w.replacement_date, w.replacement_letter, w.address_id, 
		                     e.full_name as employee_name, a.street, a.building 
		                     FROM reference_workstations w
		                     LEFT JOIN reference_employees e ON w.employee_id = e.id
		                     LEFT JOIN reference_addresses a ON w.address_id = a.id
		                     ORDER BY a.street, a.building, w.inventory_number`)
		if err != nil {
			dbError = err
		} else {
			data["workstations"] = rows
		}
	case "hosts":
		rows, err := QueryDB(`SELECT h.id, h.ip, h.ssh_port, h.address_id, h.employee_id,
		                     e.full_name as employee_name, a.street, a.building
		                     FROM reference_hosts h
		                     LEFT JOIN reference_employees e ON h.employee_id = e.id
		                     LEFT JOIN reference_addresses a ON h.address_id = a.id
		                     ORDER BY h.ip`)
		if err != nil {
			dbError = err
		} else {
			data["hosts"] = rows
		}
	case "network-equipment":
		rows, err := QueryDB(`SELECT id, name, ip, mac, model, serial_number, location, firmware, config_backup, last_check, status, address_id, created_at FROM reference_network_equipment ORDER BY name`)
		if err != nil {
			dbError = err
		} else {
			data["equipment"] = rows
		}
	case "network-mfps":
		rows, err := QueryDB(`SELECT m.id, m.name, m.ip, m.mac, m.model, m.serial_number, m.location, m.firmware, m.config_backup, m.last_check, m.status, m.address_id,
		                     a.street, a.building
		                     FROM reference_network_mfps m
		                     LEFT JOIN reference_addresses a ON m.address_id = a.id
		                     ORDER BY m.name`)
		if err != nil {
			dbError = err
		} else {
			data["mfps"] = rows
		}
	case "ip-phones":
		rows, err := QueryDB(`SELECT p.id, p.name, p.ip, p.mac, p.model, p.serial_number, p.extension, p.sip_server, p.status, p.address_id,
		                     a.street, a.building
		                     FROM reference_ip_phones p
		                     LEFT JOIN reference_addresses a ON p.address_id = a.id
		                     ORDER BY p.name`)
		if err != nil {
			dbError = err
		} else {
			data["phones"] = rows
		}
	}

	if isHtmx {
		// Для HTMX рендерим только фрагмент контента (без layout)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
		// Добавляем информацию об ошибке БД в данные
		if dbError != nil {
			data["dbError"] = dbError.Error()
		}
		
		// Пробуем рендерить шаблон, если его нет - выводим заглушку
		err := templates.ExecuteTemplate(w, templateName, data)
		if err != nil {
			// Шаблон не найден, выводим временную заглушку
			fmt.Fprintf(w, `<div class="placeholder-content">
				<h2>Раздел "%s"</h2>
				<p class="text-gray-500">Контент в разработке. Шаблон '%s.html' отсутствует.</p>
			</div>`, section, templateName)
			return
		}
	} else {
		// Для обычного запроса рендерим полную страницу с layout
		pageData := map[string]interface{}{
			"Section":    section,
			"FooterText": footerText,
			"AppVersion": CURRENT_SCHEMA_VERSION,
		}
		// Добавляем данные из БД
		for k, v := range data {
			pageData[k] = v
		}
		if dbError != nil {
			pageData["dbError"] = dbError.Error()
		}
		if err := templates.ExecuteTemplate(w, "index", pageData); err != nil {
			http.Error(w, "Ошибка рендеринга: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]string{
			"FooterText": footerText,
			"AppVersion": CURRENT_SCHEMA_VERSION,
		}
		if err := templates.ExecuteTemplate(w, "login", data); err != nil {
			http.Error(w, "Ошибка рендеринга шаблона: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON"})
		return
	}

	ok, userID, errorMsg := CheckAdminAuth(req.Username, req.Password)
	if !ok {
		logPanel("Неудачная попытка входа: логин='" + req.Username + "', ошибка='" + errorMsg + "'")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
		return
	}

	logPanel("Успешный вход: пользователь='" + req.Username + "', user_id=" + strconv.Itoa(userID))

	token, err := CreateSession(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка создания сессии"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	// Возвращаем URL для редиректа вместо простого success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"redirect": "/"})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		ExecDB("DELETE FROM sessions WHERE token = '" + cookie.Value + "'")
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-1 * time.Hour),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil || !CheckSession(cookie.Value) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
			return
		}
		next(w, r)
	}
}

// ==================== ОБРАБОТЧИК ДЛЯ ИНФОРМАЦИИ О ПОЛЬЗОВАТЕЛЕ ====================

func apiUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	// Получаем информацию о пользователе из БД по токену сессии
	escapedToken := strings.ReplaceAll(cookie.Value, "'", "''")
	rows, err := QueryDB(fmt.Sprintf(`
		SELECT u.username
		FROM admin_users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.token = '%s' AND s.expires_at > DATETIME('now')
	`, escapedToken))

	if err != nil || len(rows) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Сессия не найдена"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"username": rows[0]["username"],
	})
}

// ==================== ОБРАБОТЧИКИ ДЛЯ ЖУРНАЛА ====================

func apiLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	logType := r.URL.Query().Get("type")
	limit := 100

	switch logType {
	case "panel", "remote", "diagnostic":
		entries, err := getLogEntries(logType, limit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка чтения журнала: " + err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type":    logType,
			"entries": entries,
			"count":   len(entries),
		})
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный тип журнала"})
	}
}

// ==================== ОБРАБОТЧИКИ ДЛЯ СПРАВОЧНИКОВ ====================

// Эндпоинт для работы с адресами
func apiReferenceAddressesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB(`
			SELECT id, street, building, cabinet, corridor, floor, service_room, created_at
			FROM reference_addresses
			ORDER BY street, building, floor, cabinet, corridor, service_room
		`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		
		// Форматируем полные адреса для отображения
		for i := range rows {
			street := rows[i]["street"]
			building := rows[i]["building"]
			cabinet := rows[i]["cabinet"]
			corridor := rows[i]["corridor"]
			serviceRoom := rows[i]["service_room"]
			floorStr := rows[i]["floor"]
			
			rows[i]["full_address"] = FormatFullAddress(street, building, cabinet, corridor, serviceRoom, "", floorStr)
		}
		
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req AddressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация
		if err := req.Validate(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Экранирование значений
		escapedStreet := strings.ReplaceAll(req.Street, "'", "''")
		escapedBuilding := strings.ReplaceAll(req.Building, "'", "''")
		escapedCabinet := strings.ReplaceAll(req.Cabinet, "'", "''")
		escapedCorridor := strings.ReplaceAll(req.Corridor, "'", "''")
		escapedServiceRoom := strings.ReplaceAll(req.ServiceRoom, "'", "''")
		escapedGarage := strings.ReplaceAll(req.Garage, "'", "''")
		floorStr := "NULL"
		if req.Floor != "" {
			floorStr = fmt.Sprintf("'%s'", strings.ReplaceAll(req.Floor, "'", "''"))
		}

		query := fmt.Sprintf(`
			INSERT INTO reference_addresses (street, building, cabinet, corridor, floor, service_room, garage)
			VALUES ('%s', '%s', '%s', '%s', %s, '%s', '%s')
		`, escapedStreet, escapedBuilding, escapedCabinet, escapedCorridor, floorStr, escapedServiceRoom, escapedGarage)

		if err := ExecDB(query); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Такой адрес уже существует"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлен адрес: " + FormatFullAddress(req.Street, req.Building, req.Cabinet, req.Corridor, req.ServiceRoom, req.Garage, req.Floor))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/addresses/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID адреса"})
			return
		}

		var req AddressRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация
		if err := req.Validate(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Экранирование значений
		escapedStreet := strings.ReplaceAll(req.Street, "'", "''")
		escapedBuilding := strings.ReplaceAll(req.Building, "'", "''")
		escapedCabinet := strings.ReplaceAll(req.Cabinet, "'", "''")
		escapedCorridor := strings.ReplaceAll(req.Corridor, "'", "''")
		escapedServiceRoom := strings.ReplaceAll(req.ServiceRoom, "'", "''")
		floorStr := "NULL"
		if req.Floor != "" {
			floorStr = fmt.Sprintf("'%s'", strings.ReplaceAll(req.Floor, "'", "''"))
		}

		query := fmt.Sprintf(`
			UPDATE reference_addresses
			SET street = '%s', building = '%s', cabinet = '%s', corridor = '%s', floor = %s, service_room = '%s'
			WHERE id = %d
		`, escapedStreet, escapedBuilding, escapedCabinet, escapedCorridor, floorStr, escapedServiceRoom, id)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлён адрес с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/addresses/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID адреса"})
			return
		}

		// Проверяем, используется ли адрес в других таблицах
		usedInEmployees, _ := QueryDB(fmt.Sprintf("SELECT id FROM reference_employees WHERE address_id = %d LIMIT 1", id))
		usedInWorkstations, _ := QueryDB(fmt.Sprintf("SELECT id FROM reference_workstations WHERE address_id = %d LIMIT 1", id))
		usedInHosts, _ := QueryDB(fmt.Sprintf("SELECT id FROM reference_hosts WHERE address_id = %d LIMIT 1", id))
		
		if len(usedInEmployees) > 0 || len(usedInWorkstations) > 0 || len(usedInHosts) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Нельзя удалить адрес, так как он используется в других справочниках"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM reference_addresses WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удалён адрес с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// Эндпоинт для работы с сотрудниками
func apiReferenceEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB(`
			SELECT e.id, e.full_name, e.short_name, e.phone_city, e.phone_internal, e.address_id,
			       a.street, a.building, a.cabinet, a.corridor, a.floor, a.service_room
			FROM reference_employees e
			JOIN reference_addresses a ON e.address_id = a.id
			ORDER BY e.full_name
		`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		
		// Форматируем полные адреса для отображения
		for i := range rows {
			street := rows[i]["street"]
			building := rows[i]["building"]
			cabinet := rows[i]["cabinet"]
			corridor := rows[i]["corridor"]
			serviceRoom := rows[i]["service_room"]
			floorStr := rows[i]["floor"]
			
			rows[i]["full_address"] = FormatFullAddress(street, building, cabinet, corridor, serviceRoom, "", floorStr)
		}
		
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req struct {
			FullName      string `json:"full_name"`
			ShortName     string `json:"short_name"`
			PhoneCity     string `json:"phone_city"`
			PhoneInternal string `json:"phone_internal"`
			AddressID     int    `json:"address_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.FullName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ФИО полностью обязательно для заполнения"})
			return
		}
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}

		// Проверяем, что адрес не является служебным помещением
		addressCheck, err := QueryDB(fmt.Sprintf(`
			SELECT service_room FROM reference_addresses WHERE id = %d
		`, req.AddressID))
		if err != nil || len(addressCheck) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Указанный адрес не найден"})
			return
		}
		
		if addressCheck[0]["service_room"] != "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Сотрудники в служебных помещениях не учитываются"})
			return
		}

		// Экранирование значений
		escapedFullName := strings.ReplaceAll(req.FullName, "'", "''")
		escapedShortName := strings.ReplaceAll(req.ShortName, "'", "''")
		escapedPhoneCity := strings.ReplaceAll(req.PhoneCity, "'", "''")
		escapedPhoneInternal := strings.ReplaceAll(req.PhoneInternal, "'", "''")

		query := fmt.Sprintf(`
			INSERT INTO reference_employees (full_name, short_name, phone_city, phone_internal, address_id)
			VALUES ('%s', '%s', '%s', '%s', %d)
		`, escapedFullName, escapedShortName, escapedPhoneCity, escapedPhoneInternal, req.AddressID)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлен сотрудник: " + req.FullName)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/employees/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID сотрудника"})
			return
		}

		var req struct {
			FullName      string `json:"full_name"`
			ShortName     string `json:"short_name"`
			PhoneCity     string `json:"phone_city"`
			PhoneInternal string `json:"phone_internal"`
			AddressID     int    `json:"address_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.FullName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ФИО полностью обязательно для заполнения"})
			return
		}
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}

		// Проверяем, что адрес не является служебным помещением
		addressCheck, err := QueryDB(fmt.Sprintf(`
			SELECT service_room FROM reference_addresses WHERE id = %d
		`, req.AddressID))
		if err != nil || len(addressCheck) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Указанный адрес не найден"})
			return
		}
		
		if addressCheck[0]["service_room"] != "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Сотрудники в служебных помещениях не учитываются"})
			return
		}

		// Экранирование значений
		escapedFullName := strings.ReplaceAll(req.FullName, "'", "''")
		escapedShortName := strings.ReplaceAll(req.ShortName, "'", "''")
		escapedPhoneCity := strings.ReplaceAll(req.PhoneCity, "'", "''")
		escapedPhoneInternal := strings.ReplaceAll(req.PhoneInternal, "'", "''")

		query := fmt.Sprintf(`
			UPDATE reference_employees
			SET full_name = '%s', short_name = '%s', phone_city = '%s', phone_internal = '%s', address_id = %d
			WHERE id = %d
		`, escapedFullName, escapedShortName, escapedPhoneCity, escapedPhoneInternal, req.AddressID, id)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлён сотрудник с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/employees/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID сотрудника"})
			return
		}

		// Проверяем, используется ли сотрудник в других таблицах
		usedInWorkstations, _ := QueryDB(fmt.Sprintf("SELECT id FROM reference_workstations WHERE employee_id = %d LIMIT 1", id))
		usedInHosts, _ := QueryDB(fmt.Sprintf("SELECT id FROM reference_hosts WHERE employee_id = %d LIMIT 1", id))
		
		if len(usedInWorkstations) > 0 || len(usedInHosts) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Нельзя удалить сотрудника, так как он привязан к АРМ или хостам"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM reference_employees WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удалён сотрудник с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// Эндпоинт для работы с АРМ
func apiReferenceWorkstationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB(`
			SELECT w.id, w.address_id, w.is_vacant, w.employee_id, w.inventory_number, 
			       w.seal_numbers, w.monitor_count, w.serial_number, w.replacement_done,
			       w.replacement_date, w.replacement_letter,
			       a.street, a.building, a.cabinet, a.corridor, a.floor, a.service_room,
			       e.short_name as employee_short_name
			FROM reference_workstations w
			JOIN reference_addresses a ON w.address_id = a.id
			LEFT JOIN reference_employees e ON w.employee_id = e.id
			ORDER BY a.street, a.building, a.floor, a.cabinet, a.corridor, a.service_room
		`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		
		// Форматируем полные адреса для отображения
		for i := range rows {
			street := rows[i]["street"]
			building := rows[i]["building"]
			cabinet := rows[i]["cabinet"]
			corridor := rows[i]["corridor"]
			serviceRoom := rows[i]["service_room"]
			floorStr := rows[i]["floor"]
			
			rows[i]["full_address"] = FormatFullAddress(street, building, cabinet, corridor, serviceRoom, "", floorStr)
		}
		
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req WorkstationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация
		if err := req.Validate(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Экранирование значений
		escapedInventoryNumber := strings.ReplaceAll(req.InventoryNumber, "'", "''")
		escapedSealNumbers := strings.ReplaceAll(req.SealNumbers, "'", "''")
		escapedSerialNumber := strings.ReplaceAll(req.SerialNumber, "'", "''")
		escapedReplacementDate := "NULL"
		if req.ReplacementDate != "" {
			escapedReplacementDate = "'" + strings.ReplaceAll(req.ReplacementDate, "'", "''") + "'"
		}
		escapedReplacementLetter := "NULL"
		if req.ReplacementLetter != "" {
			escapedReplacementLetter = "'" + strings.ReplaceAll(req.ReplacementLetter, "'", "''") + "'"
		}
		employeeIDStr := "NULL"
		if req.EmployeeID != 0 {
			employeeIDStr = strconv.Itoa(req.EmployeeID)
		}

		query := fmt.Sprintf(`
			INSERT INTO reference_workstations (
				address_id, is_vacant, employee_id, inventory_number, seal_numbers,
				monitor_count, serial_number, replacement_done, replacement_date, replacement_letter
			) VALUES (
				%d, %v, %s, '%s', '%s', %d, '%s', %v, %s, %s
			)
		`, req.AddressID, req.IsVacant, employeeIDStr, escapedInventoryNumber, escapedSealNumbers,
		   req.MonitorCount, escapedSerialNumber, req.ReplacementDone, escapedReplacementDate, escapedReplacementLetter)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлено АРМ: инв. номер " + req.InventoryNumber)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/workstations/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID АРМ"})
			return
		}

		var req WorkstationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация
		if err := req.Validate(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Экранирование значений
		escapedInventoryNumber := strings.ReplaceAll(req.InventoryNumber, "'", "''")
		escapedSealNumbers := strings.ReplaceAll(req.SealNumbers, "'", "''")
		escapedSerialNumber := strings.ReplaceAll(req.SerialNumber, "'", "''")
		escapedReplacementDate := "NULL"
		if req.ReplacementDate != "" {
			escapedReplacementDate = "'" + strings.ReplaceAll(req.ReplacementDate, "'", "''") + "'"
		}
		escapedReplacementLetter := "NULL"
		if req.ReplacementLetter != "" {
			escapedReplacementLetter = "'" + strings.ReplaceAll(req.ReplacementLetter, "'", "''") + "'"
		}
		employeeIDStr := "NULL"
		if req.EmployeeID != 0 {
			employeeIDStr = strconv.Itoa(req.EmployeeID)
		}

		query := fmt.Sprintf(`
			UPDATE reference_workstations
			SET address_id = %d, is_vacant = %v, employee_id = %s, inventory_number = '%s',
			    seal_numbers = '%s', monitor_count = %d, serial_number = '%s',
			    replacement_done = %v, replacement_date = %s, replacement_letter = %s
			WHERE id = %d
		`, req.AddressID, req.IsVacant, employeeIDStr, escapedInventoryNumber, escapedSealNumbers,
		   req.MonitorCount, escapedSerialNumber, req.ReplacementDone, escapedReplacementDate, 
		   escapedReplacementLetter, id)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлено АРМ с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/workstations/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID АРМ"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM reference_workstations WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удалено АРМ с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// Эндпоинт для работы с хостами
func apiReferenceHostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB(`
			SELECT h.id, h.address_id, h.employee_id, h.ip, h.ssh_port, h.enabled,
			       a.street, a.building, a.cabinet, a.corridor, a.floor, a.service_room,
			       e.short_name as employee_short_name
			FROM reference_hosts h
			JOIN reference_addresses a ON h.address_id = a.id
			LEFT JOIN reference_employees e ON h.employee_id = e.id
			ORDER BY a.street, a.building, a.floor, a.cabinet, a.corridor, a.service_room
		`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		
		// Форматируем полные адреса для отображения
		for i := range rows {
			street := rows[i]["street"]
			building := rows[i]["building"]
			cabinet := rows[i]["cabinet"]
			corridor := rows[i]["corridor"]
			serviceRoom := rows[i]["service_room"]
			floorStr := rows[i]["floor"]
			
			rows[i]["full_address"] = FormatFullAddress(street, building, cabinet, corridor, serviceRoom, "", floorStr)
		}
		
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req HostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}
		if req.IP == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "IP адрес обязателен для заполнения"})
			return
		}

		// Экранирование значений
		escapedIP := strings.ReplaceAll(req.IP, "'", "''")
		employeeIDStr := "NULL"
		if req.EmployeeID != 0 {
			employeeIDStr = strconv.Itoa(req.EmployeeID)
		}
		sshPort := req.SSHPort
		if sshPort == 0 {
			sshPort = 22
		}
		enabled := 1
		if !req.Enabled {
			enabled = 0
		}

		query := fmt.Sprintf(`
			INSERT INTO reference_hosts (address_id, employee_id, ip, ssh_port, enabled)
			VALUES (%d, %s, '%s', %d, %d)
		`, req.AddressID, employeeIDStr, escapedIP, sshPort, enabled)

		if err := ExecDB(query); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Хост с таким IP адресом уже существует"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлен хост: " + req.IP)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/hosts/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID хоста"})
			return
		}

		var req HostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}
		if req.IP == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "IP адрес обязателен для заполнения"})
			return
		}

		// Экранирование значений
		escapedIP := strings.ReplaceAll(req.IP, "'", "''")
		employeeIDStr := "NULL"
		if req.EmployeeID != 0 {
			employeeIDStr = strconv.Itoa(req.EmployeeID)
		}
		sshPort := req.SSHPort
		if sshPort == 0 {
			sshPort = 22
		}
		enabled := 1
		if !req.Enabled {
			enabled = 0
		}

		query := fmt.Sprintf(`
			UPDATE reference_hosts
			SET address_id = %d, employee_id = %s, ip = '%s', ssh_port = %d, enabled = %d
			WHERE id = %d
		`, req.AddressID, employeeIDStr, escapedIP, sshPort, enabled, id)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлён хост с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/hosts/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID хоста"})
			return
		}

		// Проверяем, используется ли хост в задачах
		usedInTasks, _ := QueryDB(fmt.Sprintf("SELECT id FROM task_hosts WHERE host_id = %d LIMIT 1", id))
		if len(usedInTasks) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Нельзя удалить хост, так как он используется в задачах"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM reference_hosts WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удалён хост с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// Эндпоинт для работы с сетевым оборудованием
func apiReferenceNetworkEquipmentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB(`
			SELECT ne.id, ne.address_id, ne.category, ne.model, ne.type, ne.port_count,
			       a.street, a.building, a.cabinet, a.corridor, a.floor, a.service_room
			FROM reference_network_equipment ne
			JOIN reference_addresses a ON ne.address_id = a.id
			ORDER BY a.street, a.building, a.floor, a.cabinet, a.corridor, a.service_room, ne.category, ne.model
		`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		
		// Форматируем полные адреса для отображения
		for i := range rows {
			street := rows[i]["street"]
			building := rows[i]["building"]
			cabinet := rows[i]["cabinet"]
			corridor := rows[i]["corridor"]
			serviceRoom := rows[i]["service_room"]
			floorStr := rows[i]["floor"]
			
			rows[i]["full_address"] = FormatFullAddress(street, building, cabinet, corridor, serviceRoom, "", floorStr)
		}
		
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req NetworkEquipmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}
		if req.Category == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Категория обязательна для заполнения"})
			return
		}
		if req.Type == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Тип обязательна для заполнения"})
			return
		}
		if req.Model == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Модель обязательна для заполнения"})
			return
		}
		if req.PortCount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Число портов должно быть больше 0"})
			return
		}

		// Экранирование значений
		escapedCategory := strings.ReplaceAll(req.Category, "'", "''")
		escapedType := strings.ReplaceAll(req.Type, "'", "''")
		escapedModel := strings.ReplaceAll(req.Model, "'", "''")

		query := fmt.Sprintf(`
			INSERT INTO reference_network_equipment (address_id, category, type, model, port_count)
			VALUES (%d, '%s', '%s', '%s', %d)
		`, req.AddressID, escapedCategory, escapedType, escapedModel, req.PortCount)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлено сетевое оборудование: " + req.Category + " " + req.Model)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/network-equipment/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID оборудования"})
			return
		}

		var req NetworkEquipmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}
		if req.Category == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Категория обязательна для заполнения"})
			return
		}
		if req.Type == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Тип обязательна для заполнения"})
			return
		}
		if req.Model == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Модель обязательна для заполнения"})
			return
		}
		if req.PortCount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Число портов должно быть больше 0"})
			return
		}

		// Экранирование значений
		escapedCategory := strings.ReplaceAll(req.Category, "'", "''")
		escapedType := strings.ReplaceAll(req.Type, "'", "''")
		escapedModel := strings.ReplaceAll(req.Model, "'", "''")

		query := fmt.Sprintf(`
			UPDATE reference_network_equipment
			SET address_id = %d, category = '%s', type = '%s', model = '%s', port_count = %d
			WHERE id = %d
		`, req.AddressID, escapedCategory, escapedType, escapedModel, req.PortCount, id)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлено сетевое оборудование с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/network-equipment/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID оборудования"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM reference_network_equipment WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удалено сетевое оборудование с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// Эндпоинт для работы с сетевыми МФУ
func apiReferenceNetworkMFPsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB(`
			SELECT m.id, m.address_id, m.model, m.ip, m.hostname, m.serial_number,
			       a.street, a.building, a.cabinet, a.corridor, a.floor, a.service_room
			FROM reference_network_mfps m
			JOIN reference_addresses a ON m.address_id = a.id
			ORDER BY a.street, a.building, a.floor, a.cabinet, a.corridor, a.service_room, m.model
		`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		
		// Форматируем полные адреса для отображения
		for i := range rows {
			street := rows[i]["street"]
			building := rows[i]["building"]
			cabinet := rows[i]["cabinet"]
			corridor := rows[i]["corridor"]
			serviceRoom := rows[i]["service_room"]
			floorStr := rows[i]["floor"]
			
			rows[i]["full_address"] = FormatFullAddress(street, building, cabinet, corridor, serviceRoom, "", floorStr)
		}
		
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req NetworkMFPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}
		if req.Model == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Модель обязательна для заполнения"})
			return
		}
		if req.IP == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "IP адрес обязателен для заполнения"})
			return
		}
		if req.SerialNumber == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Серийный номер обязателен для заполнения"})
			return
		}

		// Экранирование значений
		escapedModel := strings.ReplaceAll(req.Model, "'", "''")
		escapedIP := strings.ReplaceAll(req.IP, "'", "''")
		escapedHostname := "NULL"
		if req.Hostname != "" {
			escapedHostname = "'" + strings.ReplaceAll(req.Hostname, "'", "''") + "'"
		}
		escapedSerialNumber := strings.ReplaceAll(req.SerialNumber, "'", "''")

		query := fmt.Sprintf(`
			INSERT INTO reference_network_mfps (address_id, model, ip, hostname, serial_number)
			VALUES (%d, '%s', '%s', %s, '%s')
		`, req.AddressID, escapedModel, escapedIP, escapedHostname, escapedSerialNumber)

		if err := ExecDB(query); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "МФУ с таким IP адресом уже существует"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлено сетевое МФУ: " + req.Model)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/network-mfps/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID МФУ"})
			return
		}

		var req NetworkMFPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}
		if req.Model == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Модель обязательна для заполнения"})
			return
		}
		if req.IP == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "IP адрес обязателен для заполнения"})
			return
		}
		if req.SerialNumber == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Серийный номер обязателен для заполнения"})
			return
		}

		// Экранирование значений
		escapedModel := strings.ReplaceAll(req.Model, "'", "''")
		escapedIP := strings.ReplaceAll(req.IP, "'", "''")
		escapedHostname := "NULL"
		if req.Hostname != "" {
			escapedHostname = "'" + strings.ReplaceAll(req.Hostname, "'", "''") + "'"
		}
		escapedSerialNumber := strings.ReplaceAll(req.SerialNumber, "'", "''")

		query := fmt.Sprintf(`
			UPDATE reference_network_mfps
			SET address_id = %d, model = '%s', ip = '%s', hostname = %s, serial_number = '%s'
			WHERE id = %d
		`, req.AddressID, escapedModel, escapedIP, escapedHostname, escapedSerialNumber, id)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлено сетевое МФУ с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/network-mfps/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID МФУ"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM reference_network_mfps WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удалено сетевое МФУ с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// Эндпоинт для работы с IP телефонами
func apiReferenceIPPhonesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB(`
			SELECT p.id, p.address_id, p.employee_full_name, p.ip,
			       a.street, a.building, a.cabinet, a.corridor, a.floor, a.service_room
			FROM reference_ip_phones p
			JOIN reference_addresses a ON p.address_id = a.id
			ORDER BY a.street, a.building, a.floor, a.cabinet, a.corridor, a.service_room, p.employee_full_name
		`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		
		// Форматируем полные адреса для отображения
		for i := range rows {
			street := rows[i]["street"]
			building := rows[i]["building"]
			cabinet := rows[i]["cabinet"]
			corridor := rows[i]["corridor"]
			serviceRoom := rows[i]["service_room"]
			floorStr := rows[i]["floor"]
			
			rows[i]["full_address"] = FormatFullAddress(street, building, cabinet, corridor, serviceRoom, "", floorStr)
		}
		
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req IPPhoneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}
		if req.EmployeeFullName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ФИО сотрудника обязательно для заполнения"})
			return
		}
		if req.IP == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "IP адрес обязателен для заполнения"})
			return
		}

		// Экранирование значений
		escapedEmployeeFullName := strings.ReplaceAll(req.EmployeeFullName, "'", "''")
		escapedIP := strings.ReplaceAll(req.IP, "'", "''")

		query := fmt.Sprintf(`
			INSERT INTO reference_ip_phones (address_id, employee_full_name, ip)
			VALUES (%d, '%s', '%s')
		`, req.AddressID, escapedEmployeeFullName, escapedIP)

		if err := ExecDB(query); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Телефон с таким IP адресом уже существует"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлен IP телефон для: " + req.EmployeeFullName)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/ip-phones/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID телефона"})
			return
		}

		var req IPPhoneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		// Валидация обязательных полей
		if req.AddressID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Адрес обязателен для заполнения"})
			return
		}
		if req.EmployeeFullName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ФИО сотрудника обязательно для заполнения"})
			return
		}
		if req.IP == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "IP адрес обязателен для заполнения"})
			return
		}

		// Экранирование значений
		escapedEmployeeFullName := strings.ReplaceAll(req.EmployeeFullName, "'", "''")
		escapedIP := strings.ReplaceAll(req.IP, "'", "''")

		query := fmt.Sprintf(`
			UPDATE reference_ip_phones
			SET address_id = %d, employee_full_name = '%s', ip = '%s'
			WHERE id = %d
		`, req.AddressID, escapedEmployeeFullName, escapedIP, id)

		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлен IP телефон с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/reference/ip-phones/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID телефона"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM reference_ip_phones WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удален IP телефон с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// ==================== ОБРАБОТЧИКИ ДЛЯ УЧЁТНЫХ ЗАПИСЕЙ SSH ====================

func apiCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB("SELECT id, username FROM credentials ORDER BY username")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		if req.Username == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Логин и пароль обязательны"})
			return
		}

		// Шифруем пароль
		encryptedPassword, err := GetCrypto().Encrypt(req.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка шифрования пароля"})
			return
		}

		escapedUsername := strings.ReplaceAll(req.Username, "'", "''")
		escapedPassword := strings.ReplaceAll(encryptedPassword, "'", "''")

		query := fmt.Sprintf("INSERT INTO credentials (username, password) VALUES ('%s', '%s')", escapedUsername, escapedPassword)
		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлена учётная запись SSH: " + req.Username)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/credentials/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID учётной записи"})
			return
		}

		var req struct {
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		if req.Username == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Логин обязателен"})
			return
		}

		escapedUsername := strings.ReplaceAll(req.Username, "'", "''")
		query := fmt.Sprintf("UPDATE credentials SET username = '%s' WHERE id = %d", escapedUsername, id)
		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлена учётная запись SSH с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/credentials/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID учётной записи"})
			return
		}

		// Проверяем, используется ли учётная запись в задачах
		usedInTasks, _ := QueryDB(fmt.Sprintf("SELECT id FROM task_hosts WHERE credential_id = %d LIMIT 1", id))
		if len(usedInTasks) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Нельзя удалить учётную запись, так как она используется в задачах"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM credentials WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удалена учётная запись SSH с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// ==================== ОБРАБОТЧИКИ ДЛЯ СКРИПТОВ ====================

func apiScriptsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB("SELECT id, name, path, parameters_schema, description, created_at FROM scripts ORDER BY name")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req struct {
			Name             string `json:"name"`
			Path             string `json:"path"`
			ParametersSchema string `json:"parameters_schema"`
			Description      string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		if req.Name == "" || req.Path == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Название и путь к скрипту обязательны"})
			return
		}

		// Валидация JSON схемы параметров
		if req.ParametersSchema != "" {
			var tmp interface{}
			if err := json.Unmarshal([]byte(req.ParametersSchema), &tmp); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный JSON в схеме параметров: " + err.Error()})
				return
			}
		} else {
			req.ParametersSchema = "{}"
		}

		escapedName := strings.ReplaceAll(req.Name, "'", "''")
		escapedPath := strings.ReplaceAll(req.Path, "'", "''")
		escapedParams := strings.ReplaceAll(req.ParametersSchema, "'", "''")
		escapedDesc := strings.ReplaceAll(req.Description, "'", "''")

		query := fmt.Sprintf(
			"INSERT INTO scripts (name, path, parameters_schema, description) VALUES ('%s', '%s', '%s', '%s')",
			escapedName, escapedPath, escapedParams, escapedDesc,
		)
		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Добавлен скрипт: " + req.Name)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/scripts/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID скрипта"})
			return
		}

		var req struct {
			Name             string `json:"name"`
			Path             string `json:"path"`
			ParametersSchema string `json:"parameters_schema"`
			Description      string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		if req.Name == "" || req.Path == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Название и путь к скрипту обязательны"})
			return
		}

		// Валидация JSON схемы параметров
		if req.ParametersSchema != "" {
			var tmp interface{}
			if err := json.Unmarshal([]byte(req.ParametersSchema), &tmp); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный JSON в схеме параметров: " + err.Error()})
				return
			}
		} else {
			req.ParametersSchema = "{}"
		}

		escapedName := strings.ReplaceAll(req.Name, "'", "''")
		escapedPath := strings.ReplaceAll(req.Path, "'", "''")
		escapedParams := strings.ReplaceAll(req.ParametersSchema, "'", "''")
		escapedDesc := strings.ReplaceAll(req.Description, "'", "''")

		query := fmt.Sprintf(
			"UPDATE scripts SET name = '%s', path = '%s', parameters_schema = '%s', description = '%s' WHERE id = %d",
			escapedName, escapedPath, escapedParams, escapedDesc, id,
		)
		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Обновлён скрипт с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/scripts/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID скрипта"})
			return
		}

		// Проверяем, используется ли скрипт в задачах
		usedInTasks, _ := QueryDB(fmt.Sprintf("SELECT id FROM task_scripts WHERE script_id = %d LIMIT 1", id))
		if len(usedInTasks) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Нельзя удалить скрипт, так как он используется в задачах"})
			return
		}

		if err := ExecDB(fmt.Sprintf("DELETE FROM scripts WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных"})
			return
		}

		logPanel("Удалён скрипт с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// ==================== ОБРАБОТЧИКИ ДЛЯ ЗАДАЧ ====================

func apiTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := QueryDB(`
			SELECT id, name, description, status, 
			       created_at, updated_at 
			FROM tasks 
			ORDER BY created_at DESC
		`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
			return
		}
		json.NewEncoder(w).Encode(rows)

	case "POST":
		var req struct {
			Name          string                   `json:"name"`
			Description   string                   `json:"description"`
			CredentialID  int                      `json:"credential_id"`
			HostIDs       []int                    `json:"host_ids"`
			BashCommands  []string                 `json:"bash_commands"`
			Scripts       []map[string]interface{} `json:"scripts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Название задачи обязательно"})
			return
		}
		if req.CredentialID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Учётная запись обязательна"})
			return
		}
		if len(req.HostIDs) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Необходимо выбрать хотя бы один хост"})
			return
		}
		if len(req.BashCommands) == 0 && len(req.Scripts) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Необходимо добавить хотя бы одну команду bash или скрипт"})
			return
		}

		// Экранируем название и описание
		escapedName := strings.ReplaceAll(req.Name, "'", "''")
		escapedDesc := strings.ReplaceAll(req.Description, "'", "''")

		// Создаём задачу
		query := fmt.Sprintf(
			"INSERT INTO tasks (name, description, status) VALUES ('%s', '%s', 'pending')",
			escapedName, escapedDesc,
		)
		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка создания задачи"})
			return
		}

		// Получаем ID новой задачи
		rows, err := QueryDB("SELECT last_insert_rowid() as id")
		if err != nil || len(rows) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка получения ID задачи"})
			return
		}
		taskID, _ := strconv.Atoi(rows[0]["id"])

		// Добавляем хосты в задачу
		for _, hostID := range req.HostIDs {
			query = fmt.Sprintf(
				"INSERT INTO task_hosts (task_id, host_id, credential_id) VALUES (%d, %d, %d)",
				taskID, hostID, req.CredentialID,
			)
			if err := ExecDB(query); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка добавления хоста в задачу"})
				// Откатываем создание задачи
				ExecDB(fmt.Sprintf("DELETE FROM tasks WHERE id = %d", taskID))
				return
			}
		}

		// Добавляем команды bash
		for idx, cmd := range req.BashCommands {
			escapedCmd := strings.ReplaceAll(cmd, "'", "''")
			query = fmt.Sprintf(
				"INSERT INTO task_bash_commands (task_id, command, order_index) VALUES (%d, '%s', %d)",
				taskID, escapedCmd, idx,
			)
			if err := ExecDB(query); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка добавления команды bash"})
				// Откатываем создание задачи
				ExecDB(fmt.Sprintf("DELETE FROM tasks WHERE id = %d", taskID))
				return
			}
		}

		// Добавляем скрипты
		for idx, script := range req.Scripts {
			scriptIDFloat, ok := script["script_id"].(float64)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID скрипта"})
				// Откатываем создание задачи
				ExecDB(fmt.Sprintf("DELETE FROM tasks WHERE id = %d", taskID))
				return
			}
			scriptID := int(scriptIDFloat)
			
			params, ok := script["parameters"].(string)
			if !ok {
				params = "{}"
			}
			
			escapedParams := strings.ReplaceAll(params, "'", "''")
			query = fmt.Sprintf(
				"INSERT INTO task_scripts (task_id, script_id, parameters, order_index) VALUES (%d, %d, '%s', %d)",
				taskID, scriptID, escapedParams, idx,
			)
			if err := ExecDB(query); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка добавления скрипта"})
				// Откатываем создание задачи
				ExecDB(fmt.Sprintf("DELETE FROM tasks WHERE id = %d", taskID))
				return
			}
		}

		logPanel("Создана задача: " + req.Name)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "task_id": taskID})

	case "PUT":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID задачи"})
			return
		}

		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Название задачи обязательно"})
			return
		}

		escapedName := strings.ReplaceAll(req.Name, "'", "''")
		escapedDesc := strings.ReplaceAll(req.Description, "'", "''")

		query := fmt.Sprintf(
			"UPDATE tasks SET name = '%s', description = '%s' WHERE id = %d",
			escapedName, escapedDesc, id,
		)
		if err := ExecDB(query); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка обновления задачи"})
			return
		}

		logPanel("Обновлена задача с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	case "DELETE":
		idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID задачи"})
			return
		}

		// Удаляем задачу (каскадное удаление удалит связанные записи)
		if err := ExecDB(fmt.Sprintf("DELETE FROM tasks WHERE id = %d", id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка удаления задачи"})
			return
		}

		logPanel("Удалена задача с ID " + strconv.Itoa(id))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// ==================== УПРАВЛЕНИЕ ЗАДАЧАМИ (СТАРТ/СТОП) ====================

func apiTaskControlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/control")
	taskID, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный ID задачи"})
		return
	}

	var req TaskControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
		return
	}

	switch req.Action {
	case "start":
		// Здесь должна быть логика запуска задачи (в реальном приложении)
		// Для упрощения просто обновляем статус
		ExecDB(fmt.Sprintf("UPDATE tasks SET status = 'running', updated_at = CURRENT_TIMESTAMP WHERE id = %d", taskID))
		logPanel("Запущена задача #" + strconv.Itoa(taskID))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Задача запущена"})

	case "stop":
		// Здесь должна быть логика остановки задачи (в реальном приложении)
		ExecDB(fmt.Sprintf("UPDATE tasks SET status = 'canceled', updated_at = CURRENT_TIMESTAMP WHERE id = %d", taskID))
		logPanel("Остановлена задача #" + strconv.Itoa(taskID))
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Задача остановлена"})

	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректное действие. Допустимые значения: start, stop"})
	}
}

// ==================== ВЫПОЛНЕНИЕ КОМАНД НАПРЯМУЮ (BASH-КОНСТРУКТОР) ====================

func apiExecuteCommandHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || !CheckSession(cookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Требуется аутентификация"})
		return
	}

	var req struct {
		HostID       int    `json:"host_id"`
		CredentialID int    `json:"credential_id"`
		Command      string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный формат JSON: " + err.Error()})
		return
	}

	logDiagnostic("BASH-конструктор: Получен запрос на выполнение команды")
	logDiagnostic("  HostID: " + strconv.Itoa(req.HostID))
	logDiagnostic("  CredentialID: " + strconv.Itoa(req.CredentialID))
	logDiagnostic("  Command (первые 50 символов): " + req.Command[:min(50, len(req.Command))])

	if req.HostID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Выберите узел для выполнения команды"})
		return
	}

	if req.CredentialID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Выберите учётную запись для подключения"})
		return
	}

	if req.Command == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Введите команду для выполнения"})
		return
	}

	// Получаем данные хоста из нового справочника
	query := fmt.Sprintf(`
		SELECT h.id, h.ip, h.ssh_port, h.enabled,
		       a.street, a.building, a.cabinet, a.corridor, a.floor, a.service_room,
		       e.full_name as employee_full_name, e.short_name as employee_short_name,
		       c.username, c.password
		FROM reference_hosts h
		JOIN reference_addresses a ON h.address_id = a.id
		LEFT JOIN reference_employees e ON h.employee_id = e.id
		JOIN credentials c ON c.id = %d
		WHERE h.id = %d AND h.enabled = 1
	`, req.CredentialID, req.HostID)

	logDiagnostic("BASH-конструктор: Сформирован запрос к БД")
	logDiagnostic("  Query: " + query)

	rows, err := QueryDB(query)
	if err != nil {
		logDiagnostic("BASH-конструктор: Ошибка выполнения запроса к БД: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка базы данных: " + err.Error()})
		return
	}

	logDiagnostic("BASH-конструктор: Получено строк из БД: " + strconv.Itoa(len(rows)))

	if len(rows) == 0 {
		hostCheck, _ := QueryDB(fmt.Sprintf("SELECT id FROM reference_hosts WHERE id = %d", req.HostID))
		credCheck, _ := QueryDB(fmt.Sprintf("SELECT id FROM credentials WHERE id = %d", req.CredentialID))
		
		var errorMsg string
		if len(hostCheck) == 0 {
			errorMsg = "Хост с ID " + strconv.Itoa(req.HostID) + " не найден или отключён"
			logDiagnostic("BASH-конструктор: Хост не найден. Проверка: " + errorMsg)
		} else if len(credCheck) == 0 {
			errorMsg = "Учётная запись с ID " + strconv.Itoa(req.CredentialID) + " не найдена"
			logDiagnostic("BASH-конструктор: Учётная запись не найдена. Проверка: " + errorMsg)
		} else {
			errorMsg = "Хост и учётная запись найдены, но не могут быть связаны"
			logDiagnostic("BASH-конструктор: Хост и учётная запись найдены, но связь не установлена")
		}
		
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
		return
	}

	host := rows[0]
	fullName := host["employee_short_name"]
	if fullName == "" {
		fullName = host["employee_full_name"]
	}
	ip := host["ip"]
	portStr := host["ssh_port"]
	username := host["username"]
	encryptedPassword := host["password"]

	logDiagnostic("BASH-конструктор: Найден хост для подключения")
	logDiagnostic("  FullName: " + fullName)
	logDiagnostic("  IP: " + ip)
	logDiagnostic("  Port: " + portStr)
	logDiagnostic("  Username: " + username)

	decryptedPassword, err := GetCrypto().Decrypt(encryptedPassword)
	if err != nil {
		logDiagnostic("BASH-конструктор: Ошибка дешифрования пароля: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка дешифрования пароля"})
		return
	}

	port, _ := strconv.Atoi(portStr)
	
	logRemote("BASH-конструктор: Выполнение команды на хосте '" + fullName + "' (" + ip + "): " + req.Command)
	logDiagnostic("BASH-конструктор: Начало выполнения SSH команды")
	
	output, err := ExecuteSSHCommand(username, decryptedPassword, ip, port, req.Command)
	
	if err != nil {
		logDiagnostic("BASH-конструктор: Ошибка выполнения SSH команды: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "failed",
			"error":  "Ошибка выполнения команды: " + err.Error(),
			"output": output,
		})
		return
	}

	logDiagnostic("BASH-конструктор: Команда выполнена успешно. Длина вывода: " + strconv.Itoa(len(output)))
	logRemote("BASH-конструктор: Успешно выполнена команда на хосте '" + fullName + "' (" + ip + "): " + req.Command)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "completed",
		"output": output,
		"host":   fullName + " (" + ip + ")",
	})
}

// ==================== НАСТРОЙКА МАРШРУТОВ ====================

func setupRoutes() {
// Инициализация шаблонов
if err := initTemplates(); err != nil {
log.Fatalf("Ошибка инициализации шаблонов: %v", err)
}

// Статические файлы из embedded FS
http.Handle("/static/", frontend.StaticFileServer(http.FS(frontend.StaticFS)))

http.HandleFunc("/", indexHandler)
http.HandleFunc("/login", loginHandler)
http.HandleFunc("/logout", logoutHandler)

// Справочники
http.HandleFunc("/api/reference/addresses", authMiddleware(apiReferenceAddressesHandler))
http.HandleFunc("/api/reference/addresses/", authMiddleware(apiReferenceAddressesHandler))
http.HandleFunc("/api/reference/employees", authMiddleware(apiReferenceEmployeesHandler))
http.HandleFunc("/api/reference/employees/", authMiddleware(apiReferenceEmployeesHandler))
http.HandleFunc("/api/reference/workstations", authMiddleware(apiReferenceWorkstationsHandler))
http.HandleFunc("/api/reference/workstations/", authMiddleware(apiReferenceWorkstationsHandler))
http.HandleFunc("/api/reference/hosts", authMiddleware(apiReferenceHostsHandler))
http.HandleFunc("/api/reference/hosts/", authMiddleware(apiReferenceHostsHandler))
http.HandleFunc("/api/reference/network-equipment", authMiddleware(apiReferenceNetworkEquipmentHandler))
http.HandleFunc("/api/reference/network-equipment/", authMiddleware(apiReferenceNetworkEquipmentHandler))
http.HandleFunc("/api/reference/network-mfps", authMiddleware(apiReferenceNetworkMFPsHandler))
http.HandleFunc("/api/reference/network-mfps/", authMiddleware(apiReferenceNetworkMFPsHandler))
http.HandleFunc("/api/reference/ip-phones", authMiddleware(apiReferenceIPPhonesHandler))
http.HandleFunc("/api/reference/ip-phones/", authMiddleware(apiReferenceIPPhonesHandler))

// Существующие эндпоинты (обновлены для работы с новой схемой)
http.HandleFunc("/api/credentials", authMiddleware(apiCredentialsHandler))
http.HandleFunc("/api/credentials/", authMiddleware(apiCredentialsHandler))
http.HandleFunc("/api/scripts", authMiddleware(apiScriptsHandler))
http.HandleFunc("/api/scripts/", authMiddleware(apiScriptsHandler))
http.HandleFunc("/api/tasks", authMiddleware(apiTasksHandler))
http.HandleFunc("/api/tasks/", authMiddleware(apiTaskControlHandler))
http.HandleFunc("/api/logs", authMiddleware(apiLogsHandler))
http.HandleFunc("/api/execute-command", authMiddleware(apiExecuteCommandHandler))
http.HandleFunc("/api/user", authMiddleware(apiUserHandler))

// Обработчик для загрузки контента разделов через HTMX
http.HandleFunc("/api/content/", authMiddleware(contentSectionHandler))
}

func main() {
	if err := InitDB(); err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}

	// Расшифровываем футер при запуске сервера
	footerRows, err := QueryDB("SELECT setting_value FROM platform_settings WHERE setting_key = 'footer_text'")
	if err != nil || len(footerRows) == 0 {
		footerText = "Powered by Karagapolov A.E." // Резервный вариант
		logDiagnostic("Не удалось загрузить футер из БД, используется резервный вариант")
	} else {
		encryptedFooter := footerRows[0]["setting_value"]
		decryptedFooter, err := GetCrypto().Decrypt(encryptedFooter)
		if err != nil {
			footerText = "Powered by Karagapolov A.E."
			logDiagnostic("Ошибка расшифровки футера: " + err.Error())
		} else {
			footerText = decryptedFooter
			logPanel("Футер платформы успешно расшифрован и загружен")
		}
	}

	setupRoutes()

	port := DEFAULT_PORT
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil && p > 0 && p < 65536 {
			port = p
		}
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p < 65536 {
			port = p
		}
	}

	logPanel("SSH Orchestrator запущен (версия " + CURRENT_SCHEMA_VERSION + ")")

	// Создаём папку для логов при старте
	if err := os.MkdirAll(LOGS_DIR, 0755); err != nil {
		log.Fatalf("Ошибка создания папки логов: %v", err)
	}

	fmt.Printf("SSH Orchestrator запущен (версия %s)\n", CURRENT_SCHEMA_VERSION)
	fmt.Printf("  -> Веб-интерфейс: http://localhost:%d\n", port)
	fmt.Printf("  -> База данных: %s\n", DB_PATH)
	fmt.Printf("  -> Мастер-ключ: %s (права 600)\n", KEY_PATH)
	fmt.Printf("  -> Логи: папка %s\n", LOGS_DIR)
	fmt.Printf("  -> Первый администратор: логин='admin', пароль='admin' (СМЕНИТЕ ПАРОЛЬ!)\n")
	fmt.Printf("  -> Футер: %s\n", footerText)
	fmt.Printf("  -> Для остановки: Ctrl+C\n\n")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var loginHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>SSH Orchestrator - Вход</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f5f5f5; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
        .login-box { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 400px; }
        .login-box h2 { text-align: center; margin-bottom: 30px; color: #333; }
        .form-group { margin-bottom: 20px; }
        .form-group label { display: block; margin-bottom: 8px; font-weight: 500; color: #555; }
        .form-group input { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 16px; }
        .btn { width: 100%; padding: 12px; background: #E95420; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
        .btn:hover { background: #C3441C; }
        .alert { padding: 10px; border-radius: 4px; margin-bottom: 15px; display: none; }
        .alert-danger { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
    </style>
</head>
<body>
    <div class="login-box">
        <h2>SSH Orchestrator</h2>
        <div id="alert" class="alert alert-danger"></div>
        <form id="login-form">
            <div class="form-group">
                <label for="username">Логин</label>
                <input type="text" id="username" name="username" required autofocus>
            </div>
            <div class="form-group">
                <label for="password">Пароль</label>
                <input type="password" id="password" name="password" required>
            </div>
            <button type="submit" class="btn">Войти</button>
        </form>
    </div>
    <script>
        document.getElementById('login-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const alert = document.getElementById('alert');
            alert.style.display = 'none';
            const response = await fetch('/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username: document.getElementById('username').value,
                    password: document.getElementById('password').value
                })
            });
            const result = await response.json();
            if (response.ok) {
                window.location.href = '/';
            } else {
                alert.textContent = result.error || 'Ошибка входа';
                alert.style.display = 'block';
            }
        });
    </script>
</body>
</html>`