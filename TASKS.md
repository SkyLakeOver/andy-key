# TASKS.md — единый реестр задач andy-key

## 1. Стоящие правила проекта
- Правило ключа: в справочниках уникален только id (PRIMARY KEY); бизнес-поля
  (улица/дом/кабинет, инвентарный номер, ФИО) могут дублироваться.
  ИСКЛЮЧЕНИЕ (решение владельца): технические уникальности СОХРАНЯЮТСЯ —
  hosts.ip, reference_network_mfps.ip, reference_ip_phones.ip, credentials.username.
- Сортировка всех списков справочников: ORDER BY id.
- Строка "NULL" никогда не передаётся аргументом в execSafe/QueryDB — только голый
  SQL-литерал в ветке запроса; bindArgs конвертирует пустые строки в NULL.
- .gitignore не изменяется ни в одной сессии; артефакты (*.db, *.key, cookies.txt,
  logs/, бинарники) живут вне репозитория.
- .gitignore ЗАМОРОЖЕН: содержимое каноническое с коммита канонизации; любая сессия при
  непустом diff HEAD по .gitignore восстанавливает файл из HEAD и рапортует; изменение
  содержимого возможно ТОЛЬКО отдельным коммитом по явному решению владельца.
- Приёмка: curl + sqlite3 в workspace кодера достаточна для коммита; браузерные
  проверки выполняет пользователь ПОСЛЕ merge на своём ПК.
- Кодер не выполняет git push; публикацию делает пользователь кнопкой.

## 2. Закрыто
- Фаза 1 «Адреса»: SSR + Vue-острова (PR #22).
- Фаза 2 «АРМ»: SSR + Vue-острова (PR #24).
- Фундамент: bindArgs; initTemplates; case nil в QueryDB; document.addEventListener
  в islands.js; универсальный парсинг data-initial-data (PR #25).
- RECOMMENDATIONS Critical: №2 RETURNING id, №3 quoted secrets, №4 PRAGMA
  foreign_keys (PR #28). №1 (бизнес-UNIQUE адресов): NULL-семантика bindArgs в PR #28,
  снятие самого UNIQUE-индекса — в очереди (Унификация-1).
- Аудит Alpine: «Адреса» и «АРМ» мигрированы; сломанность остальных Alpine-разделов
  подтверждена сопоставлением app.js с API (employees: fio/position vs full_name;
  tasks: viewTask/startTask/stopTask не определены).

## 3. Активная очередь сессий
1. Унификация-1: миграция схемы 0.0.5→0.0.6 (reference_addresses без бизнес-UNIQUE
   пересозданием таблицы), удаление isAddressDuplicate и normalizeUniqueEmptyStringsToNULL,
   ORDER BY id во всех case contentSectionHandler.
2. Порт-employees: SSR + Vue-острова по образцу адресов (контракт full_name, short_name,
   phone_city, phone_internal, address_id); удалить employeesPage из app.js.
3. Порт-hosts: контракт ip, ssh_port, enabled, address_id, employee_id; UNIQUE ip сохраняется.
4. Порт-трио: network-equipment, network-mfps, ip-phones; UNIQUE ip сохраняется;
   денормализация employee_full_name в ip-phones — см. беклог.
5. Демонтаж Alpine Фазы A-D (раздел 5).

## 4. Беклог
- Сортировка по клику на столбец (серверный ORDER BY + параметр запроса).
- modernc.org/sqlite вместо sqlite3 CLI: 0 внешних зависимостей, транзакции, нативные
   параметризованные запросы (закрывает RECOMMENDATIONS №5 и №9); на Linux и Windows
   бинарник самодостаточен (CGO_ENABLED=0).
- Security-десерт (защищённый контур, приоритет низкий): bcrypt/argon2id вместо
   SHA-256 со статической солью; ensureDefaultAdmin не трогает существующий пароль;
   Secure-cookie включать только при TLS (иначе ломает вход по HTTP в контуре).
- ip_phones: заменить денормализованное employee_full_name на employee_id (FK).
- Ротация логов; единый формат ошибок {"error":{"code","message"}}.

## 5. Демонтаж Alpine (актуальная часть аудита)
- Оставшиеся Alpine-разделы: index.html (app()), login.html (loginForm + инлайн-дубль),
  tasks.html, employees.html, hosts.html, network_equipment.html, network_mfps.html,
  ip_phones.html; app.js: 9 компонентов Alpine.data.
- Фаза A: удалить frontend/static/js/vue.global.js (dev-копия 596 КБ, не подключена);
  убрать инлайн-дубль loginForm из login.html (источник истины выбрать при порте).
- Фаза B: каркас index.html → мини-остров AppShell или чистый HTMX; login.html → остров.
- Фаза C: выполняется портами справочников (очередь раздела 3); после каждого порта
  удалять соответствующий Alpine.data из app.js и директивы из шаблона.
- Фаза D: удалить alpine.min.js, остаток app.js, теги script из index.html/login.html,
  CSS [x-cloak]; удалить мёртвые components/input.html, select.html, crud_buttons.html
  (вызовов {{template}} нет, initTemplates парсит их молча).
- Чек-лист приёмки демонтажа: grep -rin "alpine|x-data|x-show|x-model|x-for|
  x-text|x-cloak|@click|:class|:disabled|:key" frontend/templates frontend/static/js → пусто;
  ls alpine.min.js app.js vue.global.js → файлов нет; grep "alpine|app.js" index.html
  login.html → пусто; go build + запуск; smoke всех разделов /api/content/*; консоль
  без ошибок. Ожидаемый эффект: embedded JS ~897 КБ → ~235 КБ.

## 6. Известные особенности среды и процесса
- raw.githubusercontent.com кэширует по имени ветки: код проверять по SHA коммита.
- Workspace кодера изолирован: пользователь не имеет к нему доступа; ручные тесты
  только пост-merge; платформа создаёт коммит при публикации кнопкой.
- HEAD workspace кодера может отличаться от HEAD ветки (grafted-история) — фиксировать
  фактический HEAD в каждом отчёте.
- Квота инструментов сессии ограничена: при исчерпании — честный статус «не получено»,
  коммит только после зелёной приёмки.
