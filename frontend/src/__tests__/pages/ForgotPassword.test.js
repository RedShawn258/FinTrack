import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import ForgotPassword from '../../pages/ForgotPassword';
import { resetPassword } from '../../utils/api';
import { act } from 'react-dom/test-utils';

// Mock the API call
jest.mock('../../utils/api', () => {
  return {
    resetPassword: jest.fn().mockImplementation(() => {
      console.log('Mock resetPassword was called!');
      return Promise.resolve({ success: true });
    })
  };
});

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
    return render(
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

  test('renders forgot password form', () => {
    renderForgotPassword();
    expect(screen.getByText(/forgot password/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /send reset link/i })).toBeInTheDocument();
  });

  test('shows error when form is submitted without email', () => {
    const alertMock = jest.spyOn(window, 'alert').mockImplementation();
    renderForgotPassword();
    
    fireEvent.click(screen.getByRole('button', { name: /send reset link/i }));
    
    expect(alertMock).toHaveBeenCalledWith('Please enter your email address');
    alertMock.mockRestore();
  });

  test('transitions to success state after form submission', async () => {
    renderForgotPassword();
    
    // Fill in the email field
    fireEvent.change(screen.getByPlaceholderText('Email'), {
      target: { value: 'test@example.com' }
    });
    
    // Submit the form by clicking the submit button
    fireEvent.click(screen.getByRole('button', { name: /send reset link/i }));
    
    // Check that the success view is displayed
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /check your email/i })).toBeInTheDocument();
      expect(screen.getByText(/test@example.com/i)).toBeInTheDocument();
    });
  });

  test('displays link to return to login page', () => {
    renderForgotPassword();
    const loginLink = screen.getByText('Back to Login');
    expect(loginLink).toBeInTheDocument();
    expect(loginLink.getAttribute('href')).toBe('/login');
  });
}); 