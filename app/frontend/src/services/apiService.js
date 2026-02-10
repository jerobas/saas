import axios from 'axios';

const api = axios.create({
    baseURL: 'http://localhost:3000/api', 
    timeout: 10000, 
});

api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response && error.response.status === 401) {
            // Redirecionar para a página de login
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

export const createUser = async (userData) => {
    try {
        const response = await api.post('/create', userData);
        return response.data;
    } catch (error) {
        console.error('Erro ao criar usuário:', error);
        throw error;
    }
};

export const getUser = async (userId) => {
    try {
        const response = await api.get(`/users/${userId}`);
        return response.data;
    } catch (error) {
        console.error('Erro ao buscar usuário:', error);
        throw error;
    }
};

export const updateUser = async (userId, userData) => {
    try {
        const response = await api.put(`/users/${userId}`, userData);
        return response.data;
    } catch (error) {
        console.error('Erro ao atualizar usuário:', error);
        throw error;
    }
};

export const deleteUser = async (userId) => {
    try {
        const response = await api.delete(`/users/${userId}`);
        return response.data;
    } catch (error) {
        console.error('Erro ao deletar usuário:', error);
        throw error;
    }
};

export const checkLicenseStatus = async (userId, email) => {
    try {
        const response = await api.get('/license', {
            params: { userId, email },
        });
        return response.data;
    } catch (error) {
        console.error('Erro ao verificar status da licença:', error);
        throw error;
    }
};

export default api;