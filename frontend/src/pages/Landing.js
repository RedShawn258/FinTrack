import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import './Landing.css';

const Landing = () => {
  const navigate = useNavigate();

  const handleReadMore = () => {
    navigate('/budget-planning');
  };

  return (
    <>
      {/* Navigation Bar */}
      <nav className="nav-bar">
        <div className="nav-content">
          <div className="logo">
            <Link to="/">FinTrack</Link>
          </div>
          <div className="nav-buttons">
            <Link to="/login" className="login-btn">Login</Link>
            <Link to="/signup" className="signup-btn">Sign Up</Link>
          </div>
        </div>
      </nav>

      <div className="landing-container">
        {/* Hero Section with Tagline */}
        <section className="hero">
          <h1>Simplify Your Financial Journey</h1>
          <p className="tagline">Your all-in-one solution for expense tracking, budgeting, and financial planning</p>
        </section>

        {/* Features Grid */}
        <section className="features">
          <div className="features-grid">
            <div className="feature-card">
              <h3>Smart Budgeting</h3>
              <p>AI-powered budgeting that learns from your spending habits</p>
            </div>
            <div className="feature-card">
              <h3>Goal Tracking</h3>
              <p>Set and achieve your financial goals with visual progress tracking</p>
            </div>
            <div className="feature-card">
              <h3>AI Insights</h3>
              <p>Get personalized recommendations based on your spending patterns</p>
            </div>
          </div>
        </section>

        {/* Calculator Section */}
        <section className="calculator-section">
          <h2>Savings Calculator</h2>
          <div className="calculator-container">
            <div className="input-group">
              <label>Monthly Income</label>
              <input 
                type="number" 
                placeholder="Enter your monthly income"
                className="calculator-input"
              />
            </div>
            <div className="input-group">
              <label>Monthly Expenses</label>
              <input 
                type="number" 
                placeholder="Enter your monthly expenses"
                className="calculator-input"
              />
            </div>
            <button className="calculate-btn">Calculate Savings</button>
          </div>
        </section>

        {/* Key Features Section */}
        <section className="key-features">
          <h2>Why Choose FinTrack?</h2>
          <div className="features-grid">
            <div className="feature-card">
              <div className="feature-icon">
                <i className="fas fa-chart-line"></i>
              </div>
              <h3>Expense Tracking</h3>
              <p>Automatically categorize and track your daily expenses</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
                <i className="fas fa-bullseye"></i>
              </div>
              <h3>Budget Goals</h3>
              <p>Set and achieve your financial goals with smart tracking</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
                <i className="fas fa-chart-pie"></i>
              </div>
              <h3>Visual Analytics</h3>
              <p>Understand your spending with intuitive charts</p>
            </div>
          </div>
        </section>

        {/* How It Works Section */}
        <section className="how-it-works">
          <h2>How It Works</h2>
          <div className="how-it-works-content">
            <div className="how-it-works-item">
              <strong>Connect Your Accounts</strong>
              <p>Securely link your bank accounts for automatic tracking</p>
            </div>
            <div className="how-it-works-item">
              <strong>Set Your Budget</strong>
              <p>Create custom budgets based on your spending habits</p>
            </div>
            <div className="how-it-works-item">
              <strong>Track Progress</strong>
              <p>Monitor your financial goals with real-time updates</p>
            </div>
          </div>
        </section>

        {/* Stats Section */}
        <section className="stats">
          <div className="stat-container">
            <div className="stat-item">
              <i className="fas fa-shield-alt"></i>
              <h3>Secure</h3>
              <p>Your data is protected</p>
            </div>
            <div className="stat-item">
              <i className="fas fa-bolt"></i>
              <h3>Real-time</h3>
              <p>Instant tracking</p>
            </div>
            <div className="stat-item">
              <i className="fas fa-code"></i>
              <h3>Open Source</h3>
              <p>Community driven</p>
            </div>
          </div>
        </section>

        {/* Educational Resources Section */}
        <section className="education-hub">
          <h2>Financial Education Hub</h2>
          <div className="feature-cards">
            <div className="feature-card">
              <h3>Expense Tracking</h3>
              <p>Learn how to categorize and monitor your daily expenses effectively.</p>
              <div className="read-more-container">
                <Link to="/guides/expense-tracking" className="read-more-button">Read More</Link>
              </div>
            </div>
            <div className="feature-card">
              <h3>Budget Planning</h3>
              <p>Discover strategies to create and maintain a realistic budget.</p>
              <div className="read-more-container">
                <button onClick={handleReadMore} className="read-more-button">Read More</button>
              </div>
            </div>
            <div className="feature-card">
              <h3>Savings Goals</h3>
              <p>Set and achieve your personal savings targets with ease.</p>
              <div className="read-more-container">
                <Link to="/guides/savings-goals" className="read-more-button">Read More</Link>
              </div>
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="cta-section">
          <h2>Try Our Budget Management Tool</h2>
          <p>Experience better financial organization with our simple tracking system</p>
          <Link to="/signup" className="cta-button">Start Demo</Link>
        </section>
        
        {/* Feature Showcase Section */}
        <section className="feature-showcase">
          <div className="showcase-content">
            <h2>Track your monthly spending and more.</h2>
            <p className="showcase-description">
              Review your transactions, track your spending by category and receive monthly insights that help you better understand your money habits.
            </p>
            <button className="learn-more-btn">Learn more</button>
          </div>
          <div className="showcase-image">
            <div className="demo-card">
              <div className="spending-overview">
                <h3>May spending</h3>
                <div className="amount">
                  <span className="dollar">$</span>
                  <span className="value">3,507</span>
                </div>
                <p className="comparison">$100 from this time in April</p>
                <p className="days-left">16 days left this month</p>
              </div>
              <div className="spending-chart">
                <div className="chart-container">
                  <div className="bar" style={{ height: '80%' }}><span>Dec</span></div>
                  <div className="bar" style={{ height: '85%' }}><span>Jan</span></div>
                  <div className="bar" style={{ height: '70%' }}><span>Feb</span></div>
                  <div className="bar" style={{ height: '60%' }}><span>Mar</span></div>
                  <div className="bar" style={{ height: '65%' }}><span>Apr</span></div>
                  <div className="bar" style={{ height: '50%' }}><span>May</span></div>
                </div>
              </div>
              <div className="spending-categories">
                <div className="categories-header">
                  <h4>May top spending</h4>
                  <button className="see-all">See all</button>
                </div>
                <div className="category-bars">
                  <div className="category-bar">
                    <div className="bar-fill" style={{ width: '80%', backgroundColor: '#4361ee' }}></div>
                  </div>
                  <div className="category-bar">
                    <div className="bar-fill" style={{ width: '60%', backgroundColor: '#3f37c9' }}></div>
                  </div>
                  <div className="category-bar">
                    <div className="bar-fill" style={{ width: '40%', backgroundColor: '#4895ef' }}></div>
                  </div>
                </div>
              </div>
              <p className="demo-note">Not your actual amounts.<br />For display only.</p>
            </div>
          </div>
        </section>
      </div>
    </>
  );
};

export default Landing; 