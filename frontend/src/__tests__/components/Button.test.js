import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import Button from '../../components/Button';

describe('Button Component', () => {
  test('renders button with correct text', () => {
    render(<Button text="Test Button" />);
    
    const button = screen.getByRole('button');
    expect(button).toHaveTextContent('Test Button');
  });
  
  test('handles click events', () => {
    const mockOnClick = jest.fn();
    render(<Button text="Click Me" onClick={mockOnClick} />);
    
    const button = screen.getByRole('button');
    fireEvent.click(button);
    
    expect(mockOnClick).toHaveBeenCalledTimes(1);
  });
  
  test('disables button when disabled prop is true', () => {
    render(<Button text="Disabled Button" disabled={true} />);
    
    const button = screen.getByRole('button');
    expect(button).toBeDisabled();
  });
}); 