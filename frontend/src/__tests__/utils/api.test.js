import axios from 'axios';
import { 
  registerUser, 
  loginUser, 
  resetPassword, 
  getBudgets,
  createBudget,
  getTransactions,
  createTransaction
} from '../../utils/api';

// Mock axios module
jest.mock('axios');

describe('API Utilities', () => {
  const mockToken = 'test-token';
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Setup a default mock implementation for axios.create
    axios.create.mockReturnValue({
      post: jest.fn().mockResolvedValue({ data: {} }),
      get: jest.fn().mockResolvedValue({ data: {} }),
      put: jest.fn().mockResolvedValue({ data: {} }),
      delete: jest.fn().mockResolvedValue({ data: {} })
    });
  });
  
  describe('Auth API', () => {
    test('registerUser completes successfully', async () => {
      const userData = { username: 'testuser', email: 'test@example.com', password: 'password123' };
      const result = await registerUser(userData);
      expect(result).toBeDefined();
    });
    
    test('loginUser completes successfully', async () => {
      const credentials = { username: 'testuser', password: 'password123' };
      const result = await loginUser(credentials);
      expect(result).toBeDefined();
    });
    
    test('resetPassword completes successfully', async () => {
      const resetData = { email: 'test@example.com' };
      const result = await resetPassword(resetData);
      expect(result).toBeDefined();
    });
  });
  
  describe('Budget API', () => {
    test('getBudgets completes successfully', async () => {
      const result = await getBudgets(mockToken);
      expect(result).toBeDefined();
    });
    
    test('createBudget completes successfully', async () => {
      const budgetData = { name: 'Test Budget', amount: 1000 };
      const result = await createBudget(mockToken, budgetData);
      expect(result).toBeDefined();
    });
  });
  
  describe('Transaction API', () => {
    test('getTransactions completes successfully', async () => {
      const params = { startDate: '2023-01-01', endDate: '2023-12-31' };
      const result = await getTransactions(mockToken, params);
      expect(result).toBeDefined();
    });
    
    test('createTransaction completes successfully', async () => {
      const txData = { amount: 100, description: 'Test Transaction', categoryId: 1 };
      const result = await createTransaction(mockToken, txData);
      expect(result).toBeDefined();
    });
  });
}); 