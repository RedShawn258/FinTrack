import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import Login from '../../pages/Login';
import { loginUser } from '../../utils/api';
import { AuthContext } from '../../context/AuthContext';

// Mock the API call
jest.mock('../../utils/api', () => ({
  loginUser: jest.fn()
}));

// Mock the navigation
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate
}));

// Mock window.alert
const mockAlert = jest.fn();
window.alert = mockAlert;

describe('Login Page', () => {
  const mockLogin = jest.fn();
  
  const renderLogin = () => {
    render(
      <BrowserRouter>
        <AuthContext.Provider value={{ isAuthenticated: false, login: mockLogin }}>
          <Login />
        </AuthContext.Provider>
      </BrowserRouter>
    );
  };

  beforeEach(() => {
    // Clear mock function calls before each test
    loginUser.mockClear();
    mockNavigate.mockClear();
    mockAlert.mockClear();
    mockLogin.mockClear();
  });

  test('renders login form with all fields', () => {
    renderLogin();
    
    expect(screen.getByRole('heading', { name: /login/i })).toBeInTheDocument();
    expect(screen.getByLabelText('Username/Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  test('shows validation error for empty fields', async () => {
    renderLogin();
    
    const loginButton = screen.getByRole('button');
    await fireEvent.click(loginButton);

    expect(mockAlert).toHaveBeenCalledWith('Please enter both username/email and password');
  });

  test('displays forgot password and signup links', () => {
    renderLogin();
    
    const forgotPasswordLink = screen.getByRole('link', { name: /forgot password/i });
    const signupLink = screen.getByRole('link', { name: /signup/i });
    
    expect(forgotPasswordLink).toBeInTheDocument();
    expect(signupLink).toBeInTheDocument();
    expect(signupLink.closest('p')).toHaveTextContent(/don't have an account/i);
  });
}); 