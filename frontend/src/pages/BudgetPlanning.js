import React from 'react';
import './BudgetPlanning.css';

const BudgetPlanningGuide = () => {
  return (
    <div className="budget-guide-container">
      <div className="budget-guide-header">
        <h1>Budget Planning: Your Path to Financial Success</h1>
        <p className="guide-intro">
          Creating and maintaining a realistic budget is the foundation of financial wellness. 
          This comprehensive guide will help you develop effective budgeting strategies that work 
          for your lifestyle and financial goals.
        </p>
      </div>

      <div className="budget-guide-section">
        <h2>Understanding Budget Basics</h2>
        <p>
          A budget is more than just tracking expenses—it's a financial plan that helps you 
          allocate your resources effectively. The fundamental principle of budgeting is simple: 
          track your income, plan your spending, and adjust as needed.
        </p>
        
        <h3>The 50/30/20 Rule</h3>
        <p>
          One popular budgeting framework is the 50/30/20 rule:
        </p>
        <ul>
          <li><strong>50% for Needs:</strong> Essential expenses like housing, utilities, food, and basic transportation</li>
          <li><strong>30% for Wants:</strong> Non-essential items like entertainment, dining out, and hobbies</li>
          <li><strong>20% for Savings:</strong> Emergency fund, retirement, and other financial goals</li>
        </ul>
      </div>

      <div className="budget-guide-section">
        <h2>Creating Your Budget</h2>
        <h3>Step 1: Calculate Your Total Income</h3>
        <p>
          Start by determining your total monthly income after taxes. Include:
        </p>
        <ul>
          <li>Regular salary or wages</li>
          <li>Freelance income</li>
          <li>Investment returns</li>
          <li>Rental income</li>
          <li>Any other sources of income</li>
        </ul>

        <h3>Step 2: Track Your Expenses</h3>
        <p>
          List all your monthly expenses, categorized as:
        </p>
        <ul>
          <li><strong>Fixed Expenses:</strong> Rent/mortgage, car payments, insurance</li>
          <li><strong>Variable Necessities:</strong> Groceries, utilities, fuel</li>
          <li><strong>Discretionary Spending:</strong> Entertainment, shopping, dining out</li>
          <li><strong>Savings and Debt Payment:</strong> Emergency fund, retirement, credit card payments</li>
        </ul>
      </div>

      <div className="budget-guide-section">
        <h2>Smart Budgeting Strategies</h2>
        <h3>1. Use Digital Tools</h3>
        <p>
          Take advantage of budgeting apps and tools that can:
        </p>
        <ul>
          <li>Automatically track expenses</li>
          <li>Categorize spending</li>
          <li>Set bill reminders</li>
          <li>Monitor progress toward goals</li>
        </ul>

        <h3>2. Build an Emergency Fund</h3>
        <p>
          Aim to save 3-6 months of living expenses in an easily accessible account for unexpected costs.
        </p>

        <h3>3. Review and Adjust Regularly</h3>
        <p>
          Schedule monthly budget reviews to:
        </p>
        <ul>
          <li>Track progress toward goals</li>
          <li>Identify areas for improvement</li>
          <li>Adjust categories as needed</li>
          <li>Plan for upcoming expenses</li>
        </ul>
      </div>

      <div className="budget-guide-section">
        <h2>Common Budgeting Challenges</h2>
        <h3>1. Irregular Income</h3>
        <p>
          If your income varies monthly:
        </p>
        <ul>
          <li>Budget based on your lowest earning month</li>
          <li>Save extra during better months</li>
          <li>Maintain a larger emergency fund</li>
        </ul>

        <h3>2. Unexpected Expenses</h3>
        <p>
          Prepare for surprises by:
        </p>
        <ul>
          <li>Building an emergency fund</li>
          <li>Setting aside money for routine maintenance</li>
          <li>Having adequate insurance coverage</li>
        </ul>
      </div>

      <div className="budget-guide-section">
        <h2>Advanced Budgeting Tips</h2>
        <h3>Automate Your Finances</h3>
        <p>
          Set up automatic transfers for:
        </p>
        <ul>
          <li>Bill payments</li>
          <li>Savings contributions</li>
          <li>Investment deposits</li>
          <li>Debt payments</li>
        </ul>

        <h3>Use Cash Envelopes</h3>
        <p>
          For discretionary spending categories, consider using physical cash envelopes to:
        </p>
        <ul>
          <li>Control spending</li>
          <li>Avoid overspending</li>
          <li>Better visualize your budget</li>
        </ul>
      </div>

      <div className="budget-guide-conclusion">
        <div className="taking-action">
          <h2>Taking Action</h2>
          <p>Start implementing these budgeting strategies today:</p>
          <ol>
            <li>Calculate your monthly income</li>
            <li>Track your expenses for one month</li>
            <li>Create categories for your spending</li>
            <li>Set realistic financial goals</li>
            <li>Choose a budgeting tool or method</li>
            <li>Review and adjust regularly</li>
          </ol>
        </div>

        <div className="quote-section">
          <div className="quote-text">
            "A budget is telling your money where to go instead of wondering where it went."
          </div>
          <div className="quote-author">
            — Dave Ramsey
          </div>
        </div>
      </div>
    </div>
  );
};

export default BudgetPlanningGuide; 