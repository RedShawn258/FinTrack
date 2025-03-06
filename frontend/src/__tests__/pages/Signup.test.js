import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import Signup from '../../pages/Signup';
import { registerUser } from '../../utils/api';
import { AuthContext } from '../../context/AuthContext';

// Mock the API call
jest.mock('../../utils/api', () => ({
  registerUser: jest.fn()
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

describe('Signup Page', () => {
  const renderSignup = () => {
    render(
      <BrowserRouter>
        <AuthContext.Provider value={{ isAuthenticated: false, setIsAuthenticated: jest.fn() }}>
          <Signup />
        </AuthContext.Provider>
      </BrowserRouter>
    );
  };

  beforeEach(() => {
    // Clear mock function calls before each test
    registerUser.mockClear();
    mockNavigate.mockClear();
    mockAlert.mockClear();
  });

  test('renders signup form with all fields', () => {
    renderSignup();
    
    expect(screen.getByRole('heading', { name: /signup/i })).toBeInTheDocument();
    expect(screen.getByLabelText('Username')).toBeInTheDocument();
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm Password')).toBeInTheDocument();
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  test('shows validation error for empty fields', async () => {
    renderSignup();
    
    const signupButton = screen.getByRole('button');
    await fireEvent.click(signupButton);

    expect(mockAlert).toHaveBeenCalledWith('Signup failed');
  });

  test('shows validation error for mismatched passwords', async () => {
    renderSignup();
    
    const usernameInput = screen.getByLabelText('Username');
    const emailInput = screen.getByLabelText('Email');
    const passwordInput = screen.getByLabelText('Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password');
    const signupButton = screen.getByRole('button');

    fireEvent.change(usernameInput, { target: { value: 'Mahesh' } });
    fireEvent.change(emailInput, { target: { value: 'durgamaheshboppani@gmail.com' } });
    fireEvent.change(passwordInput, { target: { value: 'Mahesh@1078' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'wrongpassword' } });
    await fireEvent.click(signupButton);

    expect(mockAlert).toHaveBeenCalledWith('Passwords do not match');
    expect(registerUser).not.toHaveBeenCalled();
  });

  test('displays login link', () => {
    renderSignup();
    
    const loginLink = screen.getByRole('link', { name: /login/i });
    expect(loginLink).toBeInTheDocument();
    expect(loginLink.closest('p')).toHaveTextContent(/already have an account/i);
  });
}); 