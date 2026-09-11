// Основное приложение Alpine.js для andy-key v2.2.0

// Глобальное состояние приложения
document.addEventListener('alpine:init', () => {
    Alpine.data('app', () => ({
        activeTab: 'tasks',
        pageTitle: 'Задачи',
        currentUser: '',
        
        init() {
            // Загрузка текущего пользователя
            fetch('/api/user')
                .then(response => response.json())
                .then(data => {
                    this.currentUser = data.username || 'Гость';
                })
                .catch(err => console.error('Ошибка загрузки пользователя:', err));
        },
        
        setActiveTab(tab, title) {
            this.activeTab = tab;
            this.pageTitle = title;
        }
    }));

    // Компонент для страницы задач
    Alpine.data('tasksPage', () => ({
        searchQuery: '',
        filterStatus: 'all',
        showAddModal: false,
        formData: {
            title: '',
            description: '',
            status: 'new',
            assignee: '',
            script_id: ''
        },
        tasks: [],
        scripts: [],

        init() {
            this.loadTasks();
            this.loadScripts();
        },

        loadTasks() {
            fetch('/api/tasks')
                .then(response => response.json())
                .then(data => {
                    this.tasks = data || [];
                })
                .catch(err => console.error('Ошибка загрузки задач:', err));
        },

        loadScripts() {
            fetch('/api/scripts')
                .then(response => response.json())
                .then(data => {
                    this.scripts = data || [];
                })
                .catch(err => console.error('Ошибка загрузки скриптов:', err));
        },

        get filteredTasks() {
            return this.tasks.filter(task => {
                const matchesSearch = task.title.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                                     task.description.toLowerCase().includes(this.searchQuery.toLowerCase());
                const matchesStatus = this.filterStatus === 'all' || task.status === this.filterStatus;
                return matchesSearch && matchesStatus;
            });
        },

        openAddModal() {
            this.formData = { title: '', description: '', status: 'new', assignee: '', script_id: '' };
            this.showAddModal = true;
        },

        closeAddModal() {
            this.showAddModal = false;
        },

        submitTask() {
            fetch('/api/tasks', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            })
            .then(response => {
                if (response.ok) {
                    this.closeAddModal();
                    this.loadTasks();
                } else {
                    alert('Ошибка при создании задачи');
                }
            })
            .catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка при создании задачи');
            });
        }
    }));

    // Компонент для страницы сотрудников
    Alpine.data('employeesPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: {
            fio: '',
            position: '',
            department: '',
            phone: '',
            email: ''
        },
        employees: [],
        
        init() {
            this.loadEmployees();
        },
        
        loadEmployees() {
            fetch('/api/employees')
                .then(response => response.json())
                .then(data => {
                    this.employees = data || [];
                })
                .catch(err => console.error('Ошибка загрузки сотрудников:', err));
        },
        
        get filteredEmployees() {
            return this.employees.filter(emp => 
                emp.fio.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                emp.position.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                emp.department.toLowerCase().includes(this.searchQuery.toLowerCase())
            );
        },
        
        openAddModal() {
            this.formData = { fio: '', position: '', department: '', phone: '', email: '' };
            this.showAddModal = true;
        },
        
        closeAddModal() {
            this.showAddModal = false;
        },
        
        submitEmployee() {
            fetch('/api/employees', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            })
            .then(response => {
                if (response.ok) {
                    this.closeAddModal();
                    this.loadEmployees();
                } else {
                    alert('Ошибка при создании сотрудника');
                }
            })
            .catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка при создании сотрудника');
            });
        }
    }));

// Компонент для страницы рабочих мест
Alpine.data('workstationsPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: {
            inventory_number: '',
            model: '',
            serial_number: '',
            employee_id: null,
            status: 'active'
        },
        workstations: [],
        
        init() {
            this.loadWorkstations();
        },
        
        loadWorkstations() {
            fetch('/api/workstations')
                .then(response => response.json())
                .then(data => {
                    this.workstations = data || [];
                })
                .catch(err => console.error('Ошибка загрузки АРМ:', err));
        },
        
        get filteredWorkstations() {
            return this.workstations.filter(ws => 
                ws.inventory_number.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                ws.model.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                ws.serial_number.toLowerCase().includes(this.searchQuery.toLowerCase())
            );
        },
        
        openAddModal() {
            this.formData = { inventory_number: '', model: '', serial_number: '', employee_id: null, status: 'active' };
            this.showAddModal = true;
        },
        
        closeAddModal() {
            this.showAddModal = false;
        },
        
        submitWorkstation() {
            fetch('/api/workstations', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            })
            .then(response => {
                if (response.ok) {
                    this.closeAddModal();
                    this.loadWorkstations();
                } else {
                    alert('Ошибка при создании АРМ');
                }
            })
            .catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка при создании АРМ');
            });
        }
    }
}));

// Компонент для страницы адресов
Alpine.data('addressesPage', () => ({

        searchQuery: '',
        showAddModal: false,
        formData: {
            address: '',
            description: ''
        },
        addresses: [],
        
        init() {
            this.loadAddresses();
        },
        
        loadAddresses() {
            fetch('/api/addresses')
                .then(response => response.json())
                .then(data => {
                    this.addresses = data || [];
                })
                .catch(err => console.error('Ошибка загрузки адресов:', err));
        },
        
        get filteredAddresses() {
            return this.addresses.filter(addr => 
                addr.address.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                addr.description.toLowerCase().includes(this.searchQuery.toLowerCase())
            );
        },
        
        openAddModal() {
            this.formData = { address: '', description: '' };
            this.showAddModal = true;
        },
        
        closeAddModal() {
            this.showAddModal = false;
        },
        
        submitAddress() {
            fetch('/api/addresses', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            })
            .then(response => {
                if (response.ok) {
                    this.closeAddModal();
                    this.loadAddresses();
                } else {
                    alert('Ошибка при создании адреса');
                }
            })
            .catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка при создании адреса');
            });
        }
    }
}));

// Компонент для страницы хостов
Alpine.data('hostsPage', () => ({

        searchQuery: '',
        showAddModal: false,
        formData: {
            hostname: '',
            ip_address: '',
            mac_address: '',
            os_type: '',
            location: ''
        },
        hosts: [],
        
        init() {
            this.loadHosts();
        },
        
        loadHosts() {
            fetch('/api/hosts')
                .then(response => response.json())
                .then(data => {
                    this.hosts = data || [];
                })
                .catch(err => console.error('Ошибка загрузки хостов:', err));
        },
        
        get filteredHosts() {
            return this.hosts.filter(host => 
                host.hostname.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                host.ip_address.includes(this.searchQuery) ||
                host.mac_address.toLowerCase().includes(this.searchQuery.toLowerCase())
            );
        },
        
        openAddModal() {
            this.formData = { hostname: '', ip_address: '', mac_address: '', os_type: '', location: '' };
            this.showAddModal = true;
        },
        
        closeAddModal() {
            this.showAddModal = false;
        },
        
        submitHost() {
            fetch('/api/hosts', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            })
            .then(response => {
                if (response.ok) {
                    this.closeAddModal();
                    this.loadHosts();
                } else {
                    alert('Ошибка при создании хоста');
                }
            })
            .catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка при создании хоста');
            });
        }
    }
}));

// Компонент для страницы сетевого оборудования
Alpine.data('networkEquipmentPage', () => ({

        searchQuery: '',
        showAddModal: false,
        formData: {
            name: '',
            model: '',
            serial_number: '',
            ip_address: '',
            location: '',
            type: 'switch'
        },
        equipment: [],
        
        init() {
            this.loadEquipment();
        },
        
        loadEquipment() {
            fetch('/api/network-equipment')
                .then(response => response.json())
                .then(data => {
                    this.equipment = data || [];
                })
                .catch(err => console.error('Ошибка загрузки сетевого оборудования:', err));
        },
        
        get filteredEquipment() {
            return this.equipment.filter(eq => 
                eq.name.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                eq.model.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                eq.ip_address.includes(this.searchQuery)
            );
        },
        
        openAddModal() {
            this.formData = { name: '', model: '', serial_number: '', ip_address: '', location: '', type: 'switch' };
            this.showAddModal = true;
        },
        
        closeAddModal() {
            this.showAddModal = false;
        },
        
        submitEquipment() {
            fetch('/api/network-equipment', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            })
            .then(response => {
                if (response.ok) {
                    this.closeAddModal();
                    this.loadEquipment();
                } else {
                    alert('Ошибка при создании сетевого оборудования');
                }
            })
            .catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка при создании сетевого оборудования');
            });
        }
    }
}));

// Компонент для страницы сетевых МФУ
Alpine.data('networkMfpsPage', () => ({

        searchQuery: '',
        showAddModal: false,
        formData: {
            name: '',
            model: '',
            serial_number: '',
            ip_address: '',
            location: ''
        },
        mfps: [],
        
        init() {
            this.loadMfps();
        },
        
        loadMfps() {
            fetch('/api/network-mfps')
                .then(response => response.json())
                .then(data => {
                    this.mfps = data || [];
                })
                .catch(err => console.error('Ошибка загрузки МФУ:', err));
        },
        
        get filteredMfps() {
            return this.mfps.filter(mfp => 
                mfp.name.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                mfp.model.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                mfp.ip_address.includes(this.searchQuery)
            );
        },
        
        openAddModal() {
            this.formData = { name: '', model: '', serial_number: '', ip_address: '', location: '' };
            this.showAddModal = true;
        },
        
        closeAddModal() {
            this.showAddModal = false;
        },
        
        submitMfp() {
            fetch('/api/network-mfps', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            })
            .then(response => {
                if (response.ok) {
                    this.closeAddModal();
                    this.loadMfps();
                } else {
                    alert('Ошибка при создании МФУ');
                }
            })
            .catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка при создании МФУ');
            });
        }
    }
}));

// Компонент для страницы IP телефонов
Alpine.data('ipPhonesPage', () => ({

        searchQuery: '',
        showAddModal: false,
        formData: {
            name: '',
            model: '',
            serial_number: '',
            mac_address: '',
            extension: '',
            location: ''
        },
        phones: [],
        
        init() {
            this.loadPhones();
        },
        
        loadPhones() {
            fetch('/api/ip-phones')
                .then(response => response.json())
                .then(data => {
                    this.phones = data || [];
                })
                .catch(err => console.error('Ошибка загрузки IP телефонов:', err));
        },
        
        get filteredPhones() {
            return this.phones.filter(phone => 
                phone.name.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                phone.model.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
                phone.extension.includes(this.searchQuery)
            );
        },
        
        openAddModal() {
            this.formData = { name: '', model: '', serial_number: '', mac_address: '', extension: '', location: '' };
            this.showAddModal = true;
        },
        
        closeAddModal() {
            this.showAddModal = false;
        },
        
        submitPhone() {
            fetch('/api/ip-phones', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            })
            .then(response => {
                if (response.ok) {
                    this.closeAddModal();
                    this.loadPhones();
                } else {
                    alert('Ошибка при создании IP телефона');
                }
            })
            .catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка при создании IP телефона');
            });
        }
    }
}));

// Основной компонент приложения
Alpine.data('app', () => ({
    

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
}));

// Компонент для формы входа
Alpine.data('loginForm', () => ({

        username: '',
        password: '',
        error: '',
        
        submit() {
            fetch('/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username: this.username,
                    password: this.password
                })
            })
            .then(response => {
                if (response.ok) {
                    window.location.href = '/';
                } else {
                    return response.text().then(text => {
                        throw new Error(text || 'Ошибка входа');
                    });
                }
            })
            .catch(err => {
                this.error = err.message;
            });
        }
    }
}));

});
