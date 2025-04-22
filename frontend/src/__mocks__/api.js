// Mock API functions for testing
export const fetchProfile = jest.fn().mockResolvedValue({
  data: {
    username: 'Test User',
    email: 'test@example.com',
    firstName: 'Test',
    lastName: 'User',
    phoneNumber: '1234567890',
    currency: 'USD',
    notificationsEnabled: true,
    theme: 'light'
  }
});

export const updateProfile = jest.fn().mockResolvedValue({ 
  success: true 
});

export const registerUser = jest.fn().mockResolvedValue({
  data: { 
    user: { id: 1, username: 'testuser' },
    token: 'test-token'
  }
});

export const loginUser = jest.fn().mockResolvedValue({
  data: { 
    user: { id: 1, username: 'testuser' },
    token: 'test-token'
  }
});

export const resetPassword = jest.fn().mockResolvedValue({
  data: { message: 'Password reset email sent' }
});

export const resetPasswordWithToken = jest.fn().mockResolvedValue({
  data: { message: 'Password reset successful' }
});

export const getBudgets = jest.fn().mockResolvedValue({
  data: [{ id: 1, name: 'Monthly Budget', amount: 1000 }]
});

export const createBudget = jest.fn().mockResolvedValue({
  data: { id: 2, name: 'New Budget', amount: 500 }
});

export const updateBudget = jest.fn().mockResolvedValue({
  data: { id: 1, name: 'Updated Budget', amount: 1500 }
});

export const deleteBudget = jest.fn().mockResolvedValue({
  success: true
});

export const getCategories = jest.fn().mockResolvedValue({
  data: [{ id: 1, name: 'Food' }, { id: 2, name: 'Housing' }]
});

export const createCategory = jest.fn().mockResolvedValue({
  data: { id: 3, name: 'Entertainment' }
});

export const deleteCategory = jest.fn().mockResolvedValue({
  success: true
});

export const getTransactions = jest.fn().mockResolvedValue({
  data: [
    { id: 1, amount: 50, category: 'Food', date: '2023-01-01' },
    { id: 2, amount: 1000, category: 'Housing', date: '2023-01-02' }
  ]
});

export const createTransaction = jest.fn().mockResolvedValue({
  data: { id: 3, amount: 25, category: 'Entertainment', date: '2023-01-03' }
});

export const updateTransaction = jest.fn().mockResolvedValue({
  data: { id: 1, amount: 75, category: 'Food', date: '2023-01-01' }
});

export const deleteTransaction = jest.fn().mockResolvedValue({
  success: true
});

export const getBudgetsByCategory = jest.fn(); 