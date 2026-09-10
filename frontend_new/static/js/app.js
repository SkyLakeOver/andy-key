// Основное приложение Alpine.js для andy-key v2.2.0

function app() {
    return {
        activeTab: 'tasks',
        currentUser: '',
        pageTitle: 'Задачи',
        
        init() {
            // Загрузка информации о текущем пользователе
            this.loadCurrentUser();
            
            // Отслеживание изменений активной вкладки
            this.$watch('activeTab', (value) => {
                this.updatePageTitle(value);
                this.loadContent(value);
            });
        },
        
        loadCurrentUser() {
            fetch('/api/user')
                .then(response => response.json())
                .then(data => {
                    this.currentUser = data.username || 'Гость';
                })
                .catch(err => {
                    console.error('Ошибка загрузки пользователя:', err);
                    this.currentUser = 'Гость';
                });
        },
        
        updatePageTitle(tab) {
            const titles = {
                'tasks': 'Задачи',
                'credentials': 'Учётные записи SSH',
                'scripts': 'Скрипты',
                'bash-constructor': 'BASH-конструктор',
                'logs': 'Журнал',
                'addresses': 'Адреса',
                'employees': 'Сотрудники',
                'workstations': 'АРМ',
                'hosts': 'Хосты',
                'network-equipment': 'Сетевое оборудование',
                'network-mfps': 'Сетевые МФУ',
                'ip-phones': 'IP телефоны'
            };
            this.pageTitle = titles[tab] || 'Задачи';
        },
        
        loadContent(tab) {
            // HTMX автоматически загрузит контент через hx-get
            // Здесь можно добавить дополнительную логику при необходимости
            console.log('Загрузка контента для вкладки:', tab);
        },
        
        logout() {
            if (confirm('Вы уверены, что хотите выйти?')) {
                window.location.href = '/logout';
            }
        }
    }
}
