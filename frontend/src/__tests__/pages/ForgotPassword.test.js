import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import ForgotPassword from '../../pages/ForgotPassword';
import { resetPassword } from '../../utils/api';

// Mock the API call
jest.mock('../../utils/api', () => ({
  resetPassword: jest.fn()
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

describe('ForgotPassword Page', () => {
  const renderForgotPassword = () => {
    render(
      <BrowserRouter>
        <ForgotPassword />
      </BrowserRouter>
    );
  };

  beforeEach(() => {
    // Clear mock function calls before each test
    resetPassword.mockClear();
    mockNavigate.mockClear();
    mockAlert.mockClear();
  });

  test('renders reset password form with all fields', () => {
    renderForgotPassword();
    
    expect(screen.getByRole('heading', { name: /reset password/i })).toBeInTheDocument();
    expect(screen.getByLabelText('Username/Email')).toBeInTheDocument();
    expect(screen.getByLabelText('New Password')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm Password')).toBeInTheDocument();
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  test('shows validation error for empty fields', async () => {
    renderForgotPassword();
    
    const resetButton = screen.getByRole('button');
    await fireEvent.click(resetButton);

    expect(mockAlert).toHaveBeenCalledWith('Failed to reset password');
  });

  test('shows validation error for mismatched passwords', async () => {
    renderForgotPassword();
    
    const identifierInput = screen.getByLabelText('Username/Email');
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password');
    const resetButton = screen.getByRole('button');

    fireEvent.change(identifierInput, { target: { value: 'Mahesh' } });
    fireEvent.change(newPasswordInput, { target: { value: 'Mahesh@1078' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'wrongpassword' } });
    await fireEvent.click(resetButton);

    expect(mockAlert).toHaveBeenCalledWith('Passwords do not match');
    expect(resetPassword).not.toHaveBeenCalled();
  });

  test('displays back to login link', () => {
    renderForgotPassword();
    
    const loginLink = screen.getByRole('link', { name: /back to login/i });
    expect(loginLink).toBeInTheDocument();
  });
}); 