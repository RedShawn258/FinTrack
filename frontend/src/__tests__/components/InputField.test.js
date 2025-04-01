import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import InputField from '../../components/InputField';

describe('InputField Component', () => {
  const mockOnChange = jest.fn();
  
  afterEach(() => {
    mockOnChange.mockClear();
  });
  
  test('renders label and input correctly', () => {
    render(
      <InputField
        label="Test Label"
        type="text"
        value=""
        onChange={mockOnChange}
      />
    );
    
    expect(screen.getByLabelText('Test Label')).toBeInTheDocument();
    // The ID is auto-generated from the label
    expect(screen.getByLabelText('Test Label')).toHaveAttribute('id', 'test-label');
    expect(screen.getByLabelText('Test Label')).toHaveAttribute('type', 'text');
  });
  
  test('handles user input correctly', () => {
    render(
      <InputField
        label="Test Input"
        type="text"
        value=""
        onChange={mockOnChange}
      />
    );
    
    const input = screen.getByLabelText('Test Input');
    fireEvent.change(input, { target: { value: 'test value' } });
    
    expect(mockOnChange).toHaveBeenCalledTimes(1);
  });
  
  test('applies placeholder text when provided', () => {
    render(
      <InputField
        label="Test Input"
        type="text"
        value=""
        placeholder="Enter test value"
        onChange={mockOnChange}
      />
    );
    
    const input = screen.getByLabelText('Test Input');
    expect(input).toHaveAttribute('placeholder', 'Enter test value');
  });
  
  test('renders password input correctly', () => {
    render(
      <InputField
        label="Password"
        type="password"
        value=""
        onChange={mockOnChange}
      />
    );
    
    const input = screen.getByLabelText('Password');
    expect(input).toHaveAttribute('type', 'password');
  });
}); 