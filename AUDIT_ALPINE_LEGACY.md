# АУДИТ: Легаси Alpine.js в проекте andy-key

**Дата аудита:** 03.10.2026
**Ревизия:** `1d38919` (ветка/тег `AK-2.2.0`, merge PR #22 «vue-islands-addresses-refactoring»)
**Целевая аудитория:** ИИ-агенты, выполняющие комплексный анализ и рефакторинг кода.
**Статус сборки на момент аудита:** `go build ./...` — успешно (exit 0).

---

## 1. Резюме (TL;DR)

Проект мигрирует с **Alpine.js** на **Vue 3 (островная архитектура) + HTMX**. Мигрирован только
один раздел — **«Адреса»** (Фаза 1, AK-2.2.0). Остальные SPA-подобные страницы фронтенда всё ещё
размечены под Alpine-директивы (`x-data`, `x-model`, `x-for`, `x-show`, `@click`) и завязаны на
`app.js` (9 компонентов `Alpine.data`).

**Ключевой вывод: Alpine легаси не просто «мертв» — он функционально сломан.** Обнаружены
несоответствия имён полей между API Go-бэкенда и Alpine-компонентами, а также вызовы методов,
которые в `app.js` не определены. Удаление Alpine без переноса этих страниц на Vue/HMTX ничего
не сломает дополнительно, но и ничего не починит — эти разделы уже нерабочие на клиенте.

---

## 2. Актуальная архитектура фронтенда

| Слой | Технология | Где подключается | Статус |
|---|---|---|---|
| Загрузка контента разделов | HTMX (`hx-get="/api/content/{section}"`) | `index.html`, все шаблоны | Активен |
| Интерактивные формы (острова) | Vue 3 (`vue.global.prod.js` + `islands.js`) | `index.html` | Активен (только раздел «Адреса») |
| Реактивность старых страниц | Alpine.js (`alpine.min.js` + `app.js`) | `index.html`, `login.html` | **Легаси, частично сломан** |
| Дубликат Vue | `vue.global.js` (dev, 596 КБ) | Нигде не подключается | Мёртвый вес в embed |

Маршруты сервера: `main.go:setupRoutes()` (~строка 2493+), island-маршруты регистрируются в
`addresses_islands.go:registerAddressIslandRoutes()`.

---

## 3. Инвентаризация файлов Alpine-легаси

### 3.1 Библиотека и компонентный слой

| Файл | Размер | Назначение | Действие |
|---|---|---|---|
| `frontend/static/js/alpine.min.js` | 43 838 Б | Alpine.js 3.x (локальная копия) | **Удалить** после миграции всех страниц |
| `frontend/static/js/app.js` | 21 758 Б | 9 компонентов `Alpine.data` | **Переписать/удалить** (см. §5) |
| `frontend/static/js/vue.global.js` | 596 415 Б | Vue dev-сборка, не используется | **Удалить немедленно** (безопасно, ссылок нет) |

### 3.2 Шаблоны с Alpine-директивами

| Шаблон | Alpine-компонент (`x-data`) | Мигрирован на Vue-острова? |
|---|---|---|
| `templates/index.html` | `app()` — глобальный state (activeTab, currentUser, logout) | Нет (обёртка приложения) |
| `templates/login.html` | `loginForm()` + inline-скрипт `Alpine.data('loginForm')` (дубль!) | Нет |
| `templates/tasks.html` | `tasksPage()` | Нет |
| `templates/employees.html` | `employeesPage()` | Нет |
| `templates/workstations.html` | `workstationsPage()` | Нет |
| `templates/hosts.html` | `hostsPage()` | Нет |
| `templates/network_equipment.html` | `networkEquipmentPage()` | Нет |
| `templates/network_mfps.html` | `networkMfpsPage()` | Нет |
| `templates/ip_phones.html` | `ipPhonesPage()` | Нет |
| `templates/components/input.html` | inline `x-data="{ error: '{{.Error}}' }"` | Нет (компонент, возможно неиспользуемый — проверить использование) |
| `templates/addresses.html` | — (использует `data-island="address-form"`) | **Да** (образец для миграции) |
| `templates/components/address_edit_form.html` | — (`data-island="address-edit"`) | **Да** |

### 3.3 Точки подключения в HTML

- `templates/index.html` строки ~17–21: загрузка `app.js` ДО `alpine.min.js` (defer).
- `templates/login.html` строки ~145–150: загрузка `alpine.min.js` + **инлайн-дубль** компонента
  `loginForm` (конфликтует с определением в `app.js`; в `login.html` есть `footerText`/`loading`,
  которых нет в версии из `app.js`).

### 3.4 Серверный код (Go)

**Прямого использования Alpine в Go-коде нет.** `grep alpine|app.js|islands` по `*.go`:
- единственное упоминание — комментарий `main.go:2494` о регистрации island-маршрутов (Vue).
- `frontend/embed.go` эмбедит `static/js/*.js` целиком → удаление файлов автоматически уменьшит бинарь.
- `initTemplates()` парсит ВСЕ `templates/*.html` и `templates/components/*.html` независимо от
  использования → битые/легаси шаблоны не падают при старте, ошибки возможны только при рендере.

---

## 4. Доказательство «мертвости/сломанности» Alpine-страниц

Проверено сопоставлением JSON-полей handlers в `main.go` с полями, которые читает `app.js`:

1. **employees**: API (`apiReferenceEmployeesHandler`, main.go:757+) отдаёт `full_name`,
   `short_name`, `phone_city`, `phone_internal`. `app.js` фильтрует по `emp.fio`,
   `emp.position`, `emp.department` → **TypeError при поиске**, форма пишет `{fio, position, ...}`,
   которых бэкенд не принимает (**битый CRUD**).
2. **hosts**: шаблоны/`app.js` используют `hostname`, `ip_address`, `mac_address`, `os_type`;
   актуальная схема БД/API — `ip`, `ssh_port`, `enabled` (`hosts.html` таблица рендерит
   `host.ip`, `host.ssh_port`) → **рассинхрон контрактов**.
3. **tasks**: `tasks.html` вызывает `viewTask()`, `startTask()`, `stopTask()`, `deleteTask()`,
   `closeModal()` — grep по `app.js`: ни один из этих методов НЕ определён (есть только
   `loadTasks/loadScripts/openAddModal/closeAddModal/submitTask/filteredTasks`).
   **Кнопки просмотра/запуска/остановки/удаления и закрытия модалки не работают.**
4. **workstations/equipment/mfps/ip-phones**: аналогично — поля форм (`model`, `status`, `name`,
   `extension`...) не совпадают с полями новых SSR-шаблонов (`inventory_number`, `seal_numbers`,
   `category_label`, `employee_full_name`...). Эти страницы дублируют функциональность, которая
   должна уходить в серверный рендер + острова.
5. **Раздел addresses в app.js полностью удалён** (пустой комментарий
   `// Компонент для страницы адресов`) — подтверждение направления миграции.

Вывод для агентов: **не пытаться «починить» app.js** — он обречён; стратегия — портовая миграция
разделов на образец «Адресов» (SSR-таблица через `/api/content/{section}` + Vue-острова для форм).

---

## 5. План демонтажа Alpine (рекомендация для исполняющих агентов)

### Фаза A — безопасные удаления (не требуют изменений логики)
1. Удалить `frontend/static/js/vue.global.js` (dev-копия, 596 КБ, ни на одной странице не подключена).
2. Удалить инлайн-дубль `loginForm` из `login.html` ИЛИ из `app.js` (выбрать источник истины;
   версия в `login.html` богаче: `loading`, `footerText`).

### Фаза B — миграция каркаса приложения
3. `index.html`: заменить `x-data="app()"`, `@click.prevent`, `:class`, `x-text` на мини-остров
   `AppShell` (Vue) или чистый HTMX (`hx-get` по ссылкам навигации уже есть — `@click` нужен
   только для подсветки активной вкладки и `pageTitle`). `logout()` — маленький fetch, можно оставить vanilla JS.
4. `login.html`: переписать на Vue-остров `login-form` (или vanilla form POST).

### Фаза C — миграция справочников (по образцу раздела «Адреса»)
Для каждого раздела (`employees`, `workstations`, `hosts`, `network-equipment`, `network-mfps`,
`ip-phones`, `tasks`):
5. Таблица уже рендерится сервером через `/api/content/{section}` (SSR-шаблоны содержат и
   Alpine-разметку, и статические `{{range}}`-блоки — привести к единому SSR-виду как в `addresses.html`).
6. Модальные формы перевести на `data-island` + эндпоинты вида
   `/api/reference/{entity}/island` (см. `addresses_islands.go` — эталон: параметризованный SQL,
   HTML-фрагмент edit-form с экранированным `data-initial-data`, проверочная ссылка целостности при DELETE).
7. После перевода раздела: удалить его `Alpine.data(...)` из `app.js` и директивы из шаблона.

### Фаза D — финальный снос
8. Удалить `alpine.min.js`, остаток `app.js`, `<script src="/static/js/app.js">` и
   `alpine.min.js` из `index.html`/`login.html`, CSS-правило `[x-cloak]` из `login.html`.
9. Проверить `components/input.html`, `select.html`, `crud_buttons.html` на использование через
   `{{template ...}}` — **проверено: ни в одном шаблоне вызовов `{{template` нет**, а `initTemplates()`
   парсит их молча → это мёртвый код, удалять вместе с Alpine-легаси (учесть: `embed.go`
   эмбедит `templates/components/*.html` glob'ом).

### Риски
- `embed.go` использует glob `static/js/*.js` — новые файлы добавлять безопасно, удаление — тоже.
- `initTemplates()` падает на синтаксической ошибке ЛЮБОГО шаблона при старте — валидировать
  правки шаблонов сборкой/запуском.
- Конфликт `htmx:beforeSwap/unmountIslands` с Alpine-обработчиками внутри заменяемого DOM —
  при сосуществовании двух фреймворков возможны двойные подписки на события (уже сейчас
  `tasks.html` и т.п. живут внутри Alpine-root, а их контент меняет HTMX).

---

## 6. Смежные находки вне Alpine (важно для комплексного анализа)

> Не является частью задачи «снести Alpine», но обнаружено при аудите; вынесено отдельными
> пунктами, чтобы другие агенты не потеряли.

1. **Секреты в репозитории:** `cookies.txt` содержит живой `session_token`; `data.db` (131 КБ,
   включает таблицы `admin_users`, `credentials`, `sessions`) закоммичен в git. `.gitignore`
   не покрывает `*.db`/`cookies.txt`. Рекомендовать `git rm --cached` + ротацию сессий.
2. **Шаблонизатор vs данные:** `contentSectionHandler` при отсутствии шаблона рендерит заглушку —
   разделы `credentials`, `scripts`, `bash-constructor`, `logs` всегда показывают
   «Контент в разработке» (соответствующих `.html` в `templates/` нет). Именно поэтому кнопки
   навигации этих разделов в `index.html` ведут на Alpine-несуществующие компоненты.
3. **SSH через `exec.Command("sshpass", ...)`** (`ssh.go`) с `StrictHostKeyChecking=no` — пароль
   утекает в argv процесса; пароли хранятся AES-шифрованными (`crypto.go`, ключ `.andy-key.key`
   рядом с бинарём). Отметить как технический долг безопасности.
4. **README противоречив:** заголовок «SSH Orchestrator 0.0.4.1», CONTEXT.md описывает другой
   проект (Inventory System на HTMX). `go.mod`: `module andy-key`, `go 1.19` (CONTEXT требует 1.23+).
   Документацию синхронизировать после демонтажа.

---

## 7. Чек-лист проверки полноты демонтажа (для приёмки)

```bash
# 1. Ни одного упоминания Alpine вне документации:
grep -rin "alpine\|x-data\|x-show\|x-model\|x-for=\|x-text\|x-cloak\|@click\|:class\|:disabled\|:key" \
  frontend/templates frontend/static/js --include="*.html" --include="*.js" | grep -v htmx.min.js
# 2. Файлы удалены:
ls frontend/static/js/alpine.min.js frontend/static/js/app.js frontend/static/js/vue.global.js 2>&1
# 3. index.html/login.html не ссылаются на alpine/app.js:
grep -n "alpine\|app.js" frontend/templates/index.html frontend/templates/login.html
# 4. Сборка и запуск:
go build ./... && ./andy-key 9000
# 5. Ручной smoke: все пункты меню /api/content/* открываются, CRUD «Адресов» работает,
#    вход/выход работают, консоль браузера без ошибок.
```

Ожидаемый результат: размер embedded JS сократится с ~897 КБ до ~235 КБ
(`htmx.min.js` 47.8K + `vue.global.prod.js` 168.3K + `islands.js` 18.7K).
