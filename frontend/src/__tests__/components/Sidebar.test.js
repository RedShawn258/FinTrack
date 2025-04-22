import React from 'react';
import { render, screen } from '@testing-library/react';
import { BrowserRouter, MemoryRouter } from 'react-router-dom';
import '@testing-library/jest-dom';
import Sidebar from '../../components/Sidebar';

describe('Sidebar Component', () => {
  test('renders sidebar with navigation links', () => {
    render(
      <BrowserRouter>
        <Sidebar />
      </BrowserRouter>
    );
    
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Profile')).toBeInTheDocument();
    expect(screen.getByText('Insights')).toBeInTheDocument();
  });
  
  test('sets active class for current route', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Sidebar />
      </MemoryRouter>
    );
    
    const dashboardLink = screen.getByText('Dashboard').closest('li');
    const profileLink = screen.getByText('Profile').closest('li');
    
    expect(dashboardLink).toHaveClass('active');
    expect(profileLink).not.toHaveClass('active');
  });
  
  test('dashboard link navigates to /dashboard', () => {
    render(
      <BrowserRouter>
        <Sidebar />
      </BrowserRouter>
    );
    
    const dashboardLink = screen.getByText('Dashboard').closest('a');
    expect(dashboardLink).toHaveAttribute('href', '/dashboard');
  });
  
  test('profile link navigates to /profile', () => {
    render(
      <BrowserRouter>
        <Sidebar />
      </BrowserRouter>
    );
    
    const profileLink = screen.getByText('Profile').closest('a');
    expect(profileLink).toHaveAttribute('href', '/profile');
  });
  
  test('insights link navigates to /insights', () => {
    render(
      <BrowserRouter>
        <Sidebar />
      </BrowserRouter>
    );
    
    const insightsLink = screen.getByText('Insights').closest('a');
    expect(insightsLink).toHaveAttribute('href', '/insights');
  });
}); 