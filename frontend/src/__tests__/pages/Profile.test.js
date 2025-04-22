import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';
import Profile from '../../pages/Profile';

// Mock useTheme hook
jest.mock('../../hooks/useTheme', () => ({
    __esModule: true,
    default: () => ({ theme: 'light', setTheme: jest.fn() })
}));

// Mock API calls
jest.mock('../../utils/api', () => ({
    fetchProfile: jest.fn().mockResolvedValue({
        username: 'Test User',
        email: 'test@example.com',
        firstName: 'Test',
        lastName: 'User',
        phoneNumber: '1234567890',
        currency: 'USD',
        notificationsEnabled: true,
        theme: 'light'
    }),
    updateProfile: jest.fn().mockResolvedValue({ success: true })
}));

const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
    ...jest.requireActual('react-router-dom'),
    useNavigate: () => mockNavigate
}));

// Mock InsightsIcon
jest.mock('../../icons', () => ({
    InsightsIcon: () => <div data-testid="insights-icon">Insights Icon</div>
}));

describe('Profile Component', () => {
    const mockUser = {
        token: 'test-token',
        name: 'Test User',
        email: 'test@example.com'
    };

    const mockAuthContext = {
        user: mockUser,
        logout: jest.fn()
    };

    const renderProfile = () => {
        return render(
            <BrowserRouter>
                <AuthContext.Provider value={mockAuthContext}>
                    <Profile />
                </AuthContext.Provider>
            </BrowserRouter>
        );
    };

    beforeEach(() => {
        jest.clearAllMocks();
    });

    test('renders profile page with user information', async () => {
        renderProfile();
        await waitFor(() => {
            expect(screen.getByText('User Profile')).toBeInTheDocument();
        });
    });

    test('navigates to dashboard when back button is clicked', async () => {
        renderProfile();
        
        // Wait for profile data to load
        await waitFor(() => {
            expect(screen.getByText('User Profile')).toBeInTheDocument();
        });
        
        // Find and click back button
        const backButton = screen.getByText('← Dashboard');
        fireEvent.click(backButton);
        
        expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });

    test('calls logout when logout button is clicked', async () => {
        renderProfile();
        
        // Wait for profile data to load
        await waitFor(() => {
            expect(screen.getByText('User Profile')).toBeInTheDocument();
        });
        
        // Find and click logout button
        const logoutButton = screen.getByTitle('Logout');
        fireEvent.click(logoutButton);
        
        expect(mockAuthContext.logout).toHaveBeenCalled();
        expect(mockNavigate).toHaveBeenCalledWith('/');
    });

    test('navigates to insights when insights button is clicked', async () => {
        renderProfile();
        
        // Wait for profile data to load
        await waitFor(() => {
            expect(screen.getByText('User Profile')).toBeInTheDocument();
        });
        
        // Find and click insights button
        const insightsButton = screen.getByTitle('Insights');
        fireEvent.click(insightsButton);
        
        expect(mockNavigate).toHaveBeenCalledWith('/insights');
    });
}); 