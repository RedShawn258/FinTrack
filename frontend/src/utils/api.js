// Use the axios instance with auto-refresh interceptor
import apiClient from './axios';

// ========== Auth Endpoints ==========

export const registerUser = async (userData) => {
  return await apiClient.post("/auth/register", userData);
};

// Note: loginUser is now handled by authService.js and Redux
// This is kept for backward compatibility if needed
export const loginUser = async (credentials) => {
  return await apiClient.post("/auth/login", credentials);
};

export const resetPassword = async (data) => {
  return await apiClient.post("/auth/forgot-password", { email: data.email });
};

export const resetPasswordWithToken = async (data) => {
  return await apiClient.post("/auth/reset-password", {
    token: data.token,
    newPassword: data.password,
    confirmPassword: data.password
  });
};

export const fetchProfile = async () => {
  // Token is automatically attached by axios interceptor
  return await apiClient.get("/profile");
};

export const updateProfile = async (profileData) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.put("/profile", profileData);
};

// ========== Budget Endpoints ==========

export const getBudgets = async () => {
  // Token is automatically attached by axios interceptor
  return await apiClient.get("/budgets");
};

export const createBudget = async (budgetData) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.post("/budgets", budgetData);
};

export const updateBudget = async (budgetId, budgetData) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.put(`/budgets/${budgetId}`, budgetData);
};

export const deleteBudget = async (budgetId) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.delete(`/budgets/${budgetId}`);
};

// ========== Category Endpoints ==========

export const getCategories = async () => {
  // Token is automatically attached by axios interceptor
  return await apiClient.get("/categories");
};

export const createCategory = async (categoryData) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.post("/categories", categoryData);
};

export const deleteCategory = async (categoryId) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.delete(`/categories/${categoryId}`);
};

// ========== Transaction Endpoints ==========

export const getTransactions = async (params = {}) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.get("/transactions", { params });
};

export const createTransaction = async (txData) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.post("/transactions", txData);
};

export const updateTransaction = async (txId, txData) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.put(`/transactions/${txId}`, txData);
};

export const deleteTransaction = async (txId) => {
  // Token is automatically attached by axios interceptor
  return await apiClient.delete(`/transactions/${txId}`);
};
