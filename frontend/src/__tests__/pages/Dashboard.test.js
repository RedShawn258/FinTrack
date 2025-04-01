import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import Dashboard from '../../pages/Dashboard';
import { AuthContext } from '../../context/AuthContext';
import * as api from '../../utils/api';

// Mock the API module
jest.mock('../../utils/api');

// Mock Chart.js
jest.mock('chart.js', () => ({
  Chart: {
    register: jest.fn(),
  },
  CategoryScale: jest.fn(),
  LinearScale: jest.fn(),
  BarElement: jest.fn(),
  Title: jest.fn(),
  Tooltip: jest.fn(),
  Legend: jest.fn(),
}));

// Mock react-chartjs-2
jest.mock('react-chartjs-2', () => ({
  Pie: () => null
}));

const mockUser = {
  id: 1,
  username: 'Mahesh',
  email: 'durgamaheshboppani@gmail.com',
  balance: 1000,
  token: 'test-token'
};

const renderDashboard = () => {
  const mockLogout = jest.fn();
  const mockRefreshDashboardData = jest.fn();
  return render(
    <BrowserRouter>
      <AuthContext.Provider value={{ 
        user: mockUser, 
        isAuthenticated: true, 
        logout: mockLogout,
        dashboardData: {
          budgets: [],
          categories: [],
          transactions: []
        },
        refreshDashboardData: mockRefreshDashboardData
      }}>
        <Dashboard />
      </AuthContext.Provider>
    </BrowserRouter>
  );
};

describe('Dashboard Component', () => {
  beforeAll(() => {
    window.alert = jest.fn();
  });

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock API responses
    api.fetchProfile.mockResolvedValue({ data: { user: mockUser } });
    api.getTransactions.mockResolvedValue({ data: { transactions: [] } });
    api.getCategories.mockResolvedValue({ data: { categories: [] } });
    api.getBudgets.mockResolvedValue({ data: { budgets: [] } });
  });

  test('handles API errors gracefully', async () => {
    // Mock the refreshDashboardData to throw an error
    const mockRefreshDashboardData = jest.fn().mockImplementation(() => {
      return Promise.reject(new Error('Failed to fetch dashboard data'));
    });
    
    render(
      <BrowserRouter>
        <AuthContext.Provider value={{ 
          user: mockUser, 
          isAuthenticated: true, 
          logout: jest.fn(),
          dashboardData: {
            budgets: [],
            categories: [],
            transactions: []
          },
          refreshDashboardData: mockRefreshDashboardData
        }}>
          <Dashboard />
        </AuthContext.Provider>
      </BrowserRouter>
    );

    // Wait for error message
    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith('Error fetching data');
    });
  });
});

describe('Dashboard Component - Basic Setup', () => {
  test('renders dashboard with user information', async () => {
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText((content) => {
        return content.includes('Total Expenses: $');
      })).toBeInTheDocument();
      expect(screen.getByText('Add New Expense')).toBeInTheDocument();
      expect(screen.getByText('Budget Overview')).toBeInTheDocument();
    });
  });
});

describe('Dashboard Component - UI Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Setup default API responses for UI tests
    api.getTransactions.mockResolvedValue({ data: { transactions: [] } });
    api.getCategories.mockResolvedValue({ data: { categories: [] } });
    api.getBudgets.mockResolvedValue({ data: { budgets: [] } });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  test('renders total expenses', () => {
    renderDashboard();
    expect(screen.getByText(/Total Expenses: \$/)).toBeInTheDocument();
  });

  test('renders logout button', () => {
    renderDashboard();
    expect(screen.getByTitle('Logout')).toBeInTheDocument();
  });

  test('renders expense form fields', () => {
    renderDashboard();
    expect(screen.getByText('Description')).toBeInTheDocument();
    expect(screen.getByText('Amount')).toBeInTheDocument();
    expect(screen.getAllByText('Category')[0]).toBeInTheDocument();
    expect(screen.getByText('Date')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add Expense' })).toBeInTheDocument();
  });

  test('renders budget form fields', () => {
    renderDashboard();
    expect(screen.getByText('Set New Budget')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Select or type category')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter amount')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Set Budget' })).toBeInTheDocument();
  });

  test('renders date inputs in budget form', () => {
    renderDashboard();
    const startDateLabel = screen.getByText('Start Date');
    const endDateLabel = screen.getByText('End Date');
    expect(startDateLabel).toBeInTheDocument();
    expect(endDateLabel).toBeInTheDocument();
  });

  test('renders category datalist', () => {
    renderDashboard();
    expect(screen.getAllByRole('combobox')[0]).toBeInTheDocument();
    expect(document.getElementById('categoryList')).toBeInTheDocument();
  });

  test('renders empty state messages', () => {
    renderDashboard();
    expect(screen.getByText('No budgets found')).toBeInTheDocument();
    expect(screen.getByText('No transactions yet')).toBeInTheDocument();
    expect(screen.getByText('No recent expenses')).toBeInTheDocument();
  });

  test('renders expense distribution section', () => {
    renderDashboard();
    expect(screen.getByText('Expense Distribution')).toBeInTheDocument();
  });

  test('renders recent expenses section', () => {
    renderDashboard();
    expect(screen.getByText('Recent Expenses')).toBeInTheDocument();
  });
});