package main

import (
"fmt"
"strings"
)

// AddressRequest представляет запрос на создание/обновление адреса
type AddressRequest struct {
Street      string `json:"street"`
Building    string `json:"building"`
Cabinet     string `json:"cabinet"`      // Номер кабинета ИЛИ номер гаража
Corridor    string `json:"corridor"`
Floor       string `json:"floor"`
ServiceRoom string `json:"service_room"` // Тип служебного помещения: server_room, warehouse, office, garage, other
}

// Validate проверяет корректность данных адреса
func (a AddressRequest) Validate() error {
// Обязательные поля для любого адреса
if a.Street == "" || a.Building == "" {
return fmt.Errorf("улица и дом обязательны для заполнения")
}

// Проверяем, что заполнено хотя бы одно поле типа адреса
hasCabinet := a.Cabinet != ""
hasCorridor := a.Corridor != ""
hasServiceRoom := a.ServiceRoom != ""

if !hasCabinet && !hasCorridor && !hasServiceRoom {
return fmt.Errorf("необходимо указать тип адреса: кабинет, коридор или служебное помещение")
}

// Если выбран коридор, этаж обязателен
if hasCorridor && a.Floor == "" {
return fmt.Errorf("при выборе коридора необходимо указать этаж")
}

// Если выбран коридор, этаж должен быть в диапазоне 1-5
if hasCorridor {
var floor int
fmt.Sscanf(a.Floor, "%d", &floor)
if floor < 1 || floor > 5 {
return fmt.Errorf("этаж должен быть в диапазоне от 1 до 5")
}
}

// Валидация пройдена
return nil
}

// WorkstationRequest представляет запрос на создание/обновление АРМ
type WorkstationRequest struct {
AddressID         int    `json:"address_id"`
IsVacant          bool   `json:"is_vacant"`
EmployeeID        int    `json:"employee_id"`
InventoryNumber   string `json:"inventory_number"`
SealNumbers       string `json:"seal_numbers"`
MonitorCount      int    `json:"monitor_count"`
SerialNumber      string `json:"serial_number"`
ReplacementDone   bool   `json:"replacement_done"`
ReplacementDate   string `json:"replacement_date"`
ReplacementLetter string `json:"replacement_letter"`
}

// Validate проверяет корректность данных АРМ
func (w WorkstationRequest) Validate() error {
if w.InventoryNumber == "" {
return fmt.Errorf("инвентарный номер обязателен")
}
if w.MonitorCount < 1 || w.MonitorCount > 5 {
return fmt.Errorf("количество мониторов должно быть от 1 до 5")
}
if w.ReplacementDone && w.ReplacementDate == "" {
return fmt.Errorf("при отметке о замене необходимо указать дату")
}
return nil
}

// EmployeeRequest представляет запрос на создание/обновление сотрудника
type EmployeeRequest struct {
FullName      string `json:"full_name"`
ShortName     string `json:"short_name"`
PhoneCity     string `json:"phone_city"`
PhoneInternal string `json:"phone_internal"`
AddressID     int    `json:"address_id"`
}

// Validate проверяет корректность данных сотрудника
func (e EmployeeRequest) Validate() error {
if e.FullName == "" {
return fmt.Errorf("ФИО обязательно")
}
if e.AddressID == 0 {
return fmt.Errorf("адрес обязателен")
}
return nil
}

// HostRequest представляет запрос на создание/обновление хоста
type HostRequest struct {
AddressID  int    `json:"address_id"`
EmployeeID int    `json:"employee_id"`
IP         string `json:"ip"`
SSHPort    int    `json:"ssh_port"`
Enabled    bool   `json:"enabled"`
}

// Validate проверяет корректность данных хоста
func (h HostRequest) Validate() error {
if h.AddressID == 0 {
return fmt.Errorf("адрес обязателен")
}
if h.IP == "" {
return fmt.Errorf("IP адрес обязателен")
}
return nil
}

// NetworkEquipmentRequest представляет запрос на создание/обновление сетевого оборудования
type NetworkEquipmentRequest struct {
AddressID int    `json:"address_id"`
Category  string `json:"category"`
Type      string `json:"type"`
Model     string `json:"model"`
PortCount int    `json:"port_count"`
}

// Validate проверяет корректность данных сетевого оборудования
func (n NetworkEquipmentRequest) Validate() error {
if n.AddressID == 0 {
return fmt.Errorf("адрес обязателен")
}
if n.Category == "" || n.Type == "" || n.Model == "" {
return fmt.Errorf("категория, тип и модель обязательны")
}
if n.PortCount <= 0 {
return fmt.Errorf("число портов должно быть больше 0")
}
return nil
}

// NetworkMFPRequest представляет запрос на создание/обновление сетевого МФУ
type NetworkMFPRequest struct {
AddressID    int    `json:"address_id"`
Model        string `json:"model"`
IP           string `json:"ip"`
Hostname     string `json:"hostname"`
SerialNumber string `json:"serial_number"`
}

// Validate проверяет корректность данных сетевого МФУ
func (n NetworkMFPRequest) Validate() error {
if n.AddressID == 0 {
return fmt.Errorf("адрес обязателен")
}
if n.Model == "" || n.IP == "" || n.SerialNumber == "" {
return fmt.Errorf("модель, IP и серийный номер обязательны")
}
return nil
}

// IPPhoneRequest представляет запрос на создание/обновление IP телефона
type IPPhoneRequest struct {
AddressID        int    `json:"address_id"`
EmployeeFullName string `json:"employee_full_name"`
IP               string `json:"ip"`
}

// Validate проверяет корректность данных IP телефона
func (i IPPhoneRequest) Validate() error {
if i.AddressID == 0 {
return fmt.Errorf("адрес обязателен")
}
if i.EmployeeFullName == "" || i.IP == "" {
return fmt.Errorf("ФИО сотрудника и IP обязательны")
}
return nil
}

// Форматирование полного адреса
func FormatFullAddress(street, building, cabinet, corridor, serviceRoom, garageNumber, floor string) string {
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
if floor != "" {
var f int
fmt.Sscanf(floor, "%d", &f)
if f > 0 {
floorStr = fmt.Sprintf(" эт.%d", f)
}
}
parts = append(parts, "кор."+corridor+floorStr)
} else if serviceRoom != "" {
if serviceRoom == "garage" && cabinet != "" {
// Для гаража cabinet содержит номер гаража
parts = append(parts, "гараж "+cabinet)
} else {
parts = append(parts, serviceRoom)
}
}
return strings.Join(parts, ", ")
}
