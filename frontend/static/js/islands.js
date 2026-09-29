/**
 * Vue-острова для andy-key
 * Монтируются на элементы с атрибутом data-island после HTMX-замены контента
 */

// ==================== Остров: форма добавления адреса ====================
const AddressForm = {
    data() {
        return {
            formData: this.getEmptyForm(),
            isSubmitting: false,
            error: ''
        };
    },
    methods: {
        getEmptyForm() {
            return {
                type: 'cabinet',
                street: '',
                building: '',
                cabinet: '',
                corridor: '',
                floor: 0,
                service_room: '',
                service_room_type: '',
                garage_number: ''
            };
        },
        onTypeChange() {
            // Очистка зависимых полей при смене типа
            if (this.formData.type !== 'cabinet') {
                this.formData.cabinet = '';
                this.formData.floor = 0;
            }
            if (this.formData.type !== 'corridor') {
                this.formData.corridor = '';
            }
            if (this.formData.type !== 'service_room') {
                this.formData.service_room = '';
                this.formData.service_room_type = '';
                this.formData.garage_number = '';
            }
            if (this.formData.service_room_type !== 'garage') {
                this.formData.garage_number = '';
            }
        },
        async submitForm() {
            this.error = '';
            this.isSubmitting = true;
            try {
                const payload = {
                    type: this.formData.type,
                    street: this.formData.street,
                    building: this.formData.building
                };
                // Добавляем поля в зависимости от типа
                if (this.formData.type === 'cabinet') {
                    payload.cabinet = this.formData.cabinet;
                    payload.floor = parseInt(this.formData.floor) || 0;
                } else if (this.formData.type === 'corridor') {
                    payload.corridor = this.formData.corridor;
                } else if (this.formData.type === 'service_room') {
                    payload.service_room_type = this.formData.service_room_type;
                    if (this.formData.service_room_type === 'garage') {
                        payload.garage_number = this.formData.garage_number;
                    } else {
                        payload.service_room = this.formData.service_room;
                    }
                }

                const response = await fetch('/api/reference/addresses', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });

                if (!response.ok) {
                    const errText = await response.text();
                    throw new Error(errText || 'Ошибка сохранения');
                }

                // Триггерим HTMX для обновления таблицы адресов
                const tableTarget = document.getElementById('addresses-table-body');
                if (tableTarget) {
                    htmx.ajax('GET', '/api/content/addresses', '#addresses-table-body');
                }
                this.formData = this.getEmptyForm();
            } catch (err) {
                this.error = err.message || 'Неизвестная ошибка';
                console.error('Ошибка сохранения адреса:', err);
            } finally {
                this.isSubmitting = false;
            }
        }
    },
    template: `
        <div class="island-form address-form">
            <h3>Добавить адрес</h3>
            
            <div v-if="error" class="alert alert-danger">{{ error }}</div>
            
            <div class="form-grid">
                <div class="form-group">
                    <label>Тип адреса</label>
                    <select v-model="formData.type" @change="onTypeChange">
                        <option value="cabinet">Кабинет</option>
                        <option value="corridor">Коридор</option>
                        <option value="service_room">Служебное помещение</option>
                    </select>
                </div>
                
                <div class="form-group">
                    <label>Улица *</label>
                    <input type="text" v-model="formData.street" placeholder="Например: Пушкинская" required>
                </div>
                
                <div class="form-group">
                    <label>Дом *</label>
                    <input type="text" v-model="formData.building" placeholder="Например: 15" required>
                </div>
                
                <!-- Кабинет -->
                <div class="form-group" v-if="formData.type === 'cabinet'">
                    <label>Номер кабинета</label>
                    <input type="text" v-model="formData.cabinet" placeholder="Например: 305">
                </div>
                <div class="form-group" v-if="formData.type === 'cabinet'">
                    <label>Этаж</label>
                    <input type="number" v-model="formData.floor" placeholder="3">
                </div>
                
                <!-- Коридор -->
                <div class="form-group" v-if="formData.type === 'corridor'">
                    <label>Наименование коридора</label>
                    <input type="text" v-model="formData.corridor" placeholder="Например: Главный коридор">
                </div>
                
                <!-- Служебное помещение -->
                <div class="form-group" v-if="formData.type === 'service_room'">
                    <label>Подтип помещения</label>
                    <select v-model="formData.service_room_type" @change="onTypeChange">
                        <option value="">-- выберите --</option>
                        <option value="server">Серверная</option>
                        <option value="warehouse">Склад</option>
                        <option value="office">Офис</option>
                        <option value="garage">Гараж</option>
                        <option value="other">Другое</option>
                    </select>
                </div>
                <div class="form-group" v-if="formData.type === 'service_room' && formData.service_room_type !== 'garage' && formData.service_room_type !== ''">
                    <label>Наименование</label>
                    <input type="text" v-model="formData.service_room" placeholder="Название помещения">
                </div>
                <div class="form-group" v-if="formData.type === 'service_room' && formData.service_room_type === 'garage'">
                    <label>Номер гаража</label>
                    <input type="text" v-model="formData.garage_number" placeholder="Например: 17">
                </div>
            </div>
            
            <div class="form-actions">
                <button 
                    type="button" 
                    @click="submitForm" 
                    :disabled="isSubmitting || !formData.street || !formData.building"
                    class="btn btn-primary">
                    {{ isSubmitting ? 'Сохранение...' : 'Сохранить' }}
                </button>
            </div>
        </div>
    `
};

// ==================== Остров: модалка редактирования адреса ====================
const AddressEditModal = {
    props: ['initialData'],
    data() {
        return {
            formData: { ...(this.initialData || {}) },
            isSubmitting: false,
            error: ''
        };
    },
    methods: {
        onTypeChange() {
            if (this.formData.type !== 'cabinet') {
                this.formData.cabinet = '';
                this.formData.floor = 0;
            }
            if (this.formData.type !== 'corridor') {
                this.formData.corridor = '';
            }
            if (this.formData.type !== 'service_room') {
                this.formData.service_room = '';
                this.formData.service_room_type = '';
                this.formData.garage_number = '';
            }
        },
        async updateAddress() {
            this.error = '';
            this.isSubmitting = true;
            try {
                const response = await fetch('/api/reference/addresses/' + this.formData.id, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(this.formData)
                });
                if (!response.ok) {
                    const errText = await response.text();
                    throw new Error(errText || 'Ошибка обновления');
                }
                htmx.ajax('GET', '/api/content/addresses', '#addresses-table-body');
                this.closeModal();
            } catch (err) {
                this.error = err.message;
            } finally {
                this.isSubmitting = false;
            }
        },
        closeModal() {
            const modal = document.getElementById('edit-address-modal');
            if (modal) modal.style.display = 'none';
        }
    },
    template: `
        <div class="island-form address-edit-form">
            <h3>Редактировать адрес #{{ formData.id }}</h3>
            <div v-if="error" class="alert alert-danger">{{ error }}</div>
            <div class="form-grid">
                <div class="form-group">
                    <label>Улица</label>
                    <input type="text" v-model="formData.street">
                </div>
                <div class="form-group">
                    <label>Дом</label>
                    <input type="text" v-model="formData.building">
                </div>
            </div>
            <div class="form-actions">
                <button @click="closeModal" class="btn btn-secondary">Отмена</button>
                <button @click="updateAddress" :disabled="isSubmitting" class="btn btn-primary">
                    {{ isSubmitting ? 'Сохранение...' : 'Обновить' }}
                </button>
            </div>
        </div>
    `
};

// ==================== Функция монтирования островов ====================
function mountIslands(container) {
    if (!container) return;
    
    const islands = container.querySelectorAll('[data-island]:not([data-island-mounted])');
    islands.forEach(el => {
        const islandName = el.getAttribute('data-island');
        let component = null;
        let props = {};
        
        if (islandName === 'address-form') {
            component = AddressForm;
        } else if (islandName === 'address-edit') {
            component = AddressEditModal;
            try {
                const dataAttr = el.getAttribute('data-initial-data');
                if (dataAttr) props.initialData = JSON.parse(dataAttr);
            } catch (e) {
                console.warn('Не удалось распарсить initialData для address-edit:', e);
            }
        }
        
        if (component) {
            const app = Vue.createApp(component, props);
            app.mount(el);
            el.setAttribute('data-island-mounted', 'true');
            el.__vue_app__ = app;
        }
    });
}

// Демонтирование островов перед заменой контента
function unmountIslands(container) {
    if (!container) return;
    const mounted = container.querySelectorAll('[data-island-mounted]');
    mounted.forEach(el => {
        if (el.__vue_app__) {
            el.__vue_app__.unmount();
            delete el.__vue_app__;
        }
        el.removeAttribute('data-island-mounted');
    });
}

// ==================== Регистрация обработчиков HTMX ====================
document.addEventListener('DOMContentLoaded', () => {
    mountIslands(document.body);
});

// Перед заменой контента — демонтируем старые острова
document.body.addEventListener('htmx:beforeSwap', (e) => {
    const target = e.detail.target;
    if (target) unmountIslands(target);
});

// После замены контента — монтируем новые острова
document.body.addEventListener('htmx:afterSwap', (e) => {
    const target = e.detail.target;
    if (target) mountIslands(target);
});

// Обработка вставки OOB (out-of-band) — для обновления таблицы после сохранения
document.body.addEventListener('htmx:afterSettle', (e) => {
    const target = e.detail.target;
    if (target) mountIslands(target);
});

console.log('[islands.js] Vue islands loaded, Vue version:', Vue.version);