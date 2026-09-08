package main

import (
	"fmt"
	"strings"
)

// AddressRequest представляет запрос на создание/обновление адреса
type AddressRequest struct {
	Street       string `json:"street" validate:"required"`
	Building     string `json:"building" validate:"required"`
	Cabinet      string `json:"cabinet"`
	Corridor     string `json:"corridor"`
	Floor        int    `json:"floor"`
	ServiceRoom  string `json:"service_room"`
}

// Validate проверяет корректность данных адреса
func (a AddressRequest) Validate() error {
	// Проверяем, что заполнено ровно одно из полей: кабинет, коридор или служебное помещение
	hasCabinet := a.Cabinet != ""
	hasCorridor := a.Corridor != ""
	hasServiceRoom := a.ServiceRoom != ""

	count := 0
	if hasCabinet {
		count++
	}
	if hasCorridor {
		count++
	}
	if hasServiceRoom {
		count++
	}

	if count == 0 {
		return fmt.Errorf("необходимо указать кабинет, коридор или служебное помещение")
	}
	if count > 1 {
		return fmt.Errorf("можно указать только один тип адреса: кабинет, коридор или служебное помещение")
	}

	// Если выбран коридор, этаж обязателен
	if hasCorridor && a.Floor == 0 {
		return fmt.Errorf("при выборе коридора необходимо указать этаж")
	}

	// Если выбран коридор, этаж должен быть в диапазоне 1-5
	if hasCorridor && (a.Floor < 1 || a.Floor > 5) {
		return fmt.Errorf("этаж должен быть в диапазоне от 1 до 5")
	}

	// Если выбрано служебное помещение, проверяем его значение
	// (это можно сделать через список допустимых значений или оставить как есть)

	// Если выбран кабинет, проверяем его формат (опционально)
	// Например, можно проверить, что кабинет состоит из цифр или букв

	// Валидация пройдена
	return nil
}

// WorkstationRequest представляет запрос на создание/обновление АРМ
type WorkstationRequest struct {
	AddressID        int    `json:"address_id" validate:"required"`
	IsVacant         bool   `json:"is_vacant"`
	EmployeeID       int    `json:"employee_id"`
	InventoryNumber  string `json:"inventory_number" validate:"required"`
	SealNumbers      string `json:"seal_numbers"`
	MonitorCount     int    `json:"monitor_count" validate:"required,min=1,max=5"`
	SerialNumber     string `json:"serial_number"`  // ИСПРАВЛЕНО: убрана валидация required
	ReplacementDone  bool   `json:"replacement_done"`
	ReplacementDate  string `json:"replacement_date"`
	ReplacementLetter string `json:"replacement_letter"`
}

// Validate проверяет корректность данных АРМ
func (w WorkstationRequest) Validate() error {
	// ИСПРАВЛЕНО: серийный номер НЕ обязателен - убрана проверка

	// Проверяем инвентарный номер
	if w.InventoryNumber == "" {
		return fmt.Errorf("инвентарный номер обязателен для заполнения")
	}

	// Проверяем количество мониторов
	if w.MonitorCount < 1 || w.MonitorCount > 5 {
		return fmt.Errorf("количество мониторов должно быть от 1 до 5")
	}

	// Проверяем валидность поля вакансии
	if w.IsVacant && w.EmployeeID != 0 {
		return fmt.Errorf("вакантное место не может быть привязано к сотруднику")
	}

	// Проверяем дату замены, если замена выполнена
	if w.ReplacementDone && w.ReplacementDate == "" {
		return fmt.Errorf("при отметке о замене необходимо указать дату замены")
	}

	// Проверяем, что сотрудник существует (если указан)
	if w.EmployeeID != 0 {
		rows, err := QueryDB(fmt.Sprintf("SELECT id FROM reference_employees WHERE id = %d", w.EmployeeID))
		if err != nil || len(rows) == 0 {
			return fmt.Errorf("указанный сотрудник не найден")
		}
	}

	// Проверяем, что адрес существует
	rows, err := QueryDB(fmt.Sprintf("SELECT id FROM reference_addresses WHERE id = %d", w.AddressID))
	if err != nil || len(rows) == 0 {
		return fmt.Errorf("указанный адрес не найден")
	}

	// Валидация пройдена
	return nil
}

// EmployeeRequest представляет запрос на создание/обновление сотрудника
type EmployeeRequest struct {
	FullName      string `json:"full_name" validate:"required"`
	ShortName     string `json:"short_name"`
	PhoneCity     string `json:"phone_city"`
	PhoneInternal string `json:"phone_internal"`
	AddressID     int    `json:"address_id" validate:"required"`
}

// Validate проверяет корректность данных сотрудника
func (e EmployeeRequest) Validate() error {
	// Проверяем ФИО
	if e.FullName == "" {
		return fmt.Errorf("ФИО полностью обязательно для заполнения")
	}

	// Проверяем адрес
	if e.AddressID == 0 {
		return fmt.Errorf("адрес обязателен для заполнения")
	}

	// Проверяем, что адрес не является служебным помещением
	// (сотрудники в служебных помещениях не учитываются)
	// Это проверяется на уровне БД или при сохранении

	// Валидация пройдена
	return nil
}

// HostRequest представляет запрос на создание/обновление хоста
type HostRequest struct {
	AddressID   int    `json:"address_id" validate:"required"`
	EmployeeID  int    `json:"employee_id"`
	IP          string `json:"ip" validate:"required,ip"`
	SSHPort     int    `json:"ssh_port"`
	Enabled     bool   `json:"enabled"`
}

// Validate проверяет корректность данных хоста
func (h HostRequest) Validate() error {
	// Проверяем адрес
	if h.AddressID == 0 {
		return fmt.Errorf("адрес обязателен для заполнения")
	}

	// Проверяем IP адрес
	// (валидация формата выполняется через тег "ip" в структуре)

	// Проверяем, что хост с таким IP не существует (кроме обновления)
	// Это проверяется на уровне БД через уникальность

	// Валидация пройдена
	return nil
}

// NetworkEquipmentRequest представляет запрос на создание/обновление сетевого оборудования
type NetworkEquipmentRequest struct {
	AddressID int    `json:"address_id" validate:"required"`
	Category  string `json:"category" validate:"required"`
	Type      string `json:"type" validate:"required"`
	Model     string `json:"model" validate:"required"`
	PortCount int    `json:"port_count" validate:"required,min=1"`
}

// Validate проверяет корректность данных сетевого оборудования
func (n NetworkEquipmentRequest) Validate() error {
	// Проверяем адрес
	if n.AddressID == 0 {
		return fmt.Errorf("адрес обязателен для заполнения")
	}

	// Проверяем категорию
	if n.Category == "" {
		return fmt.Errorf("категория обязательна для заполнения")
	}

	// Проверяем тип
	if n.Type == "" {
		return fmt.Errorf("тип обязателен для заполнения")
	}

	// Проверяем модель
	if n.Model == "" {
		return fmt.Errorf("модель обязательна для заполнения")
	}

	// Проверяем количество портов
	if n.PortCount <= 0 {
		return fmt.Errorf("число портов должно быть больше 0")
	}

	// Валидация пройдена
	return nil
}

// NetworkMFPRequest представляет запрос на создание/обновление сетевого МФУ
type NetworkMFPRequest struct {
	AddressID    int    `json:"address_id" validate:"required"`
	Model        string `json:"model" validate:"required"`
	IP           string `json:"ip" validate:"required,ip"`
	Hostname     string `json:"hostname"`
	SerialNumber string `json:"serial_number" validate:"required"`
}

// Validate проверяет корректность данных сетевого МФУ
func (n NetworkMFPRequest) Validate() error {
	// Проверяем адрес
	if n.AddressID == 0 {
		return fmt.Errorf("адрес обязателен для заполнения")
	}

	// Проверяем модель
	if n.Model == "" {
		return fmt.Errorf("модель обязательна для заполнения")
	}

	// Проверяем IP адрес
	// (валидация формата выполняется через тег "ip" в структуре)

	// Проверяем серийный номер
	// (валидация обязательности выполняется через тег "required" в структуре)

	// Валидация пройдена
	return nil
}

// IPPhoneRequest представляет запрос на создание/обновление IP телефона
type IPPhoneRequest struct {
	AddressID        int    `json:"address_id" validate:"required"`
	EmployeeFullName string `json:"employee_full_name" validate:"required"`
	IP               string `json:"ip" validate:"required,ip"`
}

// Validate проверяет корректность данных IP телефона
func (i IPPhoneRequest) Validate() error {
	// Проверяем адрес
	if i.AddressID == 0 {
		return fmt.Errorf("адрес обязателен для заполнения")
	}

	// Проверяем ФИО сотрудника
	if i.EmployeeFullName == "" {
		return fmt.Errorf("ФИО сотрудника обязательно для заполнения")
	}

	// Проверяем IP адрес
	// (валидация формата выполняется через тег "ip" в структуре)

	// Валидация пройдена
	return nil
}

// TaskControlRequest представляет запрос на управление задачей (старт/стоп)
type TaskControlRequest struct {
	Action string `json:"action" validate:"required,oneof=start stop"`
}

// Validate проверяет корректность запроса управления задачей
func (t TaskControlRequest) Validate() error {
	// Проверяем действие
	// (валидация выполняется через тег "oneof=start stop" в структуре)

	// Валидация пройдена
	return nil
}

// Форматирование полного адреса
func FormatFullAddress(street, building, cabinet, corridor, serviceRoom string, floor int) string {
	parts := []string{}
	if street != "" {
		parts = append(parts, street)
	}
	if building != "" {
		parts = append(parts, building)
	}
	if cabinet != "" {
		parts = append(parts, "каб."+cabinet)
	} else if corridor != "" {
		floorStr := ""
		if floor > 0 {
			floorStr = fmt.Sprintf(" эт.%d", floor)
		}
		parts = append(parts, "кор."+corridor+floorStr)
	} else if serviceRoom != "" {
		parts = append(parts, serviceRoom)
	}
	return strings.Join(parts, ", ")
}