# SSH Orchestrator

Платформа автоматизации для управления удалёнными хостами через SSH.

## Версия
0.0.4.1 (Стабильная)

## Требования

- Go 1.19 или выше
- SQLite3
- Linux (Windows не поддерживается)


## Установка


### Быстрая установка

```bash
# Клонировать репозиторий
git clone <repository-url>
cd andy-key

# Сборка
env CGO_ENABLED=0 go build -o andy-key

# Запуск
./andy-key 9000
(или любой другой свободный порт)

Вот инструкция для README.md о зависимостях и деплое проекта **andy-key**:

```markdown
## Требования

Для работы приложения необходимы следующие зависимости:

### 1. Go 1.19 или выше
```bash
# Проверка версии Go
go version

# Установка (если не установлена)
# Подробнее: https://golang.org/doc/install
```

### 2. SQLite3 CLI
```bash
# Ubuntu/Debian
sudo apt-get install sqlite3

# CentOS/RHEL
sudo yum install sqlite3

# macOS (Homebrew)
brew install sqlite3

# Windows
# Скачать с https://www.sqlite.org/download.html
```

### 3. sshpass (для выполнения SSH команд)
```bash
# Ubuntu/Debian
sudo apt-get install sshpass

# CentOS/RHEL
sudo yum install sshpass

# macOS (Homebrew)
brew install hudochenkov/sshpass/sshpass

# Windows
# Скачать с https://sourceforge.net/projects/sshpass/
```

## Быстрый старт

### 1. Клонирование репозитория
```bash
git clone https://github.com/yourusername/andy-key.git
cd andy-key
```

### 2. Сборка проекта
```bash
env CGO_ENABLED=0 go build -o andy-key
```

### 3. Запуск приложения
```bash
# Запуск с портом по умолчанию (9000)
./andy-key

# Запуск с указанием порта
./andy-key 8080

# Запуск с указанием порта через переменную окружения
PORT=8080 ./andy-key
```

## Первоначальная настройка

После первого запуска:
1. **База данных** `andy-key.db` будет создана автоматически
2. **Мастер-ключ** `.andy-key.key` будет сгенерирован автоматически
3. **Папка логов** `logs/` будет создана автоматически
4. **Администратор** по умолчанию: логин `admin`, пароль `admin` (обязательно смените!)

## Структура файлов

```
andy-key/
├── andy-key              # Бинарный файл приложения
├── andy-key.db           # База данных SQLite
├── .andy-key.key         # Мастер-ключ для шифрования (права 600)
├── logs/                # Папка с логами
│   ├── andy-key-YYYY-MM-DD-HH-MM-SS-panel.log
│   ├── andy-key-YYYY-MM-DD-HH-MM-SS-remote.log
│   └── andy-key-YYYY-MM-DD-HH-MM-SS-diagnostic.log
└── frontend/            # Пакет с веб-интерфейсом
```

## Настройка безопасности

### 1. Смена пароля администратора
После первого входа обязательно смените пароль администратора в разделе настроек.

### 2. Защита мастер-ключа
```bash
# Установка прав 600 на мастер-ключ
chmod 600 .andy-key.key
```

### 3. Бэкап базы данных
```bash
# Регулярное резервное копирование
cp andy-key.db andy-key.db.backup.$(date +%Y%m%d_%H%M%S)
```

## Запуск в production

### 1. Системный сервис (systemd)

Создайте файл `/etc/systemd/system/andy-key.service`:

```ini
[Unit]
Description=andy-key - SSH Administration Platform
After=network.target

[Service]
Type=simple
User=andy-key
Group=andy-key
WorkingDirectory=/opt/andy-key
ExecStart=/opt/andy-key/andy-key 9000
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=andy-key

[Install]
WantedBy=multi-user.target
```

Затем выполните:

```bash
sudo systemctl daemon-reload
sudo systemctl enable andy-key
sudo systemctl start andy-key
sudo systemctl status andy-key
```

### 2. Настройка reverse proxy (Nginx)

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 3. Настройка HTTPS (Let's Encrypt)

```bash
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## Обновление

```bash
# Остановка сервиса
sudo systemctl stop andy-key

# Резервное копирование
cp andy-key.db andy-key.db.backup.$(date +%Y%m%d_%H%M%S)
cp .andy-key.key .andy-key.key.backup

# Обновление кода
git pull origin andy-key-0.0.4.2

# Пересборка
env CGO_ENABLED=0 go build -o andy-key

# Запуск сервиса
sudo systemctl start andy-key
```

## Устранение неполадок

### Ошибка: "sqlite3: command not found"
```bash
# Установите SQLite3 CLI (см. раздел "Требования")
```

### Ошибка: "sshpass: command not found"
```bash
# Установите sshpass (см. раздел "Требования")
```

### Ошибка: "port already in use"
```bash
# Проверьте, какой процесс использует порт
lsof -i :9000

# Или запустите на другом порту
./andy-key 8080
```

### Ошибка: "permission denied" на .andy-key.key
```bash
# Установите правильные права
chmod 600 .andy-key.key
```

## Лицензия

Проект разработан под лицензией MIT.

## Поддержка

Для вопросов и предложений создавайте issue в репозитории.
```

---

## Краткая версия для быстрого копирования:

```bash
# Установка зависимостей
sudo apt-get install sqlite3 sshpass

# Сборка
env CGO_ENABLED=0 go build -o andy-key

# Запуск
./andy-key

# Администратор по умолчанию: admin/admin (обязательно смените!)
```

---

## Ключевые зависимости:

| Зависимость | Назначение | Установка |
|-------------|------------|-----------|
| **Go 1.19+** | Компилятор и рантайм | `sudo apt-get install golang` |
| **sqlite3** | База данных (CLI) | `sudo apt-get install sqlite3` |
| **sshpass** | Выполнение SSH команд | `sudo apt-get install sshpass` |

**Важно:** Проект не использует сторонние Go-библиотеки — только встроенные пакеты стандартной библиотеки Go.