// Основное приложение Alpine.js для andy-key v2.2.0

document.addEventListener('alpine:init', () => {
    // Глобальное состояние приложения
    Alpine.data('app', () => ({
        activeTab: 'tasks',
        pageTitle: 'Задачи',
        currentUser: '',
        
        init() {
            fetch('/api/user')
                .then(response => response.json())
                .then(data => {
                    this.currentUser = data.username || 'Гость';
                })
                .catch(err => console.error('Ошибка загрузки пользователя:', err));
        },
        
        logout() {
            fetch('/logout', { method: 'POST' })
                .then(() => { window.location.href = '/login'; })
                .catch(() => { window.location.href = '/login'; });
        }
    }));

    // Компонент для страницы задач
    Alpine.data('tasksPage', () => ({
        searchQuery: '',
        filterStatus: 'all',
        showAddModal: false,
        formData: { title: '', description: '', status: 'new', assignee: '', script_id: '' },
        tasks: [],
        scripts: [],

        init() {
            this.loadTasks();
            this.loadScripts();
        },

        loadTasks() {
            fetch('/api/tasks')
                .then(r => r.json())
                .then(data => { this.tasks = data || []; })
                .catch(err => console.error('Ошибка загрузки задач:', err));
        },

        loadScripts() {
            fetch('/api/scripts')
                .then(r => r.json())
                .then(data => { this.scripts = data || []; })
                .catch(err => console.error('Ошибка загрузки скриптов:', err));
        },

        get filteredTasks() {
            return this.tasks.filter(task => {
                const s = this.searchQuery.toLowerCase();
                const matchesSearch = task.title.toLowerCase().includes(s) || task.description.toLowerCase().includes(s);
                const matchesStatus = this.filterStatus === 'all' || task.status === this.filterStatus;
                return matchesSearch && matchesStatus;
            });
        },

        openAddModal() {
            this.formData = { title: '', description: '', status: 'new', assignee: '', script_id: '' };
            this.showAddModal = true;
        },

        closeAddModal() { this.showAddModal = false; },

        submitTask() {
            fetch('/api/tasks', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.formData)
            }).then(r => {
                if (r.ok) { this.closeAddModal(); this.loadTasks(); }
                else { alert('Ошибка при создании задачи'); }
            }).catch(() => alert('Ошибка при создании задачи'));
        }
    }));

    // Компонент для страницы сотрудников
    Alpine.data('employeesPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: { fio: '', position: '', department: '', phone: '', email: '' },
        employees: [],
        
        init() { this.loadEmployees(); },
        
        loadEmployees() {
            fetch('/api/employees')
                .then(r => r.json())
                .then(data => { this.employees = data || []; })
                .catch(err => console.error('Ошибка загрузки сотрудников:', err));
        },
        
        get filteredEmployees() {
            return this.employees.filter(emp => {
                const s = this.searchQuery.toLowerCase();
                return emp.fio.toLowerCase().includes(s) || emp.position.toLowerCase().includes(s) || emp.department.toLowerCase().includes(s);
            });
        },
        
        openAddModal() { this.formData = { fio: '', position: '', department: '', phone: '', email: '' }; this.showAddModal = true; },
        closeAddModal() { this.showAddModal = false; },
        
        submitEmployee() {
            fetch('/api/employees', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.formData) })
            .then(r => { if (r.ok) { this.closeAddModal(); this.loadEmployees(); } else { alert('Ошибка'); } })
            .catch(() => alert('Ошибка'));
        }
    }));

    // Компонент для страницы рабочих мест
    Alpine.data('workstationsPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: { inventory_number: '', model: '', serial_number: '', employee_id: null, status: 'active' },
        workstations: [],
        
        init() { this.loadWorkstations(); },
        
        loadWorkstations() {
            fetch('/api/workstations')
                .then(r => r.json())
                .then(data => { this.workstations = data || []; })
                .catch(err => console.error('Ошибка загрузки АРМ:', err));
        },
        
        get filteredWorkstations() {
            return this.workstations.filter(ws => {
                const s = this.searchQuery.toLowerCase();
                return ws.inventory_number.toLowerCase().includes(s) || ws.model.toLowerCase().includes(s) || ws.serial_number.toLowerCase().includes(s);
            });
        },
        
        openAddModal() { this.formData = { inventory_number: '', model: '', serial_number: '', employee_id: null, status: 'active' }; this.showAddModal = true; },
        closeAddModal() { this.showAddModal = false; },
        
        submitWorkstation() {
            fetch('/api/workstations', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.formData) })
            .then(r => { if (r.ok) { this.closeAddModal(); this.loadWorkstations(); } else { alert('Ошибка'); } })
            .catch(() => alert('Ошибка'));
        }
    }));

    // Компонент для страницы адресов
    Alpine.data('addressesPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: { address: '', description: '' },
        addresses: [],
        
        init() { this.loadAddresses(); },
        
        loadAddresses() {
            fetch('/api/addresses')
                .then(r => r.json())
                .then(data => { this.addresses = data || []; })
                .catch(err => console.error('Ошибка загрузки адресов:', err));
        },
        
        get filteredAddresses() {
            return this.addresses.filter(addr => {
                const s = this.searchQuery.toLowerCase();
                return addr.address.toLowerCase().includes(s) || addr.description.toLowerCase().includes(s);
            });
        },
        
        openAddModal() { this.formData = { address: '', description: '' }; this.showAddModal = true; },
        closeAddModal() { this.showAddModal = false; },
        
        submitAddress() {
            fetch('/api/addresses', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.formData) })
            .then(r => { if (r.ok) { this.closeAddModal(); this.loadAddresses(); } else { alert('Ошибка'); } })
            .catch(() => alert('Ошибка'));
        }
    }));

    // Компонент для страницы хостов
    Alpine.data('hostsPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: { hostname: '', ip_address: '', mac_address: '', os_type: '', location: '' },
        hosts: [],
        
        init() { this.loadHosts(); },
        
        loadHosts() {
            fetch('/api/hosts')
                .then(r => r.json())
                .then(data => { this.hosts = data || []; })
                .catch(err => console.error('Ошибка загрузки хостов:', err));
        },
        
        get filteredHosts() {
            return this.hosts.filter(host => {
                const s = this.searchQuery.toLowerCase();
                return host.hostname.toLowerCase().includes(s) || host.ip_address.includes(this.searchQuery) || host.mac_address.toLowerCase().includes(s);
            });
        },
        
        openAddModal() { this.formData = { hostname: '', ip_address: '', mac_address: '', os_type: '', location: '' }; this.showAddModal = true; },
        closeAddModal() { this.showAddModal = false; },
        
        submitHost() {
            fetch('/api/hosts', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.formData) })
            .then(r => { if (r.ok) { this.closeAddModal(); this.loadHosts(); } else { alert('Ошибка'); } })
            .catch(() => alert('Ошибка'));
        }
    }));

    // Компонент для страницы сетевого оборудования
    Alpine.data('networkEquipmentPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: { name: '', model: '', serial_number: '', ip_address: '', location: '', type: 'switch' },
        equipment: [],
        
        init() { this.loadEquipment(); },
        
        loadEquipment() {
            fetch('/api/network-equipment')
                .then(r => r.json())
                .then(data => { this.equipment = data || []; })
                .catch(err => console.error('Ошибка загрузки оборудования:', err));
        },
        
        get filteredEquipment() {
            return this.equipment.filter(eq => {
                const s = this.searchQuery.toLowerCase();
                return eq.name.toLowerCase().includes(s) || eq.model.toLowerCase().includes(s) || eq.ip_address.includes(this.searchQuery);
            });
        },
        
        openAddModal() { this.formData = { name: '', model: '', serial_number: '', ip_address: '', location: '', type: 'switch' }; this.showAddModal = true; },
        closeAddModal() { this.showAddModal = false; },
        
        submitEquipment() {
            fetch('/api/network-equipment', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.formData) })
            .then(r => { if (r.ok) { this.closeAddModal(); this.loadEquipment(); } else { alert('Ошибка'); } })
            .catch(() => alert('Ошибка'));
        }
    }));

    // Компонент для страницы сетевых МФУ
    Alpine.data('networkMfpsPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: { name: '', model: '', serial_number: '', ip_address: '', location: '' },
        mfps: [],
        
        init() { this.loadMfps(); },
        
        loadMfps() {
            fetch('/api/network-mfps')
                .then(r => r.json())
                .then(data => { this.mfps = data || []; })
                .catch(err => console.error('Ошибка загрузки МФУ:', err));
        },
        
        get filteredMfps() {
            return this.mfps.filter(mfp => {
                const s = this.searchQuery.toLowerCase();
                return mfp.name.toLowerCase().includes(s) || mfp.model.toLowerCase().includes(s) || mfp.ip_address.includes(this.searchQuery);
            });
        },
        
        openAddModal() { this.formData = { name: '', model: '', serial_number: '', ip_address: '', location: '' }; this.showAddModal = true; },
        closeAddModal() { this.showAddModal = false; },
        
        submitMfp() {
            fetch('/api/network-mfps', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.formData) })
            .then(r => { if (r.ok) { this.closeAddModal(); this.loadMfps(); } else { alert('Ошибка'); } })
            .catch(() => alert('Ошибка'));
        }
    }));

    // Компонент для страницы IP телефонов
    Alpine.data('ipPhonesPage', () => ({
        searchQuery: '',
        showAddModal: false,
        formData: { name: '', model: '', serial_number: '', mac_address: '', extension: '', location: '' },
        phones: [],
        
        init() { this.loadPhones(); },
        
        loadPhones() {
            fetch('/api/ip-phones')
                .then(r => r.json())
                .then(data => { this.phones = data || []; })
                .catch(err => console.error('Ошибка загрузки телефонов:', err));
        },
        
        get filteredPhones() {
            return this.phones.filter(phone => {
                const s = this.searchQuery.toLowerCase();
                return phone.name.toLowerCase().includes(s) || phone.model.toLowerCase().includes(s) || phone.extension.includes(this.searchQuery);
            });
        },
        
        openAddModal() { this.formData = { name: '', model: '', serial_number: '', mac_address: '', extension: '', location: '' }; this.showAddModal = true; },
        closeAddModal() { this.showAddModal = false; },
        
        submitPhone() {
            fetch('/api/ip-phones', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(this.formData) })
            .then(r => { if (r.ok) { this.closeAddModal(); this.loadPhones(); } else { alert('Ошибка'); } })
            .catch(() => alert('Ошибка'));
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
                body: JSON.stringify({ username: this.username, password: this.password })
            })
            .then(response => {
                if (response.ok) { window.location.href = '/'; }
                else { return response.text().then(t => { throw new Error(t || 'Ошибка входа'); }); }
            })
            .catch(err => { this.error = err.message; });
        }
    }));
});
