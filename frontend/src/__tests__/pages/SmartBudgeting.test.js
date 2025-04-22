import React from 'react';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import '@testing-library/jest-dom';
import SmartBudgeting from '../../pages/SmartBudgeting';

describe('Smart Budgeting Guide Page', () => {
  beforeEach(() => {
    render(
      <BrowserRouter>
        <SmartBudgeting />
      </BrowserRouter>
    );
  });

  test('renders the page title correctly', () => {
    expect(screen.getByText('Smart Budgeting Guide')).toBeInTheDocument();
  });

  test('renders the back to home link', () => {
    const backLink = screen.getByText('← Back to Home');
    expect(backLink).toBeInTheDocument();
    expect(backLink).toHaveAttribute('href', '/');
  });

  test('renders the section header', () => {
    expect(screen.getByText('Creating an Effective Budget')).toBeInTheDocument();
  });

  test('renders the step 1 content', () => {
    expect(screen.getByText('1. Assess Your Income')).toBeInTheDocument();
    expect(screen.getByText('Start with your monthly income:')).toBeInTheDocument();
    expect(screen.getByText('Calculate total monthly earnings')).toBeInTheDocument();
    expect(screen.getByText('Account for variable income')).toBeInTheDocument();
    expect(screen.getByText('Consider all income sources')).toBeInTheDocument();
  });

  test('renders the step 2 content', () => {
    expect(screen.getByText('2. Plan Your Spending')).toBeInTheDocument();
    expect(screen.getByText('Allocate your money wisely:')).toBeInTheDocument();
    expect(screen.getByText('Essential expenses first')).toBeInTheDocument();
    expect(screen.getByText('Set realistic limits')).toBeInTheDocument();
    expect(screen.getByText('Include some flexibility')).toBeInTheDocument();
  });

  test('renders the step 3 content', () => {
    expect(screen.getByText('3. Monitor & Adjust')).toBeInTheDocument();
    expect(screen.getByText('Keep your budget on track:')).toBeInTheDocument();
    expect(screen.getByText('Track progress regularly')).toBeInTheDocument();
    expect(screen.getByText('Make adjustments as needed')).toBeInTheDocument();
    expect(screen.getByText('Stay committed to your goals')).toBeInTheDocument();
  });

  test('renders the correct number of steps', () => {
    const steps = screen.getAllByRole('heading', { level: 3 });
    expect(steps).toHaveLength(3);
  });
}); 