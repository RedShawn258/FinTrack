import React from 'react';
import { render, screen, fireEvent, within } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../../pages/Landing';
import { AuthContext } from '../../context/AuthContext';

// Mock navigate
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate
}));

describe('Landing Page', () => {
  const renderLanding = (isAuthenticated = false) => {
    return render(
      <BrowserRouter>
        <AuthContext.Provider value={{ isAuthenticated }}>
          <Landing />
        </AuthContext.Provider>
      </BrowserRouter>
    );
  };
  
  beforeEach(() => {
    jest.clearAllMocks();
  });
  
  test('renders hero section with title and description', () => {
    renderLanding();
    
    // Check for main title and description
    expect(screen.getByRole('heading', { name: /simplify your financial journey/i })).toBeInTheDocument();
    expect(screen.getByText(/all-in-one solution/i)).toBeInTheDocument();
  });
  
  test('renders call-to-action buttons', () => {
    renderLanding();
    
    const demoButton = screen.getByRole('link', { name: /start demo/i });
    const learnMoreButton = screen.getByRole('button', { name: /learn more/i });
    
    expect(demoButton).toBeInTheDocument();
    expect(learnMoreButton).toBeInTheDocument();
  });
  
  test('navigates to signup page when demo button is clicked', () => {
    renderLanding();
    
    const demoButton = screen.getByRole('link', { name: /start demo/i });
    
    // Should link to signup page
    expect(demoButton.getAttribute('href')).toBe('/signup');
  });
  
  test('renders features section', () => {
    renderLanding();
    
    // Use heading roles to be more specific
    expect(screen.getByText(/why choose fintrack/i)).toBeInTheDocument();
    
    // Find the "Expense Tracking" header in the key-features section
    const keyFeaturesSection = screen.getByRole('heading', { name: /why choose fintrack/i }).closest('section');
    const expenseTrackingHeading = within(keyFeaturesSection).getByRole('heading', { name: /expense tracking/i });
    
    expect(expenseTrackingHeading).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /budget goals/i })).toBeInTheDocument();
  });
  
  test('has read more button that navigates', () => {
    renderLanding();
    
    // Find the read more button for Budget Planning
    const readMoreButton = screen.getAllByText(/read more/i)[1];
    fireEvent.click(readMoreButton);
    
    // Should navigate to budget planning page
    expect(mockNavigate).toHaveBeenCalledWith('/budget-planning');
  });
}); 