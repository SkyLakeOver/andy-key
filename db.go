package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_DB_PATH        = "data.db"
	DEFAULT_LOGS_DIR       = "logs"
	LOG_PREFIX             = "andy-key"
	CURRENT_SCHEMA_VERSION = "0.0.4.2"
)

var (
	DB_PATH          string
	LOGS_DIR         string
	db               *DB
	panelLog         *os.File
	remoteLog        *os.File
	diagnosticLog    *os.File
)

// loadConfig загружает конфигурацию из файла config.conf
func loadConfig() error {
	// Устанавливаем значения по умолчанию
	DB_PATH = DEFAULT_DB_PATH
	LOGS_DIR = DEFAULT_LOGS_DIR

	configFile := "config.conf"
	file, err := os.Open(configFile)
	if err != nil {
		// Если файл не найден, используем значения по умолчанию
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("ошибка открытия конфигурационного файла: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "DB_PATH":
			DB_PATH = value
		case "LOGS_DIR":
			LOGS_DIR = value
		}
	}

	return scanner.Err()
}

// migrateOldDB проверяет наличие старых файлов БД и переименовывает их
func migrateOldDB() error {
	// Если текущий файл БД уже существует, ничего не делаем
	if _, err := os.Stat(DB_PATH); err == nil {
		return nil
	}

	// Список возможных старых имен файлов БД
	oldNames := []string{"andy-key.db", "sshkage.db"}

	for _, oldName := range oldNames {
		if _, err := os.Stat(oldName); err == nil {
			// Старый файл существует, переименовываем его
			if err := os.Rename(oldName, DB_PATH); err != nil {
				return fmt.Errorf("ошибка переименования %s в %s: %v", oldName, DB_PATH, err)
			}
			logPanel(fmt.Sprintf("База данных переименована: %s → %s", oldName, DB_PATH))
			return nil
		}
	}

	return nil
}

type DB struct {
	crypto *Crypto
}

// =============== СИСТЕМА ВЕРСИОНИРОВАНИЯ И МИГРАЦИЙ ===============
func ensureSchemaVersion() error {
	logDiagnostic("=== НАЧАЛО ПРОВЕРКИ ВЕРСИИ СХЕМЫ БАЗЫ ДАННЫХ ===")
	
	// Шаг 1: Проверяем наличие таблицы версий
	tableExists, err := tableExists("schema_version")
	if err != nil {
		return fmt.Errorf("ошибка проверки наличия таблицы версий: %v", err)
	}
	if !tableExists {
		logDiagnostic("Таблица schema_version отсутствует. Определяем текущую версию БД по структуре...")
		return initializeSchemaVersion()
	}
	
	// Шаг 2: Получаем текущую версию из БД
	versionRows, err := QueryDB("SELECT version FROM schema_version ORDER BY updated_at DESC LIMIT 1")
	if err != nil {
		return fmt.Errorf("ошибка чтения версии схемы из базы данных: %v", err)
	}
	
	// ИСПРАВЛЕНО: обработка случая, когда таблица существует, но пустая
	if len(versionRows) == 0 {
		logDiagnostic("Таблица schema_version существует, но не содержит записей. Инициализируем версию...")
		// Автоопределение текущей версии по структуре
		currentVersion := detectDBVersion()
		logDiagnostic("Автоопределение версии БД по структуре: " + currentVersion)
		// Вставляем определенную версию (не текущую приложения!)
		escapedVersion := strings.ReplaceAll(currentVersion, "'", "''")
		if err := ExecDB("INSERT INTO schema_version (version) VALUES ('" + escapedVersion + "')"); err != nil {
			return fmt.Errorf("ошибка записи версии схемы: %v", err)
		}
		logPanel("Инициализирована версия схемы базы данных: " + currentVersion)
		logDiagnostic("=== ПРОВЕРКА ВЕРСИИ ЗАВЕРШЕНА: СХЕМА ИНИЦИАЛИЗИРОВАНА ===")
		// Теперь выполняем миграции от определенной версии до текущей
		if compareVersions(currentVersion, CURRENT_SCHEMA_VERSION) < 0 {
			logDiagnostic("Требуются миграции от версии " + currentVersion + " до " + CURRENT_SCHEMA_VERSION)
			if err := migrateDB(currentVersion, CURRENT_SCHEMA_VERSION); err != nil {
				return err
			}
		}
		return nil
	}
	
	currentDBVersion := versionRows[0]["version"]
	logDiagnostic("Текущая версия схемы БД: " + currentDBVersion)
	logDiagnostic("Требуемая версия схемы: " + CURRENT_SCHEMA_VERSION)
	
	// Шаг 3: Сравниваем версии
	cmp := compareVersions(currentDBVersion, CURRENT_SCHEMA_VERSION)
	if cmp == 0 {
		logPanel("Версия схемы БД актуальна: " + CURRENT_SCHEMA_VERSION)
		logDiagnostic("=== ПРОВЕРКА ВЕРСИИ ЗАВЕРШЕНА: СХЕМА АКТУАЛЬНА ===")
		return nil
	}
	if cmp > 0 {
		// Версия БД новее версии приложения — критическая ошибка
		errorMsg := fmt.Sprintf(
			"КРИТИЧЕСКАЯ ОШИБКА: Версия базы данных (%s) НОВЕЕ версии приложения (%s). "+
				"Обновите приложение до актуальной версии или восстановите резервную копию БД.",
			currentDBVersion,
			CURRENT_SCHEMA_VERSION,
		)
		logDiagnostic("=== КРИТИЧЕСКАЯ ОШИБКА: НЕСОВМЕСТИМОСТЬ ВЕРСИЙ ===")
		logDiagnostic(errorMsg)
		return fmt.Errorf(errorMsg)
	}
	
	// Шаг 4: Выполняем миграции от текущей версии до целевой
	logPanel(fmt.Sprintf("Обнаружена устаревшая версия БД (%s). Выполняем миграции до версии %s...", currentDBVersion, CURRENT_SCHEMA_VERSION))
	if err := migrateDB(currentDBVersion, CURRENT_SCHEMA_VERSION); err != nil {
		logDiagnostic("=== ОШИБКА МИГРАЦИИ: " + err.Error() + " ===")
		return fmt.Errorf("ошибка миграции базы данных: %v", err)
	}
	
	// Шаг 5: Обновляем запись версии в БД
	escapedVersion := strings.ReplaceAll(CURRENT_SCHEMA_VERSION, "'", "''")
	if err := ExecDB("UPDATE schema_version SET version = '" + escapedVersion + "', updated_at = CURRENT_TIMESTAMP"); err != nil {
		return fmt.Errorf("ошибка обновления версии схемы: %v", err)
	}
	
	logPanel("Версия схемы базы данных успешно обновлена до: " + CURRENT_SCHEMA_VERSION)
	logDiagnostic("=== ПРОВЕРКА ВЕРСИИ ЗАВЕРШЕНА: МИГРАЦИЯ УСПЕШНА ===")
	
	return nil
}

// Инициализация таблицы версий для существующих баз данных
func initializeSchemaVersion() error {
	// Определяем текущую версию по наличию ключевых таблиц
	currentVersion := detectDBVersion()
	logDiagnostic("Автоопределение версии БД: " + currentVersion)
	
	// Проверяем наличие всех таблиц для целевой версии
	missingTables := checkRequiredTablesForVersion(CURRENT_SCHEMA_VERSION)
	if len(missingTables) > 0 {
		logDiagnostic(fmt.Sprintf("Обнаружены отсутствующие таблицы для версии %s: %v", CURRENT_SCHEMA_VERSION, missingTables))
		logDiagnostic("Выполняем создание недостающих таблиц...")
		// Создаем недостающие таблицы
		for _, table := range missingTables {
			query := getTableCreationQuery(table, CURRENT_SCHEMA_VERSION)
			if query != "" {
				if err := ExecDB(query); err != nil {
					return fmt.Errorf("ошибка создания таблицы %s: %v", table, err)
				}
				logPanel("Создана недостающая таблица: " + table)
			}
		}
	} else {
		logPanel("Структура БД соответствует версии " + CURRENT_SCHEMA_VERSION)
	}
	
	// Создаем таблицу версий (если её ещё нет)
	if exists, _ := tableExists("schema_version"); !exists {
		if err := ExecDB(`CREATE TABLE IF NOT EXISTS schema_version (
version TEXT NOT NULL,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`); err != nil {
			return fmt.Errorf("ошибка создания таблицы версий: %v", err)
		}
		logDiagnostic("Создана таблица отслеживания версий")
	}
	
	// Записываем определенную версию (НЕ текущую приложения!)
	// Это важно для корректного выполнения миграций
	escapedVersion := strings.ReplaceAll(currentVersion, "'", "''")
	if err := ExecDB("INSERT INTO schema_version (version) VALUES ('" + escapedVersion + "')"); err != nil {
		return fmt.Errorf("ошибка записи версии схемы: %v", err)
	}
	
	logPanel("Инициализирована версия схемы базы данных: " + currentVersion)
	logDiagnostic("Автоопределение завершено. Текущая версия: " + currentVersion)
	
	// Если текущая версия ниже целевой, выполняем миграции
	if compareVersions(currentVersion, CURRENT_SCHEMA_VERSION) < 0 {
		logDiagnostic("Требуются миграции от версии " + currentVersion + " до " + CURRENT_SCHEMA_VERSION)
		if err := migrateDB(currentVersion, CURRENT_SCHEMA_VERSION); err != nil {
			return err
		}
	}
	
	return nil
}

// Автоопределение версии БД по наличию ключевых таблиц и колонок
func detectDBVersion() string {
	// Проверяем наличие таблицы из версии 0.0.4.1 (platform_settings с footer_text)
	if exists, _ := tableExists("reference_addresses"); exists {
		// Дополнительная проверка: наличие platform_settings для определения 0.0.4.1
		if exists, _ := tableExists("platform_settings"); exists {
			return "0.0.4.1"
		}
		return "0.0.4"
	}
	// Проверяем наличие таблицы из версии 0.0.3
	if exists, _ := tableExists("platform_settings"); exists {
		// Дополнительная проверка: наличие футера в настройках
		footerRows, _ := QueryDB("SELECT setting_value FROM platform_settings WHERE setting_key = 'footer_text'")
		if len(footerRows) > 0 {
			return "0.0.3"
		}
	}
	// Проверяем наличие таблицы из версии 0.0.2
	if exists, _ := tableExists("task_bash_commands"); exists {
		return "0.0.2"
	}
	// Проверяем наличие базовых таблиц версии 0.0.1
	// Если есть хотя бы одна из ключевых таблиц, считаем версией 0.0.1
	keyTables := []string{"admin_users", "credentials", "hosts", "tasks"}
	for _, table := range keyTables {
		if exists, _ := tableExists(table); exists {
			return "0.0.1"
		}
	}
	// Если ни одной таблицы нет - это новая база
	return "0.0.0"
}

// Выполнение миграций от текущей версии до целевой
func migrateDB(fromVersion, toVersion string) error {
	logDiagnostic(fmt.Sprintf("Начало миграции: %s → %s", fromVersion, toVersion))
	
	// Определяем последовательность миграций
	versions := []string{"0.0.0", "0.0.1", "0.0.2", "0.0.3", "0.0.4", "0.0.4.1", "0.0.4.2"}
	startIdx := -1
	endIdx := -1
	for i, v := range versions {
		if v == fromVersion {
			startIdx = i
		}
		if v == toVersion {
			endIdx = i
		}
	}
	if startIdx == -1 || endIdx == -1 {
		return fmt.Errorf("неизвестная версия для миграции: %s → %s", fromVersion, toVersion)
	}
	
	// Выполняем миграции по шагам
	for i := startIdx; i < endIdx; i++ {
		current := versions[i]
		next := versions[i+1]
		logPanel(fmt.Sprintf("Выполняется миграция %s → %s...", current, next))
		logDiagnostic(fmt.Sprintf("Миграция %s → %s: начало", current, next))
		
		var err error
		switch next {
		case "0.0.1":
			err = migrate_0_0_0_to_0_0_1()
		case "0.0.2":
			err = migrate_0_0_1_to_0_0_2()
		case "0.0.3":
			err = migrate_0_0_2_to_0_0_3()
		case "0.0.4":
			err = migrate_0_0_3_to_0_0_4()
		case "0.0.4.1":
			err = migrate_0_0_4_to_0_0_4_1()
		case "0.0.4.2":
			err = migrate_0_0_4_1_to_0_0_4_2()
		default:
			return fmt.Errorf("неизвестная версия миграции: %s", next)
		}
		
		if err != nil {
			logDiagnostic(fmt.Sprintf("Миграция %s → %s: ОШИБКА - %v", current, next, err))
			return fmt.Errorf("ошибка миграции %s → %s: %v", current, next, err)
		}
		logDiagnostic(fmt.Sprintf("Миграция %s → %s: УСПЕШНО", current, next))
	}
	
	logPanel(fmt.Sprintf("Миграция завершена: %s → %s", fromVersion, toVersion))
	
	return nil
}

// Миграция 0.0.0 → 0.0.1: создание базовых таблиц
func migrate_0_0_0_to_0_0_1() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS admin_users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL UNIQUE,
password_hash TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS sessions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id INTEGER NOT NULL,
token TEXT NOT NULL UNIQUE,
expires_at TIMESTAMP NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (user_id) REFERENCES admin_users (id)
)`,
		`CREATE TABLE IF NOT EXISTS credentials (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL UNIQUE,
password TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS hosts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
full_name TEXT NOT NULL,
inventory_number TEXT NOT NULL,
seal_numbers TEXT NOT NULL DEFAULT '[]',
monitor_count INTEGER DEFAULT 1,
address_street TEXT NOT NULL,
address_building TEXT NOT NULL,
address_cabinet TEXT NOT NULL,
serial_number TEXT NOT NULL,
replacement_date TIMESTAMP,
replacement_letter TEXT,
old_seal_numbers TEXT DEFAULT '[]',
ip TEXT NOT NULL UNIQUE,
ssh_port INTEGER DEFAULT 22,
enabled BOOLEAN DEFAULT 1,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS scripts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
path TEXT NOT NULL,
parameters_schema TEXT NOT NULL DEFAULT '{}',
description TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS tasks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
description TEXT,
status TEXT NOT NULL DEFAULT 'pending',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS task_hosts (
task_id INTEGER NOT NULL,
host_id INTEGER NOT NULL,
credential_id INTEGER NOT NULL,
PRIMARY KEY (task_id, host_id),
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
FOREIGN KEY (host_id) REFERENCES hosts (id) ON DELETE CASCADE,
FOREIGN KEY (credential_id) REFERENCES credentials (id)
)`,
		`CREATE TABLE IF NOT EXISTS task_scripts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
task_id INTEGER NOT NULL,
script_id INTEGER NOT NULL,
parameters TEXT NOT NULL DEFAULT '{}',
order_index INTEGER NOT NULL,
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
FOREIGN KEY (script_id) REFERENCES scripts (id) ON DELETE CASCADE
)`,
		`CREATE TRIGGER IF NOT EXISTS trg_task_updated_at
AFTER UPDATE ON tasks
BEGIN
UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END`,
	}
	
	for _, query := range queries {
		if err := ExecDB(query); err != nil {
			return fmt.Errorf("ошибка создания таблицы при миграции 0.0.0→0.0.1: %v", err)
		}
	}
	
	// Создаем администратора по умолчанию
	ensureDefaultAdmin()
	
	logPanel("Миграция 0.0.0→0.0.1: созданы базовые таблицы и администратор")
	
	return nil
}

// Миграция 0.0.1 → 0.0.2: добавление таблицы команд bash
func migrate_0_0_1_to_0_0_2() error {
	// Проверяем, существует ли таблица
	if exists, _ := tableExists("task_bash_commands"); exists {
		logDiagnostic("Миграция 0.0.1→0.0.2: таблица task_bash_commands уже существует, пропускаем")
		return nil
	}
	
	query := `CREATE TABLE IF NOT EXISTS task_bash_commands (
id INTEGER PRIMARY KEY AUTOINCREMENT,
task_id INTEGER NOT NULL,
command TEXT NOT NULL,
order_index INTEGER NOT NULL,
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE
)`
	
	if err := ExecDB(query); err != nil {
		return fmt.Errorf("ошибка создания таблицы task_bash_commands: %v", err)
	}
	
	logPanel("Миграция 0.0.1→0.0.2: добавлена таблица команд bash")
	
	return nil
}

// Миграция 0.0.2 → 0.0.3: добавление таблицы настроек платформы и футера
func migrate_0_0_2_to_0_0_3() error {
	// Проверяем, существует ли таблица настроек
	if exists, _ := tableExists("platform_settings"); exists {
		logDiagnostic("Миграция 0.0.2→0.0.3: таблица platform_settings уже существует, пропускаем")
	} else {
		query := `CREATE TABLE IF NOT EXISTS platform_settings (
id INTEGER PRIMARY KEY AUTOINCREMENT,
setting_key TEXT NOT NULL UNIQUE,
setting_value TEXT NOT NULL
)`
		if err := ExecDB(query); err != nil {
			return fmt.Errorf("ошибка создания таблицы platform_settings: %v", err)
		}
		logPanel("Миграция 0.0.2→0.0.3: добавлена таблица настроек платформы")
	}
	
	// Проверяем, существует ли таблица версий (должна быть создана в initializeSchemaVersion)
	if exists, _ := tableExists("schema_version"); !exists {
		query := `CREATE TABLE IF NOT EXISTS schema_version (
version TEXT NOT NULL,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`
		if err := ExecDB(query); err != nil {
			return fmt.Errorf("ошибка создания таблицы версий: %v", err)
		}
		logPanel("Миграция 0.0.2→0.0.3: добавлена таблица отслеживания версий")
	}
	
	// Инициализируем футер, если его нет
	footerKey := "footer_text"
	rows, _ := QueryDB("SELECT setting_value FROM platform_settings WHERE setting_key = '" + footerKey + "'")
	if len(rows) == 0 {
		footerPlaintext := "Powered by andy-key"
		crypto := GetCrypto()
		if crypto != nil {
			encrypted, err := crypto.Encrypt(footerPlaintext)
			if err != nil {
				logDiagnostic("Миграция 0.0.2→0.0.3: ошибка шифрования футера: " + err.Error())
			} else {
				escapedKey := strings.ReplaceAll(footerKey, "'", "''")
				escapedValue := strings.ReplaceAll(encrypted, "'", "''")
				ExecDB("INSERT INTO platform_settings (setting_key, setting_value) VALUES ('" + escapedKey + "', '" + escapedValue + "')")
				logPanel("Миграция 0.0.2→0.0.3: инициализирован зашифрованный футер платформы")
			}
		}
	}
	
	// Обновляем версию в таблице (если она уже существует)
	// Это нужно, потому что в initializeSchemaVersion мы записали версию 0.0.2
	// А после миграции нужно обновить до 0.0.3
	escapedVersion := strings.ReplaceAll("0.0.3", "'", "''")
	ExecDB("UPDATE schema_version SET version = '" + escapedVersion + "', updated_at = CURRENT_TIMESTAMP")
	
	return nil
}

// Миграция 0.0.3 → 0.0.4: добавление справочников и перенос данных
func migrate_0_0_3_to_0_0_4() error {
	logDiagnostic("Миграция 0.0.3→0.0.4: начало создания справочников")
	
	// Создаем таблицу адресов
	addressesQuery := `CREATE TABLE IF NOT EXISTS reference_addresses (
id INTEGER PRIMARY KEY AUTOINCREMENT,
street TEXT NOT NULL,
building TEXT NOT NULL,
cabinet TEXT,
corridor TEXT,
floor INTEGER CHECK (floor BETWEEN 1 AND 5),
service_room TEXT,
description TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`
	
	if err := ExecDB(addressesQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы reference_addresses: %v", err)
	}
	logDiagnostic("Миграция 0.0.3→0.0.4: таблица reference_addresses создана")
	
	// Создаем таблицу подразделений
	departmentsQuery := `CREATE TABLE IF NOT EXISTS reference_departments (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
short_name TEXT,
parent_id INTEGER,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (parent_id) REFERENCES reference_departments (id)
)`
	
	if err := ExecDB(departmentsQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы reference_departments: %v", err)
	}
	logDiagnostic("Миграция 0.0.3→0.0.4: таблица reference_departments создана")
	
	// Создаем таблицу должностей
	positionsQuery := `CREATE TABLE IF NOT EXISTS reference_positions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`
	
	if err := ExecDB(positionsQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы reference_positions: %v", err)
	}
	logDiagnostic("Миграция 0.0.3→0.0.4: таблица reference_positions создана")
	
	// Обновляем версию схемы
	escapedVersion := strings.ReplaceAll("0.0.4", "'", "''")
	ExecDB("UPDATE schema_version SET version = '" + escapedVersion + "', updated_at = CURRENT_TIMESTAMP")
	
	return nil
}

// Миграция 0.0.4 → 0.0.4.1: проверка целостности данных (техническая версия)
func migrate_0_0_4_to_0_0_4_1() error {
	logDiagnostic("Миграция 0.0.4→0.0.4.1: проверка целостности данных")
	// Эта миграция не требует изменений структуры БД,
	// так как таблица platform_settings уже была создана ранее.
	// Просто обновляем номер версии.
	escapedVersion := strings.ReplaceAll("0.0.4.1", "'", "''")
	ExecDB("UPDATE schema_version SET version = '" + escapedVersion + "', updated_at = CURRENT_TIMESTAMP")
	logDiagnostic("Миграция 0.0.4→0.0.4.1: завершена успешно")
	return nil
}

// Миграция 0.0.4.1 → 0.0.4.2: добавление поля description в reference_addresses
func migrate_0_0_4_1_to_0_0_4_2() error {
	logDiagnostic("Миграция 0.0.4.1→0.0.4.2: добавление поля description в таблицу адресов")
	
	// Проверяем наличие колонки description
	descExists, err := columnExists("reference_addresses", "description")
	if err != nil {
		return fmt.Errorf("ошибка проверки наличия колонки description: %v", err)
	}
	
	if !descExists {
		// Добавляем колонку description
		query := `ALTER TABLE reference_addresses ADD COLUMN description TEXT`
		if err := ExecDB(query); err != nil {
			return fmt.Errorf("ошибка добавления колонки description: %v", err)
		}
		logPanel("Миграция 0.0.4.1→0.0.4.2: добавлена колонка description в таблицу reference_addresses")
	} else {
		logDiagnostic("Миграция 0.0.4.1→0.0.4.2: колонка description уже существует, пропускаем")
	}
	
	// Обновляем версию схемы
	escapedVersion := strings.ReplaceAll("0.0.4.2", "'", "''")
	ExecDB("UPDATE schema_version SET version = '" + escapedVersion + "', updated_at = CURRENT_TIMESTAMP")
	logDiagnostic("Миграция 0.0.4.1→0.0.4.2: завершена успешно")
	return nil
}

// Старый код миграции 0.0.3 → 0.0.4 (оставлен для совместимости, но теперь не используется)
/*
UNIQUE(street, building, cabinet, corridor, service_room)
)`
	if _, err := ExecDB(addressesQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы адресов: %v", err)
	}
	logPanel("Миграция 0.0.3→0.0.4: создана таблица адресов")
	
	// Создаем таблицу сотрудников
	employeesQuery := `CREATE TABLE IF NOT EXISTS reference_employees (
id INTEGER PRIMARY KEY AUTOINCREMENT,
full_name TEXT NOT NULL,
short_name TEXT,
phone_city TEXT,
phone_internal TEXT,
address_id INTEGER NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id)
)`
	if err := ExecDB(employeesQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы сотрудников: %v", err)
	}
	logPanel("Миграция 0.0.3→0.0.4: создана таблица сотрудников")
	
	// Создаем таблицу АРМ
	workstationsQuery := `CREATE TABLE IF NOT EXISTS reference_workstations (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
is_vacant BOOLEAN DEFAULT 0,
employee_id INTEGER,
inventory_number TEXT NOT NULL,
seal_numbers TEXT DEFAULT '[]',
monitor_count INTEGER CHECK (monitor_count BETWEEN 1 AND 5),
serial_number TEXT NOT NULL,
replacement_done BOOLEAN DEFAULT 0,
replacement_date TIMESTAMP,
replacement_letter TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id),
FOREIGN KEY (employee_id) REFERENCES reference_employees(id)
)`
	if err := ExecDB(workstationsQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы АРМ: %v", err)
	}
	logPanel("Миграция 0.0.3→0.0.4: создана таблица АРМ")
	
	// Создаем таблицу хостов
	hostsQuery := `CREATE TABLE IF NOT EXISTS reference_hosts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
employee_id INTEGER,
ip TEXT NOT NULL UNIQUE,
ssh_port INTEGER DEFAULT 22,
enabled BOOLEAN DEFAULT 1,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id),
FOREIGN KEY (employee_id) REFERENCES reference_employees(id)
)`
	if err := ExecDB(hostsQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы хостов: %v", err)
	}
	logPanel("Миграция 0.0.3→0.0.4: создана таблица хостов")
	
	// Создаем таблицу сетевого оборудования
	networkEqQuery := `CREATE TABLE IF NOT EXISTS reference_network_equipment (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
category TEXT NOT NULL,
model TEXT NOT NULL,
type TEXT NOT NULL,
port_count INTEGER NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id)
)`
	if err := ExecDB(networkEqQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы сетевого оборудования: %v", err)
	}
	logPanel("Миграция 0.0.3→0.0.4: создана таблица сетевого оборудования")
	
	// Создаем таблицу сетевых МФУ
	mfpsQuery := `CREATE TABLE IF NOT EXISTS reference_network_mfps (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
model TEXT NOT NULL,
ip TEXT NOT NULL UNIQUE,
hostname TEXT,
serial_number TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id)
)`
	if err := ExecDB(mfpsQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы сетевых МФУ: %v", err)
	}
	logPanel("Миграция 0.0.3→0.0.4: создана таблица сетевых МФУ")
	
	// Создаем таблицу IP телефонов
	phonesQuery := `CREATE TABLE IF NOT EXISTS reference_ip_phones (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
employee_full_name TEXT NOT NULL,
ip TEXT NOT NULL UNIQUE,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id)
)`
	if err := ExecDB(phonesQuery); err != nil {
		return fmt.Errorf("ошибка создания таблицы IP телефонов: %v", err)
	}
	logPanel("Миграция 0.0.3→0.0.4: создана таблица IP телефонов")
	
	// Переносим данные из старой таблицы hosts в новые справочники
	// Сначала создаем адреса из существующих данных
	hostsRows, _ := QueryDB("SELECT DISTINCT address_street, address_building, address_cabinet FROM hosts WHERE enabled = 1")
	for _, row := range hostsRows {
		street := row["address_street"]
		building := row["address_building"]
		cabinet := row["address_cabinet"]
		
		// Проверяем, существует ли уже такой адрес
		checkQuery := fmt.Sprintf(
			"SELECT id FROM reference_addresses WHERE street = '%s' AND building = '%s' AND cabinet = '%s'",
			strings.ReplaceAll(street, "'", "''"),
			strings.ReplaceAll(building, "'", "''"),
			strings.ReplaceAll(cabinet, "'", "''"),
		)
		existsRows, _ := QueryDB(checkQuery)
		if len(existsRows) == 0 {
			insertQuery := fmt.Sprintf(
				"INSERT INTO reference_addresses (street, building, cabinet) VALUES ('%s', '%s', '%s')",
				strings.ReplaceAll(street, "'", "''"),
				strings.ReplaceAll(building, "'", "''"),
				strings.ReplaceAll(cabinet, "'", "''"),
			)
			ExecDB(insertQuery)
		}
	}
	
	// Теперь переносим хосты в новую таблицу
	hostsData, _ := QueryDB(`
SELECT h.id, h.full_name, h.ip, h.ssh_port, h.enabled,
a.id as address_id
FROM hosts h
JOIN reference_addresses a
ON h.address_street = a.street
AND h.address_building = a.building
AND h.address_cabinet = a.cabinet
WHERE h.enabled = 1
`)
	for _, host := range hostsData {
		insertHostQuery := fmt.Sprintf(
			"INSERT INTO reference_hosts (address_id, employee_id, ip, ssh_port, enabled) VALUES (%s, NULL, '%s', %s, %s)",
			host["address_id"],
			strings.ReplaceAll(host["ip"], "'", "''"),
			host["ssh_port"],
			host["enabled"],
		)
		ExecDB(insertHostQuery)
	}
	
	// Переименовываем старую таблицу для резервной копии
	ExecDB("ALTER TABLE hosts RENAME TO hosts_legacy")
	logPanel("Миграция 0.0.3→0.0.4: старая таблица hosts переименована в hosts_legacy")
	
	// Обновляем связи в таблицах задач
	ExecDB("UPDATE task_hosts SET host_id = (SELECT rh.id FROM reference_hosts rh JOIN hosts_legacy hl ON rh.ip = hl.ip WHERE hl.id = task_hosts.host_id)")
	logPanel("Миграция 0.0.3→0.0.4: перенос данных завершен")
	
	// Обновляем версию в таблице
	escapedVersion := strings.ReplaceAll("0.0.4", "'", "''")
	ExecDB("UPDATE schema_version SET version = '" + escapedVersion + "', updated_at = CURRENT_TIMESTAMP")
	
	return nil
}
*/

// Сравнение версий в формате "X.Y.Z"
func compareVersions(v1, v2 string) int {
	// Обработка специального случая "0.0.0"
	if v1 == "0.0.0" && v2 == "0.0.0" {
		return 0
	}
	if v1 == "0.0.0" {
		return -1
	}
	if v2 == "0.0.0" {
		return 1
	}
	
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}
	
	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}
		
		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}
	
	return 0
}

// Проверка наличия таблицы в БД
func tableExists(tableName string) (bool, error) {
	rows, err := QueryDB(fmt.Sprintf(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='%s'",
		tableName,
	))
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

// Проверка наличия колонки в таблице
func columnExists(tableName, columnName string) (bool, error) {
	rows, err := QueryDB(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, err
	}
	for _, row := range rows {
		if row["name"] == columnName {
			return true, nil
		}
	}
	return false, nil
}

// Проверка наличия всех необходимых таблиц для указанной версии
func checkRequiredTablesForVersion(version string) []string {
	requiredTables := getRequiredTablesForVersion(version)
	missing := []string{}
	for _, table := range requiredTables {
		exists, err := tableExists(table)
		if err != nil || !exists {
			missing = append(missing, table)
		}
	}
	return missing
}

// Получение SQL-запроса для создания таблицы в указанной версии
func getTableCreationQuery(tableName, version string) string {
	tables := map[string]map[string]string{
		"0.0.1": {
			"admin_users": `CREATE TABLE IF NOT EXISTS admin_users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL UNIQUE,
password_hash TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
			"sessions": `CREATE TABLE IF NOT EXISTS sessions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id INTEGER NOT NULL,
token TEXT NOT NULL UNIQUE,
expires_at TIMESTAMP NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (user_id) REFERENCES admin_users (id)
)`,
			"credentials": `CREATE TABLE IF NOT EXISTS credentials (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL UNIQUE,
password TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
			"hosts": `CREATE TABLE IF NOT EXISTS hosts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
full_name TEXT NOT NULL,
inventory_number TEXT NOT NULL,
seal_numbers TEXT NOT NULL DEFAULT '[]',
monitor_count INTEGER DEFAULT 1,
address_street TEXT NOT NULL,
address_building TEXT NOT NULL,
address_cabinet TEXT NOT NULL,
serial_number TEXT NOT NULL,
replacement_date TIMESTAMP,
replacement_letter TEXT,
old_seal_numbers TEXT DEFAULT '[]',
ip TEXT NOT NULL UNIQUE,
ssh_port INTEGER DEFAULT 22,
enabled BOOLEAN DEFAULT 1,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
			"scripts": `CREATE TABLE IF NOT EXISTS scripts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
path TEXT NOT NULL,
parameters_schema TEXT NOT NULL DEFAULT '{}',
description TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
			"tasks": `CREATE TABLE IF NOT EXISTS tasks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
description TEXT,
status TEXT NOT NULL DEFAULT 'pending',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
			"task_hosts": `CREATE TABLE IF NOT EXISTS task_hosts (
task_id INTEGER NOT NULL,
host_id INTEGER NOT NULL,
credential_id INTEGER NOT NULL,
PRIMARY KEY (task_id, host_id),
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
FOREIGN KEY (host_id) REFERENCES hosts (id) ON DELETE CASCADE,
FOREIGN KEY (credential_id) REFERENCES credentials (id)
)`,
			"task_scripts": `CREATE TABLE IF NOT EXISTS task_scripts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
task_id INTEGER NOT NULL,
script_id INTEGER NOT NULL,
parameters TEXT NOT NULL DEFAULT '{}',
order_index INTEGER NOT NULL,
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
FOREIGN KEY (script_id) REFERENCES scripts (id) ON DELETE CASCADE
)`,
		},
		"0.0.2": {
			"task_bash_commands": `CREATE TABLE IF NOT EXISTS task_bash_commands (
id INTEGER PRIMARY KEY AUTOINCREMENT,
task_id INTEGER NOT NULL,
command TEXT NOT NULL,
order_index INTEGER NOT NULL,
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE
)`,
		},
		"0.0.3": {
			"platform_settings": `CREATE TABLE IF NOT EXISTS platform_settings (
id INTEGER PRIMARY KEY AUTOINCREMENT,
setting_key TEXT NOT NULL UNIQUE,
setting_value TEXT NOT NULL
)`,
			"schema_version": `CREATE TABLE IF NOT EXISTS schema_version (
version TEXT NOT NULL,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		},
		"0.0.4": {
			"reference_addresses": `CREATE TABLE IF NOT EXISTS reference_addresses (
id INTEGER PRIMARY KEY AUTOINCREMENT,
street TEXT NOT NULL,
building TEXT NOT NULL,
cabinet TEXT,
corridor TEXT,
floor INTEGER CHECK (floor BETWEEN 1 AND 5),
service_room TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
UNIQUE(street, building, cabinet, corridor, service_room)
)`,
		},
		"0.0.4.2": {
			"reference_addresses": `CREATE TABLE IF NOT EXISTS reference_addresses (
id INTEGER PRIMARY KEY AUTOINCREMENT,
street TEXT NOT NULL,
building TEXT NOT NULL,
cabinet TEXT,
corridor TEXT,
floor INTEGER CHECK (floor BETWEEN 1 AND 5),
service_room TEXT,
description TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
UNIQUE(street, building, cabinet, corridor, service_room)
)`,
			"reference_employees": `CREATE TABLE IF NOT EXISTS reference_employees (
id INTEGER PRIMARY KEY AUTOINCREMENT,
full_name TEXT NOT NULL,
short_name TEXT,
phone_city TEXT,
phone_internal TEXT,
address_id INTEGER NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses (id)
)`,
			"reference_workstations": `CREATE TABLE IF NOT EXISTS reference_workstations (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
is_vacant BOOLEAN DEFAULT 0,
employee_id INTEGER,
inventory_number TEXT NOT NULL,
seal_numbers TEXT DEFAULT '[]',
monitor_count INTEGER CHECK (monitor_count BETWEEN 1 AND 5),
serial_number TEXT NOT NULL,
replacement_done BOOLEAN DEFAULT 0,
replacement_date TIMESTAMP,
replacement_letter TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses (id),
FOREIGN KEY (employee_id) REFERENCES reference_employees (id)
)`,
			"reference_hosts": `CREATE TABLE IF NOT EXISTS reference_hosts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
employee_id INTEGER,
ip TEXT NOT NULL UNIQUE,
ssh_port INTEGER DEFAULT 22,
enabled BOOLEAN DEFAULT 1,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses (id),
FOREIGN KEY (employee_id) REFERENCES reference_employees (id)
)`,
			"reference_network_equipment": `CREATE TABLE IF NOT EXISTS reference_network_equipment (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
category TEXT NOT NULL,
model TEXT NOT NULL,
type TEXT NOT NULL,
port_count INTEGER NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses (id)
)`,
			"reference_network_mfps": `CREATE TABLE IF NOT EXISTS reference_network_mfps (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
model TEXT NOT NULL,
ip TEXT NOT NULL UNIQUE,
hostname TEXT,
serial_number TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses (id)
)`,
			"reference_ip_phones": `CREATE TABLE IF NOT EXISTS reference_ip_phones (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
employee_full_name TEXT NOT NULL,
ip TEXT NOT NULL UNIQUE,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses (id)
)`,
			"hosts_legacy": `CREATE TABLE IF NOT EXISTS hosts_legacy (
id INTEGER PRIMARY KEY AUTOINCREMENT,
full_name TEXT NOT NULL,
inventory_number TEXT NOT NULL,
seal_numbers TEXT NOT NULL DEFAULT '[]',
monitor_count INTEGER DEFAULT 1,
address_street TEXT NOT NULL,
address_building TEXT NOT NULL,
address_cabinet TEXT NOT NULL,
serial_number TEXT NOT NULL,
replacement_date TIMESTAMP,
replacement_letter TEXT,
old_seal_numbers TEXT DEFAULT '[]',
ip TEXT NOT NULL UNIQUE,
ssh_port INTEGER DEFAULT 22,
enabled BOOLEAN DEFAULT 1,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		},
	}
	
	if ver, ok := tables[version]; ok {
		if query, ok := ver[tableName]; ok {
			return query
		}
	}
	
	return ""
}

// Получение списка обязательных таблиц для указанной версии
func getRequiredTablesForVersion(version string) []string {
	switch version {
	case "0.0.0":
		return []string{}
	case "0.0.1":
		return []string{
			"admin_users", "sessions", "credentials", "hosts", "scripts",
			"tasks", "task_hosts", "task_scripts",
		}
	case "0.0.2":
		return []string{
			"admin_users", "sessions", "credentials", "hosts", "scripts",
			"tasks", "task_hosts", "task_scripts", "task_bash_commands",
		}
	case "0.0.3":
		return []string{
			"admin_users", "sessions", "credentials", "hosts", "scripts",
			"tasks", "task_hosts", "task_scripts", "task_bash_commands",
			"platform_settings", "schema_version",
		}
	case "0.0.4":
		return []string{
			"admin_users", "sessions", "credentials", "scripts",
			"tasks", "task_hosts", "task_scripts", "task_bash_commands",
			"platform_settings", "schema_version",
			"reference_addresses", "reference_employees", "reference_workstations",
			"reference_hosts", "reference_network_equipment", "reference_network_mfps",
			"reference_ip_phones", "hosts_legacy",
		}
	case "0.0.4.2":
		return []string{
			"admin_users", "sessions", "credentials", "scripts",
			"tasks", "task_hosts", "task_scripts", "task_bash_commands",
			"platform_settings", "schema_version",
			"reference_addresses", "reference_employees", "reference_workstations",
			"reference_hosts", "reference_network_equipment", "reference_network_mfps",
			"reference_ip_phones",
		}
	default:
		return []string{}
	}
}

// Получение всех запросов для создания схемы БД (версия 0.0.4)
func getSchemaQueries() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS admin_users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL UNIQUE,
password_hash TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS sessions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id INTEGER NOT NULL,
token TEXT NOT NULL UNIQUE,
expires_at TIMESTAMP NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (user_id) REFERENCES admin_users (id)
)`,
		`CREATE TABLE IF NOT EXISTS credentials (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL UNIQUE,
password TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS scripts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
path TEXT NOT NULL,
parameters_schema TEXT NOT NULL DEFAULT '{}',
description TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS tasks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL UNIQUE,
description TEXT,
status TEXT NOT NULL DEFAULT 'pending',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS task_hosts (
task_id INTEGER NOT NULL,
host_id INTEGER NOT NULL,
credential_id INTEGER NOT NULL,
PRIMARY KEY (task_id, host_id),
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
FOREIGN KEY (host_id) REFERENCES reference_hosts (id) ON DELETE CASCADE,
FOREIGN KEY (credential_id) REFERENCES credentials (id)
)`,
		`CREATE TABLE IF NOT EXISTS task_scripts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
task_id INTEGER NOT NULL,
script_id INTEGER NOT NULL,
parameters TEXT NOT NULL DEFAULT '{}',
order_index INTEGER NOT NULL,
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
FOREIGN KEY (script_id) REFERENCES scripts (id) ON DELETE CASCADE
)`,
		`CREATE TABLE IF NOT EXISTS task_bash_commands (
id INTEGER PRIMARY KEY AUTOINCREMENT,
task_id INTEGER NOT NULL,
command TEXT NOT NULL,
order_index INTEGER NOT NULL,
FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE
)`,
		`CREATE TABLE IF NOT EXISTS platform_settings (
id INTEGER PRIMARY KEY AUTOINCREMENT,
setting_key TEXT NOT NULL UNIQUE,
setting_value TEXT NOT NULL
)`,
		`CREATE TABLE IF NOT EXISTS schema_version (
version TEXT NOT NULL,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE TABLE IF NOT EXISTS reference_addresses (
id INTEGER PRIMARY KEY AUTOINCREMENT,
street TEXT NOT NULL,
building TEXT NOT NULL,
cabinet TEXT,
corridor TEXT,
floor INTEGER CHECK (floor BETWEEN 1 AND 5),
service_room TEXT,
description TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
UNIQUE(street, building, cabinet, corridor, service_room)
)`,
		`CREATE TABLE IF NOT EXISTS reference_employees (
id INTEGER PRIMARY KEY AUTOINCREMENT,
full_name TEXT NOT NULL,
short_name TEXT,
phone_city TEXT,
phone_internal TEXT,
address_id INTEGER NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id)
)`,
		`CREATE TABLE IF NOT EXISTS reference_workstations (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
is_vacant BOOLEAN DEFAULT 0,
employee_id INTEGER,
inventory_number TEXT NOT NULL,
seal_numbers TEXT DEFAULT '[]',
monitor_count INTEGER CHECK (monitor_count BETWEEN 1 AND 5),
serial_number TEXT NOT NULL,
replacement_done BOOLEAN DEFAULT 0,
replacement_date TIMESTAMP,
replacement_letter TEXT,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id),
FOREIGN KEY (employee_id) REFERENCES reference_employees(id)
)`,
		`CREATE TABLE IF NOT EXISTS reference_hosts (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
employee_id INTEGER,
ip TEXT NOT NULL UNIQUE,
ssh_port INTEGER DEFAULT 22,
enabled BOOLEAN DEFAULT 1,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id),
FOREIGN KEY (employee_id) REFERENCES reference_employees(id)
)`,
		`CREATE TABLE IF NOT EXISTS reference_network_equipment (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
category TEXT NOT NULL,
model TEXT NOT NULL,
type TEXT NOT NULL,
port_count INTEGER NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id)
)`,
		`CREATE TABLE IF NOT EXISTS reference_network_mfps (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
model TEXT NOT NULL,
ip TEXT NOT NULL UNIQUE,
hostname TEXT,
serial_number TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id)
)`,
		`CREATE TABLE IF NOT EXISTS reference_ip_phones (
id INTEGER PRIMARY KEY AUTOINCREMENT,
address_id INTEGER NOT NULL,
employee_full_name TEXT NOT NULL,
ip TEXT NOT NULL UNIQUE,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (address_id) REFERENCES reference_addresses(id)
)`,
		`CREATE TRIGGER IF NOT EXISTS trg_task_updated_at
AFTER UPDATE ON tasks
BEGIN
UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END`,
	}
}

// =============== ИНИЦИАЛИЗАЦИЯ БАЗЫ ДАННЫХ ===============
func InitDB() error {
	// Загружаем конфигурацию из файла
	if err := loadConfig(); err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %v", err)
	}
	
	// Проверяем наличие старых файлов БД и переименовываем их при необходимости
	if err := migrateOldDB(); err != nil {
		return fmt.Errorf("ошибка миграции старой БД: %v", err)
	}
	
	var err error
	crypto, err := NewCrypto()
	if err != nil {
		return fmt.Errorf("ошибка инициализации криптографии: %v", err)
	}
	
	if _, err := exec.LookPath("sqlite3"); err != nil {
		return fmt.Errorf("требуется утилита sqlite3: %v", err)
	}
	
	// Инициализация журнала
	if err := initLogger(); err != nil {
		return fmt.Errorf("ошибка инициализации журнала: %v", err)
	}
	
	// Выполняем все запросы создания таблиц (безопасно благодаря IF NOT EXISTS)
	// Это гарантирует, что все таблицы существуют перед проверкой версии
	queries := getSchemaQueries()
	for _, query := range queries {
		cmd := exec.Command("sqlite3", DB_PATH, query)
		if _, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("ошибка инициализации БД: %v", err)
		}
	}
	
	// КРИТИЧЕСКИ ВАЖНО: Проверяем и обновляем версию схемы ДО инициализации футера
	// Это гарантирует, что таблица platform_settings существует
	if err := ensureSchemaVersion(); err != nil {
		logDiagnostic("ОШИБКА ВЕРСИОНИРОВАНИЯ: " + err.Error())
		return err
	}
	
	// Инициализация зашифрованного футера (теперь безопасно вызывать GetCrypto)
	footerKey := "footer_text"
	footerPlaintext := "Powered by andy-key"
	rows, _ := QueryDB("SELECT setting_value FROM platform_settings WHERE setting_key = '" + footerKey + "'")
	if len(rows) == 0 {
		encrypted, err := crypto.Encrypt(footerPlaintext)
		if err != nil {
			logDiagnostic("Ошибка шифрования текста футера: " + err.Error())
		} else {
			escapedKey := strings.ReplaceAll(footerKey, "'", "''")
			escapedValue := strings.ReplaceAll(encrypted, "'", "''")
			ExecDB("INSERT INTO platform_settings (setting_key, setting_value) VALUES ('" + escapedKey + "', '" + escapedValue + "')")
			logPanel("Инициализирован зашифрованный футер платформы")
		}
	}
	
	ensureDefaultAdmin()
	
	db = &DB{crypto: crypto}
	
	// Финальная проверка версии для лога
	versionCheck, _ := QueryDB("SELECT version FROM schema_version ORDER BY updated_at DESC LIMIT 1")
	if len(versionCheck) > 0 {
		logPanel("База данных инициализирована. Версия схемы: " + versionCheck[0]["version"])
	} else {
		logPanel("База данных инициализирована. Версия схемы не определена (требуется перезапуск)")
	}
	
	return nil
}

// =============== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ===============
// Инициализация журнала
func initLogger() error {
	if err := os.MkdirAll(LOGS_DIR, 0755); err != nil {
		return fmt.Errorf("ошибка создания папки логов: %v", err)
	}
	
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	panelLogPath := filepath.Join(LOGS_DIR, fmt.Sprintf("%s-%s-panel.log", LOG_PREFIX, timestamp))
	remoteLogPath := filepath.Join(LOGS_DIR, fmt.Sprintf("%s-%s-remote.log", LOG_PREFIX, timestamp))
	diagnosticLogPath := filepath.Join(LOGS_DIR, fmt.Sprintf("%s-%s-diagnostic.log", LOG_PREFIX, timestamp))
	
	var err error
	panelLog, err = os.OpenFile(panelLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("ошибка создания файла лога panel: %v", err)
	}
	
	remoteLog, err = os.OpenFile(remoteLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("ошибка создания файла лога remote: %v", err)
	}
	
	diagnosticLog, err = os.OpenFile(diagnosticLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("ошибка создания файла лога diagnostic: %v", err)
	}
	
	logPanel("Журнал инициализирован. Файлы: " + filepath.Base(panelLogPath) + ", " + filepath.Base(remoteLogPath) + ", " + filepath.Base(diagnosticLogPath))
	
	return nil
}

// Функции записи в журнал
func logPanel(message string) {
	writeLog(panelLog, "PANEL", message)
}
func logRemote(message string) {
	writeLog(remoteLog, "REMOTE", message)
}
func logDiagnostic(message string) {
	writeLog(diagnosticLog, "DIAGNOSTIC", message)
}
func writeLog(file *os.File, logType, message string) {
	if file == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := "[" + timestamp + "] [" + logType + "] " + message + "\n"
	file.WriteString(line)
}

// Получение последних записей из журнала
func getLogEntries(logType string, limit int) ([]string, error) {
	pattern := filepath.Join(LOGS_DIR, fmt.Sprintf("%s-*-%s.log", LOG_PREFIX, logType))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return []string{"Журнал пуст"}, nil
	}
	
	// Читаем содержимое файла (самый новый)
	content, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		return nil, err
	}
	
	// Разбиваем на строки и берём последние N
	lines := strings.Split(string(content), "\n")
	if len(lines) > limit {
		lines = lines[len(lines)-limit-1 : len(lines)-1]
	}
	
	// Удаляем пустые строки в конце
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	
	return lines, nil
}

// Вспомогательная функция: хеширование пароля с солью
func hashPassword(password string) string {
	salt := "andy-key"
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

// ГАРАНТИРОВАННОЕ создание администратора с проверкой хеша
func ensureDefaultAdmin() {
	username := "admin"
	expectedHash := hashPassword("admin")
	
	rows, err := QueryDB(fmt.Sprintf(
		"SELECT id, password_hash FROM admin_users WHERE username = '%s'",
		strings.ReplaceAll(username, "'", "''"),
	))
	if err != nil || len(rows) == 0 {
		ExecDB(fmt.Sprintf(
			"INSERT OR IGNORE INTO admin_users (username, password_hash) VALUES ('%s', '%s')",
			username, expectedHash,
		))
		logPanel("Создан администратор 'admin'")
	} else {
		userIdStr := rows[0]["id"]
		storedHash := rows[0]["password_hash"]
		if storedHash != expectedHash {
			ExecDB(fmt.Sprintf(
				"UPDATE admin_users SET password_hash = '%s' WHERE id = %s",
				expectedHash, userIdStr,
			))
			logPanel("Исправлен хеш пароля для администратора 'admin'")
		}
	}
}

// ИСПРАВЛЕННАЯ ФУНКЦИЯ: корректный парсинг всех типов данных из SQLite JSON
func QueryDB(query string, args ...string) ([]map[string]string, error) {
	cmdArgs := []string{"-json", DB_PATH}
	for _, arg := range args {
		cmdArgs = append(cmdArgs, arg)
	}
	cmdArgs = append(cmdArgs, query)
	
	cmd := exec.Command("sqlite3", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	
	if len(output) == 0 {
		return []map[string]string{}, nil
	}
	
	// Парсим в []map[string]interface{} для обработки разных типов (числа, строки, булевы)
	var rawResult []map[string]interface{}
	if err := json.Unmarshal(output, &rawResult); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %v. Вывод: %s", err, string(output))
	}
	
	// Конвертируем все значения в строки
	// ИСПРАВЛЕНО: булевы значения конвертируем в "1"/"0" для совместимости с фронтендом
	result := make([]map[string]string, len(rawResult))
	for i, row := range rawResult {
		result[i] = make(map[string]string)
		for key, val := range row {
			switch v := val.(type) {
			case string:
				result[i][key] = v
			case float64:
				// JSON числа парсятся как float64, конвертируем в строку без точки для целых
				if v == float64(int64(v)) {
					result[i][key] = strconv.FormatInt(int64(v), 10)
				} else {
					result[i][key] = strconv.FormatFloat(v, 'f', -1, 64)
				}
			case bool:
				// ИСПРАВЛЕНО: конвертируем булевы в "1"/"0"
				if v {
					result[i][key] = "1"
				} else {
					result[i][key] = "0"
				}
			default:
				result[i][key] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	return result, nil
}

// ИСПРАВЛЕННАЯ ФУНКЦИЯ: возврат к оригинальной логике без ломания компиляции
func ExecDB(query string, args ...string) error {
	cmdArgs := []string{DB_PATH}
	for _, arg := range args {
		cmdArgs = append(cmdArgs, arg)
	}
	cmdArgs = append(cmdArgs, query)
	
	cmd := exec.Command("sqlite3", cmdArgs...)
	_, err := cmd.CombinedOutput()  // ← ИСПОЛЬЗУЕМ ПОДЧЁРКИВАНИЕ _
	
	return err
}

func GetCrypto() *Crypto {
	// Защита от вызова до инициализации
	if db == nil || db.crypto == nil {
		logDiagnostic("CRITICAL ERROR: GetCrypto() вызван до инициализации БД")
		// Попытка восстановления — создание нового криптографического объекта
		crypto, err := NewCrypto()
		if err != nil {
			logDiagnostic("CRITICAL ERROR: Не удалось создать резервный криптографический объект: " + err.Error())
			return nil
		}
		if db == nil {
			db = &DB{crypto: crypto}
		} else {
			db.crypto = crypto
		}
		logDiagnostic("Восстановлен криптографический объект после ошибки инициализации")
	}
	
	return db.crypto
}

// СТРОГАЯ проверка аутентификации с детальной диагностикой
func CheckAdminAuth(username, password string) (bool, int, string) {
	// === ШАГ 1: Проверка существования пользователя по логину ===
	escapedUsername := strings.ReplaceAll(username, "'", "''")
	rows, err := QueryDB(fmt.Sprintf(
		"SELECT id, password_hash FROM admin_users WHERE username = '%s'",
		escapedUsername,
	))
	if err != nil {
		logDiagnostic("DEBUG AUTH [ШАГ 1]: Ошибка запроса к БД при проверке логина '" + username + "': " + err.Error())
		return false, 0, "Ошибка базы данных"
	}
	if len(rows) == 0 {
		logDiagnostic("DEBUG AUTH [ШАГ 1]: Пользователь с логином '" + username + "' НЕ НАЙДЕН в базе данных")
		return false, 0, "Неверный логин"
	}
	
	// === ШАГ 2: Пользователь найден, получаем данные ===
	userId, _ := strconv.Atoi(rows[0]["id"])
	storedHash := rows[0]["password_hash"]
	
	// === ШАГ 3: Хешируем введенный пароль с солью ===
	computedHash := hashPassword(password)
	
	// === ШАГ 4: Сравнение хешей ===
	if computedHash != storedHash {
		logDiagnostic("DEBUG AUTH [ШАГ 4]: Хеши НЕ СОВПАДАЮТ для пользователя '" + username + "'")
		return false, 0, "Неверный пароль"
	}
	
	return true, userId, ""
}

// ИСПРАВЛЕННАЯ ФУНКЦИЯ: создание сессии с детальной проверкой
func CreateSession(userID int) (string, error) {
	token := make([]byte, 16)
	n, err := rand.Read(token)
	if err != nil {
		logDiagnostic("CRITICAL ERROR: Ошибка генерации случайных байт: " + err.Error())
		return "", fmt.Errorf("ошибка генерации токена: %v", err)
	}
	if n != 16 {
		logDiagnostic("CRITICAL ERROR: Сгенерировано только " + strconv.Itoa(n) + " байт вместо 16")
		return "", fmt.Errorf("неполная генерация токена")
	}
	
	// Конвертируем в шестнадцатеричный формат
	tokenHex := hex.EncodeToString(token)
	
	// Форматирование даты истечения в UTC
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	expiresAtStr := expiresAt.Format("2006-01-02 15:04:05")
	
	// Экранирование токена
	escapedToken := strings.ReplaceAll(tokenHex, "'", "''")
	
	// Формируем запрос
	query := "INSERT INTO sessions (user_id, token, expires_at) VALUES (" + strconv.Itoa(userID) + ", '" + escapedToken + "', '" + expiresAtStr + "')"
	
	// Выполняем запрос
	cmdArgs := []string{DB_PATH, query}
	cmd := exec.Command("sqlite3", cmdArgs...)
	_, err = cmd.CombinedOutput()  // ← ИСПОЛЬЗУЕМ ПОДЧЁРКИВАНИЕ _
	
	if err != nil {
		logDiagnostic("CRITICAL ERROR: Ошибка вставки сессии в БД: " + err.Error())
		// Проверяем существование таблицы
		_, tableErr := QueryDB("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'")
		if tableErr != nil {
			logDiagnostic("CRITICAL ERROR: Таблица 'sessions' не существует!")
			// Попробуем создать таблицу
			ExecDB(`CREATE TABLE IF NOT EXISTS sessions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id INTEGER NOT NULL,
token TEXT NOT NULL UNIQUE,
expires_at TIMESTAMP NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (user_id) REFERENCES admin_users (id)
)`)
		}
		
		return "", fmt.Errorf("ошибка создания сессии: %v", err)
	}
	
	return tokenHex, nil
}

// ИСПРАВЛЕННАЯ ФУНКЦИЯ: проверка сессии с детальной диагностикой
func CheckSession(token string) bool {
	// Экранируем токен
	escapedToken := strings.ReplaceAll(token, "'", "''")
	
	// Используем правильный формат для сравнения дат
	// SQLite использует 'now' для текущего времени в формате UTC
	rows, err := QueryDB("SELECT id FROM sessions WHERE token = '" + escapedToken + "' AND expires_at > DATETIME('now')")
	if err != nil {
		logDiagnostic("CRITICAL ERROR: Ошибка проверки сессии: " + err.Error())
		return false
	}
	if len(rows) == 0 {
		// Удаляем просроченные сессии (очистка)
		ExecDB("DELETE FROM sessions WHERE expires_at < DATETIME('now')")
		return false
	}
	
	return true
}