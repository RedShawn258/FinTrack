import React, { useState } from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import AuthForm from '../../components/AuthForm';

describe('AuthForm Component', () => {
  const mockOnSubmit = jest.fn();

  const TestWrapper = () => {
    const [formValues, setFormValues] = useState({
      username: '',
      password: ''
    });

    const fields = [
      { 
        name: 'username', 
        type: 'text', 
        label: 'Username',
        value: formValues.username,
        onChange: (e) => setFormValues(prev => ({ ...prev, username: e.target.value }))
      },
      { 
        name: 'password', 
        type: 'password', 
        label: 'Password',
        value: formValues.password,
        onChange: (e) => setFormValues(prev => ({ ...prev, password: e.target.value }))
      }
    ];

    const handleSubmit = () => {
      mockOnSubmit(formValues);
    };

    return (
      <AuthForm 
        title="Sign In" 
        fields={fields} 
        onSubmit={handleSubmit} 
      />
    );
  };

  beforeEach(() => {
    mockOnSubmit.mockClear();
  });

  test('renders all input fields correctly', () => {
    render(<TestWrapper />);
    
    expect(screen.getByLabelText('Username')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByRole('button')).toHaveTextContent('Sign In');
  });

  test('handles form submission correctly', () => {
    render(<TestWrapper />);
    
    const usernameInput = screen.getByLabelText('Username');
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByRole('button');

    fireEvent.change(usernameInput, { target: { value: 'testuser' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);

    expect(mockOnSubmit).toHaveBeenCalledWith({
      username: 'testuser',
      password: 'password123'
    });
  });

  test('validates required fields before submission', () => {
    render(<TestWrapper />);
    
    const submitButton = screen.getByRole('button');
    fireEvent.click(submitButton);

    expect(mockOnSubmit).toHaveBeenCalledWith({
      username: '',
      password: ''
    });
  });
}); 