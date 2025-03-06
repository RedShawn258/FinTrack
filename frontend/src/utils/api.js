import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || "http://localhost:8080/api/v1";

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// ========== Auth Endpoints ==========

export const registerUser = async (userData) => {
  return await api.post("/auth/register", userData);
};

export const loginUser = async (credentials) => {
  return await api.post("/auth/login", credentials);
};

export const resetPassword = async (data) => {
  return await api.post("/auth/reset-password", data);
};

export const fetchProfile = async (token) => {
  return await api.get("/profile", {
    headers: { Authorization: `Bearer ${token}` },
  });
};

// ========== Budget Endpoints ==========

export const getBudgets = async (token) => {
  return await api.get("/budgets", {
    headers: { Authorization: `Bearer ${token}` },
  });
};

export const createBudget = async (token, budgetData) => {
  return await api.post("/budgets", budgetData, {
    headers: { Authorization: `Bearer ${token}` },
  });
};

export const updateBudget = async (token, budgetId, budgetData) => {
  return await api.put(`/budgets/${budgetId}`, budgetData, {
    headers: { Authorization: `Bearer ${token}` },
  });
};

export const deleteBudget = async (token, budgetId) => {
  return await api.delete(`/budgets/${budgetId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
};

// ========== Category Endpoints ==========

export const getCategories = async (token) => {
  return await api.get("/categories", {
    headers: { Authorization: `Bearer ${token}` },
  });
};

export const createCategory = async (token, categoryData) => {
  return await api.post("/categories", categoryData, {
    headers: { Authorization: `Bearer ${token}` },
  });
};

export const deleteCategory = async (token, categoryId) => {
  return await api.delete(`/categories/${categoryId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
};


// ========== Transaction Endpoints ==========

export const getTransactions = async (token, params = {}) => {
  return await api.get("/transactions", {
    headers: { Authorization: `Bearer ${token}` },
    params,
  });
};

export const createTransaction = async (token, txData) => {
  return await api.post("/transactions", txData, {
    headers: { Authorization: `Bearer ${token}` },
  });
};

export const updateTransaction = async (token, txId, txData) => {
  return await api.put(`/transactions/${txId}`, txData, {
    headers: { Authorization: `Bearer ${token}` },
  });
};

export const deleteTransaction = async (token, txId) => {
  return await api.delete(`/transactions/${txId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
};
