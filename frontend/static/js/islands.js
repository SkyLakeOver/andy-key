/**
 * Vue-острова для andy-key (Фаза 1, раздел «Адреса»)
 *
 * Контракт данных (единый для всех слоёв, все поля — СТРОКИ, включая floor):
 *   street, building, type (cabinet|corridor|service_room), cabinet, corridor,
 *   floor, service_room, service_room_type (server|warehouse|office|garage|other), garage_number
 *
 * Island-эндпоинты:
 *   POST   /api/reference/addresses/island        — создание
 *   PUT    /api/reference/addresses/island/{id}   — обновление
 *   DELETE /api/reference/addresses/island/{id}   — удаление
 *   GET    /api/reference/addresses/edit-form/{id}— HTML-фрагмент модалки редактирования
 *
 * Острова монтируются на элементы [data-island] после HTMX-замены контента.
 */

// ==================== Общий mixin для форм адреса ====================
const AddressFormMixin = {
    data() {
        return {
            formData: this.emptyForm(),
            isSubmitting: false,
            error: ''
        };
    },
    methods: {
        // Пустая форма: ВСЕ поля — строки, floor === '' (не 0!)
        emptyForm() {
            return {
                type: 'cabinet',
                street: '',
                building: '',
                cabinet: '',
                corridor: '',
                floor: '',
                service_room: '',
                service_room_type: '',
                garage_number: ''
            };
        },
        // Сброс полей, не относящихся к выбранному типу
        onTypeChange() {
            if (this.formData.type !== 'cabinet') {
                this.formData.cabinet = '';
                this.formData.floor = '';
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
        // Сборка payload по контракту: floor ВСЕГДА строка
        buildPayload() {
            const payload = {
                street: String(this.formData.street || ''),
                building: String(this.formData.building || ''),
                type: this.formData.type || 'cabinet'
            };
            if (payload.type === 'cabinet') {
                payload.cabinet = String(this.formData.cabinet || '');
                payload.floor = String(this.formData.floor || '');
            } else if (payload.type === 'corridor') {
                payload.corridor = String(this.formData.corridor || '');
            } else if (payload.type === 'service_room') {
                payload.service_room_type = String(this.formData.service_room_type || '');
                if (payload.service_room_type === 'garage') {
                    payload.garage_number = String(this.formData.garage_number || '');
                } else {
                    payload.service_room = String(this.formData.service_room || '');
                }
            }
            return payload;
        },
        readError(errText, fallback) {
            // Сервер возвращает человекочитаемый текст ошибки (текст или JSON)
            try {
                const j = JSON.parse(errText);
                return j.error || j.message || errText || fallback;
            } catch (e) {
                return (errText && errText.trim()) ? errText.trim() : fallback;
            }
        }
    }
};

// ==================== Остров: форма добавления адреса ====================
const AddressForm = {
    mixins: [AddressFormMixin],
    methods: {
        async submitForm() {
            this.error = '';
            this.isSubmitting = true;
            try {
                const response = await fetch('/api/reference/addresses/island', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(this.buildPayload())
                });
                if (!response.ok) {
                    throw new Error(this.readError(await response.text(), 'Ошибка сохранения'));
                }
                this.formData = this.emptyForm();
                window.refreshAddressesTable();
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
                    <input type="number" v-model="formData.floor" placeholder="3" min="0">
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
                <div class="form-group" v-if="formData.type === 'service_room' && formData.service_room_type === 'garage'">
                    <label>Номер гаража</label>
                    <input type="text" v-model="formData.garage_number" placeholder="Например: 17">
                </div>
                <div class="form-group" v-if="formData.type === 'service_room' && formData.service_room_type !== 'garage' && formData.service_room_type !== ''">
                    <label>Наименование</label>
                    <input type="text" v-model="formData.service_room" placeholder="Название помещения">
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
    mixins: [AddressFormMixin],
    props: ['initialData'],
    data() {
        // Vue мержит data из mixin и здесь; addressId дополняет состояние
        return {
            addressId: ''
        };
    },
    created() {
        // Предзаполнение из data-initial-data (все поля — строки)
        const d = this.initialData || {};
        const base = this.emptyForm();
        Object.keys(base).forEach(k => {
            if (d[k] !== undefined && d[k] !== null) {
                base[k] = String(d[k]);
            }
        });
        this.formData = base;
        this.addressId = d.id ? String(d.id) : '';
    },
    methods: {
        async updateAddress() {
            this.error = '';
            this.isSubmitting = true;
            try {
                const response = await fetch('/api/reference/addresses/island/' + this.addressId, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(this.buildPayload())
                });
                if (!response.ok) {
                    throw new Error(this.readError(await response.text(), 'Ошибка обновления'));
                }
                this.closeModal();
                window.refreshAddressesTable();
            } catch (err) {
                this.error = err.message || 'Неизвестная ошибка';
                console.error('Ошибка обновления адреса:', err);
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
            <h3>Редактировать адрес #{{ addressId }}</h3>

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
                    <input type="text" v-model="formData.street" placeholder="Например: Пушкинская">
                </div>

                <div class="form-group">
                    <label>Дом *</label>
                    <input type="text" v-model="formData.building" placeholder="Например: 15">
                </div>

                <!-- Кабинет -->
                <div class="form-group" v-if="formData.type === 'cabinet'">
                    <label>Номер кабинета</label>
                    <input type="text" v-model="formData.cabinet" placeholder="Например: 305">
                </div>
                <div class="form-group" v-if="formData.type === 'cabinet'">
                    <label>Этаж</label>
                    <input type="number" v-model="formData.floor" placeholder="3" min="0">
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
                <div class="form-group" v-if="formData.type === 'service_room' && formData.service_room_type === 'garage'">
                    <label>Номер гаража</label>
                    <input type="text" v-model="formData.garage_number" placeholder="Например: 17">
                </div>
                <div class="form-group" v-if="formData.type === 'service_room' && formData.service_room_type !== 'garage' && formData.service_room_type !== ''">
                    <label>Наименование</label>
                    <input type="text" v-model="formData.service_room" placeholder="Название помещения">
                </div>
            </div>

            <div class="form-actions">
                <button type="button" @click="closeModal" class="btn btn-secondary">Отмена</button>
                <button
                    type="button"
                    @click="updateAddress"
                    :disabled="isSubmitting || !formData.street || !formData.building"
                    class="btn btn-primary">
                    {{ isSubmitting ? 'Сохранение...' : 'Обновить' }}
                </button>
            </div>
        </div>
    `
};

// ==================== Обновление SSR-таблицы адресов ====================
// Глобальная функция: вызывается из hx-on::after-request в шаблонах и из островов.
// Перечитывает /api/content/addresses с ТЕКУЩИМИ значениями фильтров и заменяет
// tbody через select + outerHTML (без select в tbody попал бы весь ответ секции).
window.refreshAddressesTable = function () {
    const searchEl = document.querySelector('input[name="search"]');
    const typeEl = document.querySelector('select[name="type_filter"]');
    const params = new URLSearchParams();
    if (searchEl && searchEl.value) params.set('search', searchEl.value);
    if (typeEl && typeEl.value) params.set('type_filter', typeEl.value);
    const qs = params.toString();
    const url = '/api/content/addresses' + (qs ? '?' + qs : '');

    htmx.ajax('GET', url, {
        target: '#addresses-table-body',
        swap: 'outerHTML',
        select: '#addresses-table-body'
    });
};

// ==================== Функции монтирования/демонтирования островов ==========
function mountIslands(container) {
    if (!container || !container.querySelectorAll) return;

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
            el.__vue_app__ = app; // сохраняем экземпляр для последующего unmount
        }
    });
}

function unmountIslands(container) {
    if (!container || !container.querySelectorAll) return;
    const mounted = container.querySelectorAll('[data-island-mounted]');
    mounted.forEach(el => {
        if (el.__vue_app__) {
            el.__vue_app__.unmount();
            delete el.__vue_app__;
        }
        el.removeAttribute('data-island-mounted');
    });
}

// ==================== Регистрация обработчиков ==============================
document.addEventListener('DOMContentLoaded', () => {
    mountIslands(document.body);
});

// Перед заменой контента — демонтируем острова внутри целевого элемента
document.body.addEventListener('htmx:beforeSwap', (e) => {
    const target = e.detail && e.detail.target;
    if (target) unmountIslands(target);
});

// После замены контента — монтируем новые острова
document.body.addEventListener('htmx:afterSwap', (e) => {
    const target = e.detail && e.detail.target;
    if (target) mountIslands(target);
});

// Дополнительная страховка для OOB-вставок
document.body.addEventListener('htmx:afterSettle', () => {
    mountIslands(document.body);
});

console.log('[islands.js] Vue islands loaded, Vue version:', Vue.version);
