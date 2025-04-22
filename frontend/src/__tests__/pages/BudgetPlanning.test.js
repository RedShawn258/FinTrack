import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import BudgetPlanningGuide from '../../pages/BudgetPlanning';
import { BrowserRouter } from 'react-router-dom';

describe('Budget Planning Guide Page', () => {
  beforeEach(() => {
    render(
      <BrowserRouter>
        <BudgetPlanningGuide />
      </BrowserRouter>
    );
  });

  test('renders the page title correctly', () => {
    expect(screen.getByText('Budget Planning: Your Path to Financial Success')).toBeInTheDocument();
  });

  test('renders the introduction paragraph', () => {
    expect(screen.getByText(/Creating and maintaining a realistic budget is the foundation/)).toBeInTheDocument();
  });

  test('renders the 50/30/20 rule section', () => {
    expect(screen.getByText('The 50/30/20 Rule')).toBeInTheDocument();
    expect(screen.getByText('50% for Needs:')).toBeInTheDocument();
    expect(screen.getByText('30% for Wants:')).toBeInTheDocument();
    expect(screen.getByText('20% for Savings:')).toBeInTheDocument();
  });

  test('renders the steps for creating a budget', () => {
    expect(screen.getByText('Step 1: Calculate Your Total Income')).toBeInTheDocument();
    expect(screen.getByText('Step 2: Track Your Expenses')).toBeInTheDocument();
  });

  test('renders budgeting strategies section', () => {
    expect(screen.getByText('Smart Budgeting Strategies')).toBeInTheDocument();
    expect(screen.getByText('1. Use Digital Tools')).toBeInTheDocument();
    expect(screen.getByText('2. Build an Emergency Fund')).toBeInTheDocument();
    expect(screen.getByText('3. Review and Adjust Regularly')).toBeInTheDocument();
  });

  test('renders common budgeting challenges section', () => {
    expect(screen.getByText('Common Budgeting Challenges')).toBeInTheDocument();
    expect(screen.getByText('1. Irregular Income')).toBeInTheDocument();
    expect(screen.getByText('2. Unexpected Expenses')).toBeInTheDocument();
  });

  test('renders advanced budgeting tips section', () => {
    expect(screen.getByText('Advanced Budgeting Tips')).toBeInTheDocument();
    expect(screen.getByText('Automate Your Finances')).toBeInTheDocument();
    expect(screen.getByText('Use Cash Envelopes')).toBeInTheDocument();
  });

  test('renders the conclusion with action steps', () => {
    expect(screen.getByText('Taking Action')).toBeInTheDocument();
    const steps = screen.getAllByRole('listitem');
    expect(steps.length).toBeGreaterThanOrEqual(6);
  });

  test('renders the quote section', () => {
    expect(screen.getByText(/A budget is telling your money where to go/)).toBeInTheDocument();
    expect(screen.getByText('— Dave Ramsey')).toBeInTheDocument();
  });
}); 