import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import '@testing-library/jest-dom';
import ResetPassword from '../../pages/ResetPassword';

// Mock react-router-dom hooks
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useSearchParams: () => [
    { get: (param) => param === 'token' ? 'valid-token' : null }
  ],
  useNavigate: () => jest.fn()
}));

// Mock window.alert
window.alert = jest.fn();

describe('Reset Password Page', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders the reset password form with correct fields', () => {
    render(
      <BrowserRouter>
        <ResetPassword />
      </BrowserRouter>
    );

    expect(screen.getByText('Reset Your Password')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('New Password')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Confirm Password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /submit/i })).toBeInTheDocument();
    expect(screen.getByText('Back to Login')).toBeInTheDocument();
  });

  test('shows error when passwords do not match', async () => {
    render(
      <BrowserRouter>
        <ResetPassword />
      </BrowserRouter>
    );

    fireEvent.change(screen.getByPlaceholderText('New Password'), { target: { value: 'newpassword123' } });
    fireEvent.change(screen.getByPlaceholderText('Confirm Password'), { target: { value: 'differentpassword' } });
    
    fireEvent.click(screen.getByRole('button', { name: /submit/i }));
    
    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith('Passwords do not match');
    });
  });

  test('shows error when password fields are empty', async () => {
    render(
      <BrowserRouter>
        <ResetPassword />
      </BrowserRouter>
    );
    
    fireEvent.click(screen.getByRole('button', { name: /submit/i }));
    
    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith('Please fill out both password fields');
    });
  });
}); 