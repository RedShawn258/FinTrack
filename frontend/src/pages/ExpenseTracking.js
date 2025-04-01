import { Link } from 'react-router-dom';
import './Guide.css';

export default function ExpenseTracking() {
  return (
    <div className="guide-page">
      <Link to="/" className="back-link">← Back to Home</Link>
      <h1>Expense Tracking Guide</h1>
      
      <section className="guide-section">
        <h2>Getting Started with Expense Tracking</h2>
        <p>Track your expenses effectively with these simple steps:</p>
        
        <div className="steps">
          <div className="step">
            <h3>1. Categorize Your Expenses</h3>
            <p>Group your spending into categories like:</p>
            <ul>
              <li>Food & Groceries</li>
              <li>Transportation</li>
              <li>Entertainment</li>
              <li>Bills & Utilities</li>
              <li>Shopping</li>
            </ul>
          </div>

          <div className="step">
            <h3>2. Regular Updates</h3>
            <p>Make it a habit to log your expenses:</p>
            <ul>
              <li>Record expenses daily</li>
              <li>Keep your receipts</li>
              <li>Use our mobile app for on-the-go tracking</li>
            </ul>
          </div>

          <div className="step">
            <h3>3. Review & Analyze</h3>
            <p>Understanding your spending patterns:</p>
            <ul>
              <li>Check weekly summaries</li>
              <li>Identify major spending areas</li>
              <li>Spot opportunities to save</li>
            </ul>
          </div>
        </div>
      </section>
    </div>
  );
} 