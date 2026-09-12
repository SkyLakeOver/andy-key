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
        showEditModal: false,
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
        filterAddress: '',
        showAddModal: false,
        showEditModal: false,
        formData: { fio: '', position: '', department: '', phone: '', email: '', address_id: null },
        employees: [],
        addresses: [],
        
        init() { this.loadEmployees(); this.loadAddresses(); },
        
        loadEmployees() {
            fetch('/api/employees')
                .then(r => r.json())
                .then(data => { this.employees = data || []; })
                .catch(err => console.error('Ошибка загрузки сотрудников:', err));
        },

        loadAddresses() {
            fetch('/api/addresses')
                .then(r => r.json())
                .then(data => { this.addresses = data || []; })
                .catch(err => console.error('Ошибка загрузки адресов:', err));
        },
        
        get filteredEmployees() {
            return this.employees.filter(emp => {
                const s = this.searchQuery.toLowerCase();
                const matchesSearch = emp.fio.toLowerCase().includes(s) || (emp.position && emp.position.toLowerCase().includes(s)) || (emp.department && emp.department.toLowerCase().includes(s));
                const matchesAddress = !this.filterAddress || emp.address_id == this.filterAddress;
                return matchesSearch && matchesAddress;
            });
        },
        
        openAddModal() { this.formData = { fio: '', position: '', department: '', phone: '', email: '', address_id: null }; this.showAddModal = true; },
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
        filterVacant: '',
        filterAddress: '',
        showAddModal: false,
        showEditModal: false,
        formData: { inventory_number: '', model: '', serial_number: '', employee_id: null, status: 'active', address_id: null },
        workstations: [],
        addresses: [],
        employees: [],
        
        init() { this.loadWorkstations(); this.loadAddresses(); },
        
        loadWorkstations() {
            fetch('/api/workstations')
                .then(r => r.json())
                .then(data => { this.workstations = data || []; })
                .catch(err => console.error('Ошибка загрузки АРМ:', err));
        },

        loadAddresses() {
            fetch('/api/addresses')
                .then(r => r.json())
                .then(data => { this.addresses = data || []; })
                .catch(err => console.error('Ошибка загрузки адресов:', err));
        },
        
        get filteredWorkstations() {
            return this.workstations.filter(ws => {
                const s = this.searchQuery.toLowerCase();
                const matchesSearch = ws.inventory_number.toLowerCase().includes(s) || (ws.model && ws.model.toLowerCase().includes(s)) || (ws.serial_number && ws.serial_number.toLowerCase().includes(s));
                const matchesVacant = !this.filterVacant || (this.filterVacant === 'true' && !ws.employee_id) || (this.filterVacant === 'false' && ws.employee_id);
                const matchesAddress = !this.filterAddress || ws.address_id == this.filterAddress;
                return matchesSearch && matchesVacant && matchesAddress;
            });
        },
        
        openAddModal() { this.formData = { inventory_number: '', model: '', serial_number: '', employee_id: null, status: 'active', address_id: null }; this.showAddModal = true; },
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
        filterType: '',
        showAddModal: false,
        showEditModal: false,
        formData: { address: '', description: '', type: 'cabinet' },
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
                const matchesSearch = addr.address.toLowerCase().includes(s) || (addr.description && addr.description.toLowerCase().includes(s));
                const matchesType = !this.filterType || addr.type === this.filterType;
                return matchesSearch && matchesType;
            });
        },
        
        openAddModal() { this.formData = { address: '', description: '', type: 'cabinet' }; this.showAddModal = true; },
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
        filterEnabled: '',
        filterAddress: '',
        showAddModal: false,
        showEditModal: false,
        formData: { hostname: '', ip_address: '', mac_address: '', os_type: '', location: '', enabled: true, address_id: null },
        hosts: [],
        addresses: [],
        employees: [],
        
        init() { this.loadHosts(); this.loadAddresses(); },
        
        loadHosts() {
            fetch('/api/hosts')
                .then(r => r.json())
                .then(data => { this.hosts = data || []; })
                .catch(err => console.error('Ошибка загрузки хостов:', err));
        },

        loadAddresses() {
            fetch('/api/addresses')
                .then(r => r.json())
                .then(data => { this.addresses = data || []; })
                .catch(err => console.error('Ошибка загрузки адресов:', err));
        },
        
        get filteredHosts() {
            return this.hosts.filter(host => {
                const s = this.searchQuery.toLowerCase();
                const matchesSearch = host.hostname.toLowerCase().includes(s) || host.ip_address.includes(this.searchQuery) || (host.mac_address && host.mac_address.toLowerCase().includes(s));
                const matchesEnabled = !this.filterEnabled || (this.filterEnabled === 'true' && host.enabled) || (this.filterEnabled === 'false' && !host.enabled);
                const matchesAddress = !this.filterAddress || host.address_id == this.filterAddress;
                return matchesSearch && matchesEnabled && matchesAddress;
            });
        },
        
        openAddModal() { this.formData = { hostname: '', ip_address: '', mac_address: '', os_type: '', location: '', enabled: true, address_id: null }; this.showAddModal = true; },
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
        filterCategory: '',
        filterAddress: '',
        showAddModal: false,
        showEditModal: false,
        formData: { name: '', model: '', serial_number: '', ip_address: '', location: '', type: 'switch', address_id: null },
        equipment: [],
        addresses: [],
        
        init() { this.loadEquipment(); this.loadAddresses(); },
        
        loadEquipment() {
            fetch('/api/network-equipment')
                .then(r => r.json())
                .then(data => { this.equipment = data || []; })
                .catch(err => console.error('Ошибка загрузки оборудования:', err));
        },

        loadAddresses() {
            fetch('/api/addresses')
                .then(r => r.json())
                .then(data => { this.addresses = data || []; })
                .catch(err => console.error('Ошибка загрузки адресов:', err));
        },
        
        get filteredEquipment() {
            return this.equipment.filter(eq => {
                const s = this.searchQuery.toLowerCase();
                const matchesSearch = eq.name.toLowerCase().includes(s) || (eq.model && eq.model.toLowerCase().includes(s)) || eq.ip_address.includes(this.searchQuery);
                const matchesCategory = !this.filterCategory || eq.type === this.filterCategory;
                const matchesAddress = !this.filterAddress || eq.address_id == this.filterAddress;
                return matchesSearch && matchesCategory && matchesAddress;
            });
        },
        
        openAddModal() { this.formData = { name: '', model: '', serial_number: '', ip_address: '', location: '', type: 'switch', address_id: null }; this.showAddModal = true; },
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
        filterAddress: '',
        showAddModal: false,
        showEditModal: false,
        formData: { name: '', model: '', serial_number: '', ip_address: '', location: '', address_id: null },
        mfps: [],
        addresses: [],
        
        init() { this.loadMfps(); this.loadAddresses(); },
        
        loadMfps() {
            fetch('/api/network-mfps')
                .then(r => r.json())
                .then(data => { this.mfps = data || []; })
                .catch(err => console.error('Ошибка загрузки МФУ:', err));
        },

        loadAddresses() {
            fetch('/api/addresses')
                .then(r => r.json())
                .then(data => { this.addresses = data || []; })
                .catch(err => console.error('Ошибка загрузки адресов:', err));
        },
        
        get filteredMfps() {
            return this.mfps.filter(mfp => {
                const s = this.searchQuery.toLowerCase();
                const matchesSearch = mfp.name.toLowerCase().includes(s) || (mfp.model && mfp.model.toLowerCase().includes(s)) || mfp.ip_address.includes(this.searchQuery);
                const matchesAddress = !this.filterAddress || mfp.address_id == this.filterAddress;
                return matchesSearch && matchesAddress;
            });
        },
        
        openAddModal() { this.formData = { name: '', model: '', serial_number: '', ip_address: '', location: '', address_id: null }; this.showAddModal = true; },
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
        filterAddress: '',
        showAddModal: false,
        showEditModal: false,
        formData: { name: '', model: '', serial_number: '', mac_address: '', extension: '', location: '', address_id: null },
        phones: [],
        addresses: [],
        
        init() { this.loadPhones(); this.loadAddresses(); },
        
        loadPhones() {
            fetch('/api/ip-phones')
                .then(r => r.json())
                .then(data => { this.phones = data || []; })
                .catch(err => console.error('Ошибка загрузки телефонов:', err));
        },

        loadAddresses() {
            fetch('/api/addresses')
                .then(r => r.json())
                .then(data => { this.addresses = data || []; })
                .catch(err => console.error('Ошибка загрузки адресов:', err));
        },
        
        get filteredPhones() {
            return this.phones.filter(phone => {
                const s = this.searchQuery.toLowerCase();
                const matchesSearch = phone.name.toLowerCase().includes(s) || (phone.model && phone.model.toLowerCase().includes(s)) || (phone.extension && phone.extension.includes(this.searchQuery));
                const matchesAddress = !this.filterAddress || phone.address_id == this.filterAddress;
                return matchesSearch && matchesAddress;
            });
        },
        
        openAddModal() { this.formData = { name: '', model: '', serial_number: '', mac_address: '', extension: '', location: '', address_id: null }; this.showAddModal = true; },
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
