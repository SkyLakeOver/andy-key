package frontend

// uiHTML содержит только HTML-разметку с открывающим <script>
var uiHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>andy-key</title>
<style>
:root {
--primary: #E95420;
--primary-dark: #C3441C;
--secondary: #555;
--success: #3F8C44;
--danger: #D94F4F;
--warning: #F5A623;
--dark: #2D2D2D;
--light: #F5F5F5;
--gray: #666;
--border: #DDD;
--shadow: 0 2px 4px rgba(0,0,0,0.1);
--transition: all 0.3s ease;
}
* {
margin: 0;
padding: 0;
box-sizing: border-box;
}
body {
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
background-color: #f8f9fa;
color: #333;
line-height: 1.5;
min-height: 100vh;
display: flex;
flex-direction: column;
}
.app-container {
display: flex;
flex: 1;
min-height: calc(100vh - 40px);
}
.sidebar {
width: 260px;
background-color: var(--dark);
color: white;
transition: var(--transition);
z-index: 1000;
box-shadow: var(--shadow);
overflow-y: auto;
max-height: 100vh;
}
.sidebar-header {
padding: 20px;
text-align: center;
border-bottom: 1px solid rgba(255,255,255,0.1);
background-color: rgba(0,0,0,0.2);
}
.sidebar-header h1 {
font-size: 1.4rem;
font-weight: 600;
color: white;
}
.sidebar-header .subtitle {
font-size: 0.85rem;
color: #aaa;
margin-top: 4px;
}
.nav-menu {
list-style: none;
padding: 0;
}
.nav-section {
margin-top: 15px;
border-top: 1px solid rgba(255,255,255,0.08);
padding-top: 15px;
}
.nav-section-title {
padding: 8px 20px;
font-size: 0.85rem;
font-weight: 600;
text-transform: uppercase;
letter-spacing: 0.5px;
color: rgba(255,255,255,0.6);
}
.nav-item {
border-bottom: 1px solid rgba(255,255,255,0.08);
}
.nav-link {
display: block;
padding: 12px 20px 12px 40px;
color: rgba(255,255,255,0.85);
text-decoration: none;
transition: var(--transition);
font-weight: 500;
font-size: 0.95rem;
}
.nav-link:hover, .nav-link.active {
background-color: rgba(255,255,255,0.1);
color: white;
padding-left: 45px;
}
.nav-link:hover::before, .nav-link.active::before {
content: "•";
position: absolute;
left: 20px;
color: var(--primary);
font-size: 1.2rem;
}
.main-content {
flex: 1;
overflow-x: hidden;
padding-bottom: 50px;
}
.header {
background-color: white;
box-shadow: var(--shadow);
padding: 15px 30px;
display: flex;
justify-content: space-between;
align-items: center;
z-index: 999;
}
.header h1 {
font-size: 1.5rem;
font-weight: 600;
color: var(--dark);
}
.content-container {
padding: 25px;
max-width: 1600px;
margin: 0 auto;
}
.card {
background-color: white;
border-radius: 6px;
box-shadow: var(--shadow);
margin-bottom: 25px;
overflow: hidden;
}
.card-header {
padding: 18px 25px;
background-color: #f8f9fa;
border-bottom: 1px solid var(--border);
display: flex;
justify-content: space-between;
align-items: center;
flex-direction: column;
align-items: flex-start;
}
.card-header h2 {
font-size: 1.3rem;
font-weight: 600;
color: var(--dark);
margin-bottom: 10px;
}
.card-header > div {
width: 100%;
}
.btn {
display: inline-block;
padding: 8px 16px;
background-color: var(--primary);
color: white;
border: none;
border-radius: 4px;
cursor: pointer;
font-weight: 500;
transition: var(--transition);
text-decoration: none;
font-size: 0.875rem;
}
.btn:hover {
background-color: var(--primary-dark);
}
.btn-success {
background-color: var(--success);
}
.btn-success:hover {
background-color: #357a39;
}
.btn-danger {
background-color: var(--danger);
}
.btn-danger:hover {
background-color: #c93030;
}
.btn-sm {
padding: 6px 12px;
font-size: 0.875rem;
}
.card-body {
padding: 25px;
}
.table-container {
overflow-x: auto;
}
table {
width: 100%;
border-collapse: collapse;
min-width: 800px;
}
table th, table td {
padding: 12px 14px;
text-align: left;
border-bottom: 1px solid var(--border);
font-size: 0.9rem;
}
table th {
background-color: #f8f9fa;
font-weight: 600;
color: var(--dark);
text-transform: uppercase;
font-size: 0.85rem;
position: relative;
cursor: pointer;
}
table th:hover {
background-color: #e9ecef;
}
table tr:hover {
background-color: #f9f9f9;
}
.form-group {
margin-bottom: 20px;
}
.form-group label {
display: block;
margin-bottom: 8px;
font-weight: 500;
color: var(--dark);
font-size: 0.95rem;
}
.form-group label::after {
content: " *";
color: var(--danger);
font-weight: normal;
}
.form-group label.optional::after {
content: "";
}
.form-control {
width: 100%;
padding: 10px 15px;
border: 1px solid var(--border);
border-radius: 4px;
font-size: 1rem;
transition: var(--transition);
}
.form-control:focus {
border-color: var(--primary);
outline: none;
box-shadow: 0 0 0 3px rgba(233, 84, 32, 0.15);
}
.form-text {
font-size: 0.875rem;
color: var(--gray);
margin-top: 6px;
}
.alert {
padding: 15px;
border-radius: 4px;
margin-bottom: 20px;
font-weight: 500;
font-size: 0.95rem;
}
.alert-success {
background-color: #d4edda;
color: #155724;
border: 1px solid #c3e6cb;
}
.alert-danger {
background-color: #f8d7da;
color: #721c24;
border: 1px solid #f5c6cb;
}
.empty-state {
text-align: center;
padding: 40px 20px;
color: var(--gray);
font-size: 1.1rem;
}
.tabs {
display: flex;
margin-bottom: 20px;
border-bottom: 1px solid var(--border);
}
.tab {
padding: 12px 24px;
cursor: pointer;
background: #f8f9fa;
border: 1px solid var(--border);
border-bottom: none;
border-radius: 4px 4px 0 0;
margin-right: 5px;
font-weight: 500;
font-size: 0.95rem;
}
.tab.active {
background: white;
border-bottom: 2px solid var(--primary);
font-weight: 600;
}
.tab-content {
display: none;
}
.tab-content.active {
display: block;
}
/* Стили для кнопок редактирования */
.edit-btn {
padding: 4px 8px;
font-size: 0.75rem;
margin-right: 3px;
border-radius: 3px;
}
.edit-btn:hover {
opacity: 0.85;
}
.btn-edit {
background-color: #FFA500;
color: white;
}
.btn-save {
background-color: #3F8C44;
color: white;
display: none;
}
.btn-cancel {
background-color: #D94F4F;
color: white;
display: none;
}
/* Стили для режима редактирования */
.edit-mode input, .edit-mode select, .edit-mode textarea {
width: 100%;
padding: 4px 8px;
border: 1px solid #ccc;
border-radius: 3px;
font-size: 0.9rem;
}
.edit-mode td:not(:last-child) {
padding: 2px;
}
/* Стили для футера */
.app-footer {
position: fixed;
bottom: 0;
left: 0;
right: 0;
background-color: var(--dark);
color: rgba(255, 255, 255, 0.85);
text-align: center;
padding: 8px 16px;
font-size: 0.85rem;
z-index: 2000;
box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.2);
border-top: 1px solid rgba(255, 255, 255, 0.1);
}
.app-footer span {
font-weight: 500;
}
.app-footer .version {
color: #FFA500;
margin-left: 10px;
font-weight: 600;
}
/* Стили для BASH-конструктора */
.bash-console {
width: 100%;
height: 300px;
font-family: monospace;
font-size: 14px;
padding: 15px;
border: 1px solid var(--border);
border-radius: 4px;
background-color: #1e1e1e;
color: #f8f8f2;
resize: vertical;
margin-bottom: 15px;
}
.bash-console:focus {
outline: none;
border-color: var(--primary);
box-shadow: 0 0 0 3px rgba(233, 84, 32, 0.15);
}
.bash-output {
width: 100%;
min-height: 200px;
max-height: 400px;
overflow-y: auto;
font-family: monospace;
font-size: 13px;
padding: 15px;
border: 1px solid var(--border);
border-radius: 4px;
background-color: #2d2d2d;
color: #f8f8f2;
white-space: pre-wrap;
margin-top: 15px;
}
.bash-output .success { color: #a6e22e; }
.bash-output .error { color: #f92672; }
.bash-output .info { color: #66d9ef; }
/* Стили для форм справочников */
.form-row {
display: grid;
grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
gap: 15px;
margin-bottom: 15px;
}
.address-type-group {
display: flex;
gap: 10px;
margin-top: 10px;
padding: 12px;
background-color: #f8f9fa;
border-radius: 4px;
border: 1px solid var(--border);
}
.address-type-option {
display: flex;
align-items: center;
gap: 5px;
}
.address-type-option input[type="radio"] {
margin: 0;
}
/* Стили для прокрутки таблиц - ИСПРАВЛЕНО: фиксированная высота */
.scrollable-table-container {
height: 950px; /* Фиксированная высота ~20 строк (45px * 20 + заголовки + поиск) */
overflow-y: auto;
position: relative;
border: 1px solid var(--border);
border-radius: 4px;
background-color: white;
}
.scrollable-table-container table {
position: relative;
width: 100%;
}
/* ИСПРАВЛЕНО: увеличен z-index для предотвращения наложения строк */
.scrollable-table-container thead {
position: sticky;
top: 0;
background-color: white;
z-index: 100;
box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}
/* ИСПРАВЛЕНО: белый фон для строки поиска и её ячеек */
.search-row {
background-color: white;
border-bottom: 1px solid var(--border);
}
.search-row td {
padding: 8px 14px;
border-top: 1px solid var(--border);
background-color: white;
}
.search-row input.form-control-sm {
height: auto;
padding: 4px 8px;
font-size: 0.85rem;
border-radius: 3px;
}
/* Стили для футера таблицы (пагинация) */
.table-footer {
display: flex;
justify-content: space-between;
align-items: center;
margin-top: 15px;
padding-top: 15px;
border-top: 1px solid #ddd;
}
.table-info {
margin-top: 10px;
color: #666;
font-size: 0.9em;
}
.pagination {
display: flex;
list-style: none;
padding: 0;
margin: 0;
}
.pagination .page-item {
margin: 0 2px;
}
.pagination .page-item a.page-link {
display: block;
padding: 6px 12px;
text-decoration: none;
color: var(--primary);
border: 1px solid var(--border);
border-radius: 4px;
transition: var(--transition);
}
.pagination .page-item a.page-link:hover {
background-color: #f8f9fa;
}
.pagination .page-item.active a.page-link {
background-color: var(--primary);
color: white;
border-color: var(--primary);
}
.pagination .page-item.disabled a.page-link {
color: #ccc;
cursor: not-allowed;
}
/* Стили для кастомных селектов с поиском */
.searchable-select-container {
position: relative;
}
.searchable-input {
width: 100%;
padding: 10px 15px;
border: 1px solid var(--border);
border-radius: 4px;
font-size: 1rem;
}
.searchable-options {
max-height: 200px;
overflow-y: auto;
border: 1px solid #ddd;
border-radius: 4px;
background: white;
position: absolute;
z-index: 1000;
width: 100%;
}
.searchable-option {
padding: 8px 12px;
cursor: pointer;
border-bottom: 1px solid #f0f0f0;
}
.searchable-option:hover {
background-color: #f5f5f5;
}
.searchable-option.selected {
background-color: #e9ecef;
font-weight: 500;
}
/* Адаптивность */
@media (max-width: 992px) {
.sidebar {
width: 220px;
}
.sidebar-header h1 {
font-size: 1.2rem;
}
.nav-link {
padding-left: 35px;
}
}
@media (max-width: 768px) {
.app-container {
flex-direction: column;
}
.sidebar {
width: 100%;
height: auto;
max-height: 40vh;
}
.header {
flex-direction: column;
align-items: flex-start;
padding: 15px;
}
.main-content {
padding-bottom: 70px;
}
.form-row {
grid-template-columns: 1fr;
}
.table-footer {
flex-direction: column;
gap: 10px;
align-items: flex-start;
}
.table-footer select {
margin-bottom: 10px;
}
}
</style>
</head>
<body>
<div class="app-container">
<aside class="sidebar">
<div class="sidebar-header">
<h1>andy-key</h1>
<div class="subtitle">andy-key</div>
</div>
<ul class="nav-menu">
<!-- АДМИНИСТРИРОВАНИЕ -->
<li class="nav-section">
<div class="nav-section-title">Администрирование</div>
<ul class="nav-menu">
<li class="nav-item">
<a href="#" class="nav-link active" data-tab="tasks">Задачи</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="credentials">Учётные записи SSH</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="scripts">Скрипты</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="bash-constructor">BASH-конструктор</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="logs">Журнал</a>
</li>
</ul>
</li>
<!-- СПРАВОЧНИКИ -->
<li class="nav-section">
<div class="nav-section-title">Справочники</div>
<ul class="nav-menu">
<li class="nav-item">
<a href="#" class="nav-link" data-tab="addresses">Адреса</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="employees">Сотрудники</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="workstations">АРМ</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="hosts">Хосты</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="network-equipment">Сетевое оборудование</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="network-mfps">Сетевые МФУ</a>
</li>
<li class="nav-item">
<a href="#" class="nav-link" data-tab="ip-phones">IP телефоны</a>
</li>
</ul>
</li>
</ul>
</aside>
<main class="main-content">
<header class="header">
<h1 id="page-title">Задачи</h1>
<div style="display:flex;gap:10px;align-items:center;">
<span id="current-user" style="color:#555;"></span>
<button onclick="logout()" class="btn btn-sm">Выход</button>
</div>
</header>
<div class="content-container">
<!-- ==================== АДМИНИСТРИРОВАНИЕ ==================== -->
<!-- Задачи -->
<div id="tasks-tab" class="tab-content active">
<div class="card">
<div class="card-header">
<h2>Создать новую задачу</h2>
</div>
<div class="card-body">
<div id="task-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="task-name">Название задачи</label>
<input type="text" id="task-name" class="form-control" placeholder="например, Ежедневное обновление ПО">
</div>
<div class="form-group">
<label for="task-description" class="optional">Описание</label>
<textarea id="task-description" class="form-control" rows="2" placeholder="Подробное описание задачи"></textarea>
</div>
<div class="form-group">
<label for="task-credential">Учётная запись для задачи</label>
<select id="task-credential" class="form-control">
<option value="">-- Выберите учётную запись --</option>
</select>
<small class="form-text">Будет использоваться для подключения ко всем выбранным узлам</small>
</div>
<div class="form-group">
<label>Узлы для задачи</label>
<div class="host-list-container" id="task-hosts-list" style="max-height:300px;overflow-y:auto;border:1px solid #ddd;border-radius:4px;padding:10px;margin-top:10px;">
<p class="empty-state">Сначала добавьте хосты в разделе "Справочники → Хосты"</p>
</div>
<small class="form-text">Отметьте галочкой хосты, на которых нужно выполнить задачу</small>
</div>
<div class="form-group">
<label class="optional">Команды bash</label>
<div id="bash-commands-container">
<div class="bash-command-row" style="display:flex;gap:10px;margin-bottom:10px;">
<input type="text" class="form-control bash-command-input" placeholder="Введите команду bash (например: apt update)">
<button class="btn btn-sm btn-danger" onclick="removeBashCommand(this)">Удалить</button>
</div>
</div>
<button id="add-bash-command-btn" class="btn btn-sm" style="margin-top:10px;">+ Добавить команду</button>
<small class="form-text">Команды будут выполнены в указанном порядке перед скриптами</small>
</div>
<div class="form-group">
<label class="optional">Скрипты из библиотеки</label>
<div id="task-scripts-container">
<p class="empty-state">Добавьте скрипты для задачи</p>
</div>
<button id="add-task-script-btn" class="btn btn-sm" style="margin-top:10px;">+ Добавить скрипт</button>
<small class="form-text">Скрипты будут выполнены после команд bash в указанном порядке</small>
</div>
<div style="display: flex; gap: 10px; margin-top: 20px;">
<button id="save-task" class="btn btn-success">Создать задачу</button>
</div>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список задач</h2>
</div>
<div class="card-body">
<div class="table-container">
<table id="tasks-table">
<thead>
<tr>
<th>ID</th>
<th>Название</th>
<th>Описание</th>
<th>Статус</th>
<th>Создано</th>
<th>Действия</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="6" class="empty-state">Задачи не найдены</td>
</tr>
</tbody>
</table>
</div>
</div>
</div>
</div>
<!-- Учётные записи SSH -->
<div id="credentials-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить учётную запись SSH</h2>
</div>
<div class="card-body">
<div id="credential-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="cred-username">Логин SSH</label>
<input type="text" id="cred-username" class="form-control" placeholder="например, administrator">
</div>
<div class="form-group">
<label for="cred-password">Пароль SSH</label>
<input type="password" id="cred-password" class="form-control" placeholder="Введите пароль">
</div>
<button id="save-credential" class="btn btn-success">Добавить учётную запись</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список учётных записей SSH</h2>
</div>
<div class="card-body">
<div class="table-container">
<table id="credentials-table">
<thead>
<tr>
<th>ID</th>
<th>Логин</th>
<th>Действия</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="3" class="empty-state">Учётные записи не найдены</td>
</tr>
</tbody>
</table>
</div>
</div>
</div>
</div>
<!-- Скрипты -->
<div id="scripts-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить скрипт</h2>
</div>
<div class="card-body">
<div id="script-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="script-name">Название скрипта</label>
<input type="text" id="script-name" class="form-control" placeholder="например, Установка обновлений">
</div>
<div class="form-group">
<label for="script-path">Путь к скрипту</label>
<input type="text" id="script-path" class="form-control" placeholder="/opt/scripts/update.sh">
</div>
<div class="form-group">
<label for="script-parameters" class="optional">Схема параметров (JSON)</label>
<textarea id="script-parameters" class="form-control" rows="5" placeholder='{"param1": "value1", "param2": "value2"}'>{}</textarea>
</div>
<div class="form-group">
<label for="script-description" class="optional">Описание</label>
<textarea id="script-description" class="form-control" rows="3" placeholder="Описание назначения скрипта"></textarea>
</div>
<button id="save-script" class="btn btn-success">Добавить скрипт</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Библиотека скриптов</h2>
</div>
<div class="card-body">
<div class="table-container">
<table id="scripts-table">
<thead>
<tr>
<th>ID</th>
<th>Название</th>
<th>Путь</th>
<th>Параметры</th>
<th>Описание</th>
<th>Действия</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="6" class="empty-state">Скрипты не найдены</td>
</tr>
</tbody>
</table>
</div>
</div>
</div>
</div>
<!-- BASH-конструктор -->
<div id="bash-constructor-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>BASH-конструктор</h2>
</div>
<div class="card-body">
<div id="bash-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="bash-credential">Учётная запись для подключения</label>
<select id="bash-credential" class="form-control">
<option value="">-- Выберите учётную запись --</option>
</select>
</div>
<div class="form-group">
<label for="bash-host">Удалённый хост</label>
<select id="bash-host" class="form-control">
<option value="">-- Выберите хост --</option>
</select>
</div>
<div class="form-group">
<label>Консоль команд</label>
<textarea id="bash-console" class="bash-console" placeholder="Введите команду или скрипт bash...&#10;Например:&#10;apt update&#10;apt upgrade -y&#10;systemctl status nginx"></textarea>
</div>
<button id="bash-execute-btn" class="btn btn-success">Выполнить</button>
<div class="form-group" style="margin-top: 20px;">
<label>Результат выполнения</label>
<div id="bash-output" class="bash-output">
<div class="info">Готово к выполнению команд. Выберите учётную запись и хост, введите команду и нажмите "Выполнить".</div>
</div>
</div>
</div>
</div>
</div>
<!-- Журнал -->
<div id="logs-tab" class="tab-content">
<div class="tabs">
<div class="tab active" onclick="showLogTab('panel')">Действия в панели</div>
<div class="tab" onclick="showLogTab('remote')">Действия с хостами</div>
<div class="tab" onclick="showLogTab('diagnostic')">Диагностика</div>
</div>
<div id="panel-log" class="tab-content active">
<div class="card">
<div class="card-header">
<h2>Действия внутри панели</h2>
<button class="btn btn-sm" onclick="loadLogs('panel')">Обновить</button>
</div>
<div class="card-body">
<div class="log-container" id="panel-log-container" style="max-height:600px;overflow-y:auto;background-color:#f8f9fa;padding:10px;border-radius:4px;font-family:monospace;font-size:0.85rem;">
<p class="empty-state">Журнал пуст</p>
</div>
</div>
</div>
</div>
<div id="remote-log" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Действия с удалёнными хостами</h2>
<button class="btn btn-sm" onclick="loadLogs('remote')">Обновить</button>
</div>
<div class="card-body">
<div class="log-container" id="remote-log-container" style="max-height:600px;overflow-y:auto;background-color:#f8f9fa;padding:10px;border-radius:4px;font-family:monospace;font-size:0.85rem;">
<p class="empty-state">Журнал пуст</p>
</div>
</div>
</div>
</div>
<div id="diagnostic-log" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Диагностические сообщения</h2>
<button class="btn btn-sm" onclick="loadLogs('diagnostic')">Обновить</button>
</div>
<div class="card-body">
<div class="log-container" id="diagnostic-log-container" style="max-height:600px;overflow-y:auto;background-color:#f8f9fa;padding:10px;border-radius:4px;font-family:monospace;font-size:0.85rem;">
<p class="empty-state">Журнал пуст</p>
</div>
</div>
</div>
</div>
</div>
<!-- ==================== СПРАВОЧНИКИ ==================== -->
<!-- Адреса -->
<div id="addresses-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить адрес</h2>
</div>
<div class="card-body">
<div id="address-alert" class="alert" style="display:none;"></div>
<div class="form-row">
<div class="form-group">
<label for="address-street">Улица</label>
<input type="text" id="address-street" class="form-control" placeholder="например, Ленина">
</div>
<div class="form-group">
<label for="address-building">Дом</label>
<input type="text" id="address-building" class="form-control" placeholder="например, 15">
</div>
</div>
<div class="address-type-group">
<div class="address-type-option">
<input type="radio" id="address-type-cabinet" name="address-type" value="cabinet" checked>
<label for="address-type-cabinet" style="margin:0;font-weight:normal;">Кабинет</label>
</div>
<div class="address-type-option">
<input type="radio" id="address-type-corridor" name="address-type" value="corridor">
<label for="address-type-corridor" style="margin:0;font-weight:normal;">Коридор</label>
</div>
<div class="address-type-option">
<input type="radio" id="address-type-service" name="address-type" value="service">
<label for="address-type-service" style="margin:0;font-weight:normal;">Служебное помещение</label>
</div>
</div>
<div id="address-cabinet-group" class="form-row">
<div class="form-group">
<label for="address-cabinet" class="optional">Номер кабинета</label>
<input type="text" id="address-cabinet" class="form-control" placeholder="например, 305">
</div>
</div>
<div id="address-corridor-group" class="form-row" style="display:none;">
<div class="form-group">
<label for="address-corridor">Название коридора</label>
<input type="text" id="address-corridor" class="form-control" placeholder="например, центральный">
</div>
<div class="form-group">
<label for="address-floor">Этаж</label>
<select id="address-floor" class="form-control">
<option value="">-- Выберите этаж --</option>
<option value="1">1</option>
<option value="2">2</option>
<option value="3">3</option>
<option value="4">4</option>
<option value="5">5</option>
</select>
</div>
</div>
<div id="address-service-group" class="form-row" style="display:none;">
<div class="form-group">
<label for="address-service-room">Служебное помещение</label>
<select id="address-service-room" class="form-control">
<option value="">-- Выберите помещение --</option>
<option value="гараж">Гараж</option>
<option value="пост охраны">Пост охраны</option>
</select>
</div>
</div>
<button id="save-address" class="btn btn-success">Добавить адрес</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список адресов</h2>
</div>
<div class="card-body">
<div class="table-container">
<table id="addresses-table">
<thead>
<tr>
<th onclick="sortAddressesTable('street')">Улица</th>
<th onclick="sortAddressesTable('building')">Дом</th>
<th onclick="sortAddressesTable('cabinet')">Кабинет</th>
<th onclick="sortAddressesTable('corridor')">Коридор</th>
<th onclick="sortAddressesTable('floor')">Этаж</th>
<th onclick="sortAddressesTable('service_room')">Служебное помещение</th>
<th>Действия</th>
</tr>
<tr class="search-row">
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterAddressesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterAddressesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterAddressesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterAddressesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterAddressesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterAddressesTable()"></td>
<td></td>
</tr>
</thead>
<tbody>
<tr>
<td colspan="7" class="empty-state">Адреса не найдены</td>
</tr>
</tbody>
</table>
</div>
<div class="table-footer">
<div>
<select id="addresses-page-size" class="form-control" style="width:auto;display:inline-block;">
<option value="10">10 строк</option>
<option value="25" selected>25 строк</option>
<option value="50">50 строк</option>
<option value="100">100 строк</option>
<option value="0">Все строки</option>
</select>
</div>
<div>
<nav>
<ul class="pagination" id="addresses-pagination">
<li class="page-item disabled">
<a class="page-link" href="#" onclick="changeAddressesPage(event, 'prev')" aria-label="Previous">
<span aria-hidden="true">&laquo;</span>
</a>
</li>
<li class="page-item active">
<a class="page-link" href="#" onclick="changeAddressesPage(event, 1)">1</a>
</li>
<li class="page-item">
<a class="page-link" href="#" onclick="changeAddressesPage(event, 'next')" aria-label="Next">
<span aria-hidden="true">&raquo;</span>
</a>
</li>
</ul>
</nav>
</div>
</div>
<div class="table-info">
<span id="addresses-info">Показано 0 из 0 записей</span>
</div>
</div>
</div>
</div>
<!-- Сотрудники -->
<div id="employees-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить сотрудника</h2>
</div>
<div class="card-body">
<div id="employee-alert" class="alert" style="display:none;"></div>
<div class="form-row">
<div class="form-group">
<label for="employee-full-name">ФИО полностью</label>
<input type="text" id="employee-full-name" class="form-control" placeholder="например, Иванов Иван Иванович">
</div>
<div class="form-group">
<label for="employee-short-name" class="optional">ФИО с инициалами</label>
<input type="text" id="employee-short-name" class="form-control" placeholder="например, Иванов И.И.">
</div>
</div>
<div class="form-row">
<div class="form-group">
<label for="employee-phone-city" class="optional">Городской телефон</label>
<input type="text" id="employee-phone-city" class="form-control" placeholder="например, +7 (495) 123-45-67">
</div>
<div class="form-group">
<label for="employee-phone-internal" class="optional">Внутренний телефон</label>
<input type="text" id="employee-phone-internal" class="form-control" placeholder="например, 123">
</div>
</div>
<div class="form-group">
<label for="employee-address">Адрес</label>
<select id="employee-address" class="form-control">
<option value="">-- Выберите адрес --</option>
</select>
<small class="form-text">Сотрудники в служебных помещениях не учитываются</small>
</div>
<button id="save-employee" class="btn btn-success">Добавить сотрудника</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список сотрудников</h2>
</div>
<div class="card-body">
<div class="table-container scrollable-table-container">
<table id="employees-table">
<thead>
<tr>
<th onclick="sortEmployeesTable('full_name')">ФИО полностью</th>
<th onclick="sortEmployeesTable('short_name')">ФИО (инициалы)</th>
<th onclick="sortEmployeesTable('phone_city')">Городской тел.</th>
<th onclick="sortEmployeesTable('phone_internal')">Внутренний тел.</th>
<th onclick="sortEmployeesTable('full_address')">Адрес</th>
<th>Действия</th>
</tr>
<tr class="search-row">
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterEmployeesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterEmployeesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterEmployeesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterEmployeesTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterEmployeesTable()"></td>
<td></td>
</tr>
</thead>
<tbody>
<tr>
<td colspan="6" class="empty-state">Сотрудники не найдены</td>
</tr>
</tbody>
</table>
</div>
<div class="table-footer">
<div>
<select id="employees-page-size" class="form-control" style="width:auto;display:inline-block;">
<option value="10">10 строк</option>
<option value="25" selected>25 строк</option>
<option value="50">50 строк</option>
<option value="100">100 строк</option>
<option value="0">Все строки</option>
</select>
</div>
<div>
<nav>
<ul class="pagination" id="employees-pagination">
<li class="page-item disabled">
<a class="page-link" href="#" onclick="changeEmployeesPage(event, 'prev')" aria-label="Previous">
<span aria-hidden="true">&laquo;</span>
</a>
</li>
<li class="page-item active">
<a class="page-link" href="#" onclick="changeEmployeesPage(event, 1)">1</a>
</li>
<li class="page-item">
<a class="page-link" href="#" onclick="changeEmployeesPage(event, 'next')" aria-label="Next">
<span aria-hidden="true">&raquo;</span>
</a>
</li>
</ul>
</nav>
</div>
</div>
<div class="table-info">
<span id="employees-info">Показано 0 из 0 записей</span>
</div>
</div>
</div>
</div>
</div>
<!-- АРМ -->
<div id="workstations-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить АРМ</h2>
</div>
<div class="card-body">
<div id="workstation-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="workstation-address">Адрес местонахождения</label>
<div id="workstation-address-container" class="searchable-select-container">
<input type="text" id="workstation-address-input" class="form-control searchable-input" placeholder="Начните вводить для поиска..." autocomplete="off">
<div id="workstation-address-options" class="searchable-options" style="display:none;"></div>
<input type="hidden" id="workstation-address-value">
</div>
</div>
<div class="form-group">
<div style="display:flex;align-items:center;gap:10px;">
<input type="checkbox" id="workstation-vacant" style="margin:0;">
<label for="workstation-vacant" style="margin:0;font-weight:500;">Вакантное место</label>
</div>
<small class="form-text">Установите, если место не закреплено за сотрудником</small>
</div>
<div id="workstation-employee-group" class="form-group">
<label for="workstation-employee">Сотрудник (ФИО)</label>
<div id="workstation-employee-container" class="searchable-select-container">
<input type="text" id="workstation-employee-input" class="form-control searchable-input" placeholder="Начните вводить для поиска..." autocomplete="off">
<div id="workstation-employee-options" class="searchable-options" style="display:none;"></div>
<input type="hidden" id="workstation-employee-value">
</div>
</div>
<div class="form-row">
<div class="form-group">
<label for="workstation-inventory">Инвентарный номер</label>
<input type="text" id="workstation-inventory" class="form-control" placeholder="например, ИНВ-2024-001 или б/н">
</div>
<div class="form-group">
<label for="workstation-serial" class="optional">Серийный номер системного блока</label>
<input type="text" id="workstation-serial" class="form-control" placeholder="например, ABC123456789">
</div>
</div>
<div class="form-row">
<div class="form-group">
<label for="workstation-seals" class="optional">Номера пломб</label>
<input type="text" id="workstation-seals" class="form-control" placeholder="через запятую, например: ПЛОМ-001, ПЛОМ-002">
</div>
<div class="form-group">
<label for="workstation-monitors">Количество мониторов</label>
<select id="workstation-monitors" class="form-control">
<option value="1">1</option>
<option value="2" selected>2</option>
<option value="3">3</option>
<option value="4">4</option>
<option value="5">5</option>
</select>
</div>
</div>
<div class="form-group">
<div style="display:flex;align-items:center;gap:10px;">
<input type="checkbox" id="workstation-replacement" style="margin:0;">
<label for="workstation-replacement" style="margin:0;font-weight:500;">Замена оборудования произведена</label>
</div>
</div>
<div id="workstation-replacement-group" style="display:none;">
<div class="form-row">
<div class="form-group">
<label for="workstation-replacement-date">Дата замены</label>
<input type="date" id="workstation-replacement-date" class="form-control">
</div>
<div class="form-group">
<label for="workstation-replacement-letter" class="optional">Номер/дата письма основания</label>
<input type="text" id="workstation-replacement-letter" class="form-control" placeholder="например, №123 от 01.01.2024">
</div>
</div>
</div>
<button id="save-workstation" class="btn btn-success">Добавить АРМ</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список АРМ</h2>
</div>
<div class="card-body">
<!-- ИСПРАВЛЕНО: контейнер прокрутки с фиксированной высотой -->
<div class="table-container scrollable-table-container">
<table id="workstations-table">
<thead>
<tr>
<th onclick="sortWorkstationsTable('full_address')">Адрес</th>
<th onclick="sortWorkstationsTable('employee_short_name')">Сотрудник</th>
<th onclick="sortWorkstationsTable('inventory_number')">Инв. номер</th>
<th onclick="sortWorkstationsTable('serial_number')">Серийный номер</th>
<th onclick="sortWorkstationsTable('seal_numbers')">Номера пломб</th>
<th onclick="sortWorkstationsTable('monitor_count')">Мониторы</th>
<th onclick="sortWorkstationsTable('is_vacant')">Статус</th>
<th onclick="sortWorkstationsTable('replacement_date')">Дата замены</th>
<th onclick="sortWorkstationsTable('replacement_letter')">Письмо основания</th>
<th>Действия</th>
</tr>
<tr class="search-row">
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td><input type="text" class="form-control form-control-sm" placeholder="Поиск..." oninput="filterWorkstationsTable()"></td>
<td></td>
</tr>
</thead>
<tbody id="workstations-table-body">
<tr>
<td colspan="10" class="empty-state">АРМ не найдены</td>
</tr>
</tbody>
</table>
</div>
<div class="table-footer">
<div>
<select id="workstations-page-size" class="form-control" style="width:auto;display:inline-block;">
<option value="10">10 строк</option>
<option value="25" selected>25 строк</option>
<option value="50">50 строк</option>
<option value="100">100 строк</option>
<option value="0">Все строки</option>
</select>
</div>
<div>
<nav>
<ul class="pagination" id="workstations-pagination">
<li class="page-item disabled">
<a class="page-link" href="#" onclick="changeWorkstationsPage(event, 'prev')" aria-label="Previous">
<span aria-hidden="true">&laquo;</span>
</a>
</li>
<li class="page-item active">
<a class="page-link" href="#" onclick="changeWorkstationsPage(event, 1)">1</a>
</li>
<li class="page-item">
<a class="page-link" href="#" onclick="changeWorkstationsPage(event, 'next')" aria-label="Next">
<span aria-hidden="true">&raquo;</span>
</a>
</li>
</ul>
</nav>
</div>
<div>
<span id="workstations-info">Показано 0 из 0 записей</span>
</div>
</div>
</div>
</div>
</div>
<!-- ... остальные разделы (хосты, оборудование, МФУ, телефоны) без изменений ... -->
<!-- Хосты -->
<div id="hosts-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить хост</h2>
</div>
<div class="card-body">
<div id="host-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="host-address">Адрес</label>
<select id="host-address" class="form-control">
<option value="">-- Выберите адрес --</option>
</select>
</div>
<div class="form-group">
<label for="host-employee" class="optional">Сотрудник (ФИО)</label>
<select id="host-employee" class="form-control">
<option value="">-- Не закреплён за сотрудником --</option>
</select>
</div>
<div class="form-row">
<div class="form-group">
<label for="host-ip">IP адрес</label>
<input type="text" id="host-ip" class="form-control" placeholder="например, 192.168.1.100">
</div>
<div class="form-group">
<label for="host-ssh-port" class="optional">SSH порт</label>
<input type="number" id="host-ssh-port" class="form-control" value="22" min="1" max="65535">
</div>
</div>
<div class="form-group">
<div style="display:flex;align-items:center;gap:10px;">
<input type="checkbox" id="host-enabled" checked style="margin:0;">
<label for="host-enabled" style="margin:0;font-weight:500;">Включён</label>
</div>
</div>
<button id="save-host" class="btn btn-success">Добавить хост</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список хостов</h2>
</div>
<div class="card-body">
<div class="table-container">
<table id="hosts-table">
<thead>
<tr>
<th>Адрес</th>
<th>Сотрудник</th>
<th>IP адрес</th>
<th>SSH порт</th>
<th>Статус</th>
<th>Действия</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="6" class="empty-state">Хосты не найдены</td>
</tr>
</tbody>
</table>
</div>
</div>
</div>
</div>
<!-- Сетевое оборудование -->
<div id="network-equipment-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить сетевое оборудование</h2>
</div>
<div class="card-body">
<div id="network-equipment-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="network-equipment-address">Адрес</label>
<select id="network-equipment-address" class="form-control">
<option value="">-- Выберите адрес --</option>
</select>
</div>
<div class="form-row">
<div class="form-group">
<label for="network-equipment-category">Категория</label>
<select id="network-equipment-category" class="form-control">
<option value="">-- Выберите категорию --</option>
<option value="коммутатор">Коммутатор</option>
<option value="роутер">Роутер</option>
</select>
</div>
<div class="form-group">
<label for="network-equipment-type">Тип</label>
<select id="network-equipment-type" class="form-control">
<option value="">-- Выберите тип --</option>
<option value="управляемый">Управляемый</option>
<option value="неуправляемый">Неуправляемый</option>
</select>
</div>
</div>
<div class="form-row">
<div class="form-group">
<label for="network-equipment-model">Модель</label>
<input type="text" id="network-equipment-model" class="form-control" placeholder="например, Cisco Catalyst 2960">
</div>
<div class="form-group">
<label for="network-equipment-ports">Число портов</label>
<input type="number" id="network-equipment-ports" class="form-control" min="1" value="24">
</div>
</div>
<button id="save-network-equipment" class="btn btn-success">Добавить оборудование</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список сетевого оборудования</h2>
</div>
<div class="card-body">
<div class="table-container">
<table id="network-equipment-table">
<thead>
<tr>
<th>Адрес</th>
<th>Категория</th>
<th>Тип</th>
<th>Модель</th>
<th>Порты</th>
<th>Действия</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="6" class="empty-state">Оборудование не найдено</td>
</tr>
</tbody>
</table>
</div>
</div>
</div>
</div>
<!-- Сетевые МФУ -->
<div id="network-mfps-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить сетевое МФУ</h2>
</div>
<div class="card-body">
<div id="network-mfp-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="network-mfp-address">Адрес</label>
<select id="network-mfp-address" class="form-control">
<option value="">-- Выберите адрес --</option>
</select>
</div>
<div class="form-row">
<div class="form-group">
<label for="network-mfp-model">Модель</label>
<input type="text" id="network-mfp-model" class="form-control" placeholder="например, HP LaserJet Pro MFP M430fdw">
</div>
<div class="form-group">
<label for="network-mfp-ip">IP адрес</label>
<input type="text" id="network-mfp-ip" class="form-control" placeholder="например, 192.168.1.200">
</div>
</div>
<div class="form-row">
<div class="form-group">
<label for="network-mfp-hostname" class="optional">Имя хоста</label>
<input type="text" id="network-mfp-hostname" class="form-control" placeholder="например, mfp-printer-01">
</div>
<div class="form-group">
<label for="network-mfp-serial">Серийный номер</label>
<input type="text" id="network-mfp-serial" class="form-control" placeholder="например, CN12345678">
</div>
</div>
<button id="save-network-mfp" class="btn btn-success">Добавить МФУ</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список сетевых МФУ</h2>
</div>
<div class="card-body">
<div class="table-container">
<table id="network-mfps-table">
<thead>
<tr>
<th>Адрес</th>
<th>Модель</th>
<th>IP адрес</th>
<th>Имя хоста</th>
<th>Серийный номер</th>
<th>Действия</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="6" class="empty-state">МФУ не найдены</td>
</tr>
</tbody>
</table>
</div>
</div>
</div>
</div>
<!-- IP телефоны -->
<div id="ip-phones-tab" class="tab-content">
<div class="card">
<div class="card-header">
<h2>Добавить IP телефон</h2>
</div>
<div class="card-body">
<div id="ip-phone-alert" class="alert" style="display:none;"></div>
<div class="form-group">
<label for="ip-phone-address">Адрес</label>
<select id="ip-phone-address" class="form-control">
<option value="">-- Выберите адрес --</option>
</select>
</div>
<div class="form-row">
<div class="form-group">
<label for="ip-phone-employee">ФИО сотрудника (полностью)</label>
<input type="text" id="ip-phone-employee" class="form-control" placeholder="например, Иванов Иван Иванович">
</div>
<div class="form-group">
<label for="ip-phone-ip">IP адрес</label>
<input type="text" id="ip-phone-ip" class="form-control" placeholder="например, 192.168.1.210">
</div>
</div>
<button id="save-ip-phone" class="btn btn-success">Добавить телефон</button>
</div>
</div>
<div class="card">
<div class="card-header">
<h2>Список IP телефонов</h2>
</div>
<div class="card-body">
<div class="table-container">
<table id="ip-phones-table">
<thead>
<tr>
<th>Адрес</th>
<th>Сотрудник</th>
<th>IP адрес</th>
<th>Действия</th>
</tr>
</thead>
<tbody>
<tr>
<td colspan="4" class="empty-state">Телефоны не найдены</td>
</tr>
</tbody>
</table>
</div>
</div>
</div>
</div>
</div>
</main>
</div>
<!-- ГЛОБАЛЬНЫЙ ФУТЕР -->
<footer class="app-footer">
<span>{{.FooterText}}</span>
<span class="version">Версия {{.AppVersion}}</span>
</footer>
<script>
`