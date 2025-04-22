import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import Insights from '../pages/Insights';
import { act } from 'react';

// Mock the chart.js library
jest.mock('react-chartjs-2', () => ({
  Line: () => null,
  Bar: () => null,
  Pie: () => null
}));

// Mock the useTheme hook
jest.mock('../hooks/useTheme', () => ({
  __esModule: true,
  default: () => ({ theme: 'light', setTheme: jest.fn() })
}));

// Mock the API calls
jest.mock('../utils/api', () => ({
  fetchDashboard: jest.fn().mockResolvedValue({
    transactions: [],
    categories: [],
    budgets: []
  })
}));

const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
  BrowserRouter: ({ children }) => <div>{children}</div>
}));

const mockDashboardData = {
  transactions: [
    {
      ID: 1,
      Description: 'Test Transaction',
      Amount: 100,
      CategoryID: 1,
      TransactionDate: '2024-03-01'
    }
  ],
  categories: [
    {
      ID: 1,
      Name: 'Test Category'
    }
  ],
  budgets: [
    {
      ID: 1,
      CategoryID: 1,
      LimitAmount: 500,
      StartDate: '2024-03-01',
      EndDate: '2024-03-31'
    }
  ]
};

const mockAuthContext = {
  user: { token: 'test-token' },
  dashboardData: mockDashboardData,
  refreshDashboardData: jest.fn().mockResolvedValue(mockDashboardData),
  logout: jest.fn()
};

describe('Insights Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    document.body.className = '';
    document.body.removeAttribute('data-theme');
  });

  const renderInsights = async (contextValue = mockAuthContext) => {
    let utils;
    await act(async () => {
      utils = render(
        <BrowserRouter>
          <AuthContext.Provider value={contextValue}>
            <Insights />
          </AuthContext.Provider>
        </BrowserRouter>
      );
      await new Promise(resolve => setTimeout(resolve, 0));
    });
    return utils;
  };

  test('renders insights page with header', async () => {
    await renderInsights();
    expect(await screen.findByText('Insights & Achievements')).toBeInTheDocument();
  });

  test('switches between analytics and achievements tabs', async () => {
    await renderInsights();
    
    // Initially shows analytics
    expect(await screen.findByText('Spending Trends')).toBeInTheDocument();
    
    // Switch to achievements tab
    await act(async () => {
      fireEvent.click(screen.getByText('Achievements'));
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    // Wait for achievements content to appear
    const achievementsElement = await screen.findByText('More Achievements Coming Soon!');
    expect(achievementsElement).toBeInTheDocument();
  });

  test('displays loading state when isLoading is true', async () => {
    const slowContext = {
      ...mockAuthContext,
      dashboardData: null,
      refreshDashboardData: jest.fn().mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve(mockDashboardData), 100))
      )
    };

    await act(async () => {
      render(
        <BrowserRouter>
          <AuthContext.Provider value={slowContext}>
            <Insights />
          </AuthContext.Provider>
        </BrowserRouter>
      );
    });

    expect(await screen.findByText('Loading insights data...')).toBeInTheDocument();
  });

  test('navigates back to dashboard when back button is clicked', async () => {
    await renderInsights();
    
    await act(async () => {
      fireEvent.click(screen.getByTitle('Back to Dashboard'));
    });

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });

  test('logs out when logout button is clicked', async () => {
    await renderInsights();
    
    await act(async () => {
      fireEvent.click(screen.getByTitle('Logout'));
    });

    expect(mockAuthContext.logout).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  test('displays no data message when there are no transactions', async () => {
    const noDataContext = {
      ...mockAuthContext,
      dashboardData: { ...mockDashboardData, transactions: [] }
    };

    await renderInsights(noDataContext);

    // Find the no-data message by its text content
    expect(await screen.findByText('No category data available')).toBeInTheDocument();
    expect(await screen.findByText('Categorize your expenses to see this breakdown')).toBeInTheDocument();
  });

  test('displays achievements when data is available', async () => {
    await renderInsights();
    
    await act(async () => {
      fireEvent.click(screen.getByText('Achievements'));
      await new Promise(resolve => setTimeout(resolve, 0));
    });
    
    // Wait for achievements to load
    const firstStepsAchievement = await screen.findByText('First Steps');
    expect(firstStepsAchievement).toBeInTheDocument();
    expect(await screen.findByText('Recorded your first expense')).toBeInTheDocument();
  });

  test('applies theme class to body', async () => {
    await renderInsights();
    
    await waitFor(() => {
      expect(document.body.classList.contains('insights-page')).toBe(true);
      expect(document.body.getAttribute('data-theme')).toBe('light');
    });
  });

  const debugDOM = () => {
    console.log('Current DOM:');
    console.log(screen.debug());
  };
}); 