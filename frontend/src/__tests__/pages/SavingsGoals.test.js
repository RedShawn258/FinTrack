import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import SavingsGoalsGuide from '../../pages/SavingsGoals';

describe('Savings Goals Guide Page', () => {
  beforeEach(() => {
    render(<SavingsGoalsGuide />);
  });

  test('renders the page title correctly', () => {
    expect(screen.getByText('Mastering Your Savings Goals: A Comprehensive Guide')).toBeInTheDocument();
  });

  test('renders the introduction paragraph', () => {
    expect(screen.getByText(/Setting and achieving savings goals is a cornerstone of financial wellness/)).toBeInTheDocument();
  });

  test('renders the psychology of saving section', () => {
    expect(screen.getByText('Understanding the Psychology of Saving')).toBeInTheDocument();
    expect(screen.getByText(/Saving money isn't just about numbers/)).toBeInTheDocument();
    expect(screen.getByText((content) => {
      return content.includes('mental accounting') && content.includes('Richard Thaler');
    })).toBeInTheDocument();
  });

  test('renders the SMART savings goals section', () => {
    expect(screen.getByText('Setting SMART Savings Goals')).toBeInTheDocument();
    expect(screen.getByText('Specific:')).toBeInTheDocument();
    expect(screen.getByText('Measurable:')).toBeInTheDocument();
    expect(screen.getByText('Achievable:')).toBeInTheDocument();
    expect(screen.getByText('Relevant:')).toBeInTheDocument();
    expect(screen.getByText('Time-bound:')).toBeInTheDocument();
  });

  test('renders the prioritizing goals section', () => {
    expect(screen.getByText('Prioritizing Multiple Savings Goals')).toBeInTheDocument();
    expect(screen.getByText('Time Sensitivity:')).toBeInTheDocument();
    expect(screen.getByText('Financial Impact:')).toBeInTheDocument();
    expect(screen.getByText('Return on Investment:')).toBeInTheDocument();
    expect(screen.getByText('Emotional Significance:')).toBeInTheDocument();
  });

  test('renders the strategic savings plan section', () => {
    expect(screen.getByText('Creating a Strategic Savings Plan')).toBeInTheDocument();
    expect(screen.getByText('1. Calculate Your Target Savings Rate')).toBeInTheDocument();
    expect(screen.getByText('2. Analyze Your Current Financial Situation')).toBeInTheDocument();
    expect(screen.getByText('3. Automate Your Savings')).toBeInTheDocument();
    expect(screen.getByText('4. Choose the Right Savings Vehicles')).toBeInTheDocument();
  });

  test('renders the advanced savings strategies section', () => {
    expect(screen.getByText('Advanced Savings Strategies')).toBeInTheDocument();
    expect(screen.getByText('Savings Ladders')).toBeInTheDocument();
    expect(screen.getByText('Value-Based Budgeting')).toBeInTheDocument();
    expect(screen.getByText('Savings Windfalls')).toBeInTheDocument();
    expect(screen.getByText('Micro-Saving Apps')).toBeInTheDocument();
  });

  test('renders the overcoming obstacles section', () => {
    expect(screen.getByText('Overcoming Savings Obstacles')).toBeInTheDocument();
    expect(screen.getByText('Inconsistent Income')).toBeInTheDocument();
    expect(screen.getByText('High Debt Burden')).toBeInTheDocument();
    expect(screen.getByText('Lifestyle Inflation')).toBeInTheDocument();
  });
}); 