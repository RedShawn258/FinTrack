import { Link } from 'react-router-dom';
import './Guide.css';

export default function SmartBudgeting() {
  return (
    <div className="guide-page">
      <Link to="/" className="back-link">← Back to Home</Link>
      <h1>Smart Budgeting Guide</h1>

      <section className="guide-section">
        <h2>Creating an Effective Budget</h2>
        <p>Follow these principles to create a budget that works:</p>

        <div className="steps">
          <div className="step">
            <h3>1. Assess Your Income</h3>
            <p>Start with your monthly income:</p>
            <ul>
              <li>Calculate total monthly earnings</li>
              <li>Account for variable income</li>
              <li>Consider all income sources</li>
            </ul>
          </div>

          <div className="step">
            <h3>2. Plan Your Spending</h3>
            <p>Allocate your money wisely:</p>
            <ul>
              <li>Essential expenses first</li>
              <li>Set realistic limits</li>
              <li>Include some flexibility</li>
            </ul>
          </div>

          <div className="step">
            <h3>3. Monitor & Adjust</h3>
            <p>Keep your budget on track:</p>
            <ul>
              <li>Track progress regularly</li>
              <li>Make adjustments as needed</li>
              <li>Stay committed to your goals</li>
            </ul>
          </div>
        </div>
      </section>
    </div>
  );
} 