import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import Navbar from '../../components/Navbar';
import { AuthContext } from '../../context/AuthContext';

// Mock navigation
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate
}));

describe('Navbar Component', () => {
  const renderNavbar = (user = null) => {
    const mockLogout = jest.fn();
    
    return render(
      <BrowserRouter>
        <AuthContext.Provider value={{ user, logout: mockLogout }}>
          <Navbar />
        </AuthContext.Provider>
      </BrowserRouter>
    );
  };

  test('renders logo and navigation links when not authenticated', () => {
    renderNavbar(null);
    
    expect(screen.getByText(/fintrack/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /login/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /signup/i })).toBeInTheDocument();
  });

  test('renders dashboard link when authenticated', () => {
    renderNavbar({ id: 1, name: 'Test User' });
    
    expect(screen.getByText(/fintrack/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /dashboard/i })).toBeInTheDocument();
  });

  test('logout button works correctly', () => {
    const { getByRole } = renderNavbar({ id: 1, name: 'Test User' });
    
    const logoutButton = getByRole('button', { name: /logout/i });
    expect(logoutButton).toBeInTheDocument();
    
    fireEvent.click(logoutButton);
  });
}); 