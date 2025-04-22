import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import Profile from '../pages/Profile';
import * as api from '../utils/api';
import { act } from 'react';

// Mock the API module
jest.mock('../utils/api');

// Mock the useTheme hook
jest.mock('../hooks/useTheme', () => ({
  __esModule: true,
  default: () => ({ theme: 'light', setTheme: jest.fn() })
}));

const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate
}));

const mockProfileData = {
  username: 'testuser',
  email: 'test@example.com',
  firstName: 'Test',
  lastName: 'User',
  phoneNumber: '1234567890',
  currency: 'USD',
  notificationsEnabled: true,
  theme: 'light'
};

const mockAuthContext = {
  user: { token: 'test-token' },
  logout: jest.fn()
};

// Configure mocks before tests
beforeEach(() => {
  jest.clearAllMocks();
  
  // Mock the api module's functions to return properly structured responses
  api.fetchProfile.mockResolvedValue({ data: mockProfileData });
  api.updateProfile.mockResolvedValue({ data: { message: 'Profile updated successfully' } });
});

const renderProfile = async (contextValue = mockAuthContext) => {
  const utils = render(
    <BrowserRouter>
      <AuthContext.Provider value={contextValue}>
        <Profile />
      </AuthContext.Provider>
    </BrowserRouter>
  );

  // Wait for initial loading to complete
  await waitFor(() => {
    expect(screen.queryByText('Loading profile...')).not.toBeInTheDocument();
  });

  return utils;
};

describe('Profile Component', () => {
  test('renders profile page with header', async () => {
    await renderProfile();
    expect(screen.getByText('User Profile')).toBeInTheDocument();
  });

  test('loads and displays user profile data', async () => {
    await renderProfile();
    
    await waitFor(() => {
      expect(screen.getByDisplayValue('testuser')).toBeInTheDocument();
      expect(screen.getByDisplayValue('test@example.com')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Test')).toBeInTheDocument();
      expect(screen.getByDisplayValue('User')).toBeInTheDocument();
    });
  });

  test('handles profile update successfully', async () => {
    await renderProfile();

    await waitFor(() => {
      expect(screen.getByDisplayValue('Test')).toBeInTheDocument();
    });

    // Update first name
    await act(async () => {
      const firstNameInput = screen.getByPlaceholderText('Enter your first name');
      fireEvent.change(firstNameInput, { target: { value: 'Updated' } });
    });

    // Submit form
    await act(async () => {
      const saveButton = screen.getByText('Save Changes');
      fireEvent.click(saveButton);
    });

    await waitFor(() => {
      expect(api.updateProfile).toHaveBeenCalledWith('test-token', expect.objectContaining({
        firstName: 'Updated'
      }));
      expect(screen.getByText('Profile updated successfully!')).toBeInTheDocument();
    });
  });

  test('handles theme change', async () => {
    await renderProfile();

    await waitFor(() => {
      expect(screen.getByText('Light')).toBeInTheDocument();
      expect(screen.getByText('Dark')).toBeInTheDocument();
    });

    // Find and click the Dark theme radio button
    await act(async () => {
      const darkThemeRadio = screen.getByLabelText('Dark');
      fireEvent.click(darkThemeRadio);
    });

    expect(screen.getByLabelText('Dark')).toBeChecked();
  });

  test('handles notification toggle', async () => {
    await renderProfile();

    await waitFor(() => {
      expect(screen.getByText('Enable Notifications')).toBeInTheDocument();
    });

    // Find and click the checkbox
    await act(async () => {
      const notificationToggle = screen.getByRole('checkbox', { name: /enable notifications/i });
      fireEvent.click(notificationToggle);
    });

    expect(screen.getByRole('checkbox', { name: /enable notifications/i })).not.toBeChecked();
  });

  test('navigates back to dashboard', async () => {
    await renderProfile();
    
    await act(async () => {
      const backButton = screen.getByText('← Dashboard');
      fireEvent.click(backButton);
    });

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });

  test('handles logout', async () => {
    await renderProfile();
    
    await act(async () => {
      const logoutButton = screen.getByTitle('Logout');
      fireEvent.click(logoutButton);
    });

    expect(mockAuthContext.logout).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  test('displays loading state', async () => {
    // Mock a delayed API response
    api.fetchProfile.mockImplementationOnce(() => 
      new Promise(resolve => setTimeout(() => resolve({ data: mockProfileData }), 100))
    );
    
    render(
      <BrowserRouter>
        <AuthContext.Provider value={mockAuthContext}>
          <Profile />
        </AuthContext.Provider>
      </BrowserRouter>
    );
    
    expect(screen.getByText('Loading profile...')).toBeInTheDocument();
  });

  test('handles API error', async () => {
    api.fetchProfile.mockRejectedValueOnce(new Error('API Error'));
    
    render(
      <BrowserRouter>
        <AuthContext.Provider value={mockAuthContext}>
          <Profile />
        </AuthContext.Provider>
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Failed to load profile data. Please try again.')).toBeInTheDocument();
    });
  });

  test('applies theme class to body', async () => {
    await renderProfile();
    
    await waitFor(() => {
      expect(document.body.classList.contains('auth-page')).toBe(true);
      expect(document.body.getAttribute('data-theme')).toBe('light');
    });
  });
}); 