import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || "http://localhost:8080/api/v1";

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        "Content-Type": "application/json",
    },
});

// Register User
export const registerUser = async (userData) => {
    return await api.post("/auth/register", userData);
};

// Login User
export const loginUser = async (credentials) => {
    return await api.post("/auth/login", credentials);
};

// Reset Password
export const resetPassword = async (data) => {
    return await api.post("/auth/reset-password", data);
};

// Fetch Profile (Authenticated Route)
export const fetchProfile = async (token) => {
    return await api.get("/profile", {
        headers: { Authorization: `Bearer ${token}` },
    });
};
