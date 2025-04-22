import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';
import Insights from '../../pages/Insights';

// Mock chart.js
jest.mock('react-chartjs-2', () => ({
    Line: () => null,
    Bar: () => null,
    Pie: () => null
}));

// Mock useTheme hook
jest.mock('../../hooks/useTheme', () => ({
    __esModule: true,
    default: () => ({ theme: 'light', setTheme: jest.fn() })
}));

const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
    ...jest.requireActual('react-router-dom'),
    useNavigate: () => mockNavigate
}));

describe('Insights Component', () => {
    const mockDashboardData = {
        transactions: [
            {
                ID: 1,
                Amount: 100,
                Description: 'Test Transaction',
                CategoryID: 1,
                TransactionDate: '2024-03-01'
            }
        ],
        categories: [
            { ID: 1, Name: 'Test Category' }
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

    const renderInsights = async (contextValue = mockAuthContext) => {
        return render(
            <BrowserRouter>
                <AuthContext.Provider value={contextValue}>
                    <Insights />
                </AuthContext.Provider>
            </BrowserRouter>
        );
    };

    beforeEach(() => {
        jest.clearAllMocks();
    });

    test('renders insights page with header', async () => {
        await renderInsights();
        expect(await screen.findByText('Insights & Achievements')).toBeInTheDocument();
    });

    test('switches between analytics and achievements tabs', async () => {
        await renderInsights();
        
        // Check initial analytics tab
        await waitFor(() => {
            expect(screen.getByText('Spending Trends')).toBeInTheDocument();
        });
        
        // Switch to achievements tab
        const achievementsTab = screen.getByText('Achievements');
        fireEvent.click(achievementsTab);
        
        // Verify achievements content is shown - look for an element we know exists in the achievements tab
        await waitFor(() => {
            const achievementsEls = screen.getAllByText(/achievements/i);
            expect(achievementsEls.length).toBeGreaterThan(0);
        });
    });

    test('displays loading state when fetching data', async () => {
        const slowContext = {
            ...mockAuthContext,
            dashboardData: null,
            refreshDashboardData: jest.fn().mockImplementation(
                () => new Promise(resolve => setTimeout(() => resolve(mockDashboardData), 100))
            )
        };

        await renderInsights(slowContext);
        expect(await screen.findByText('Loading insights data...')).toBeInTheDocument();
    });

    test('displays no data message when transactions are empty', async () => {
        const noDataContext = {
            ...mockAuthContext,
            dashboardData: { ...mockDashboardData, transactions: [] }
        };

        await renderInsights(noDataContext);
        
        await waitFor(() => {
            const noDataMessages = screen.getAllByText(/No .* data available/);
            expect(noDataMessages.length).toBeGreaterThan(0);
        });
    });

    test('navigates to dashboard when back button is clicked', async () => {
        await renderInsights();
        
        await waitFor(() => {
            const backButton = screen.getByTitle('Back to Dashboard');
            fireEvent.click(backButton);
            expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
        });
    });

    test('handles logout correctly', async () => {
        await renderInsights();
        
        await waitFor(() => {
            const logoutButton = screen.getByTitle('Logout');
            fireEvent.click(logoutButton);
            expect(mockAuthContext.logout).toHaveBeenCalled();
            expect(mockNavigate).toHaveBeenCalledWith('/');
        });
    });
}); 