import React, { useState, useEffect, useContext, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import { Line, Bar, Pie } from 'react-chartjs-2';
import './Insights.css';
import useTheme from '../hooks/useTheme';

const Insights = () => {
  const { user, dashboardData, refreshDashboardData, logout } = useContext(AuthContext);
  const navigate = useNavigate();
  const token = user?.token;
  const { theme } = useTheme();
  
  const [activeTab, setActiveTab] = useState('analytics');
  const [isLoading, setIsLoading] = useState(true);
  const [spendingTrends, setSpendingTrends] = useState({});
  const [categoryBreakdown, setCategoryBreakdown] = useState({});
  const [monthlySavings, setMonthlySavings] = useState({});
  const [achievements, setAchievements] = useState([]);
  
  const processSpendingTrends = useCallback(() => {
    const transactions = dashboardData?.transactions || [];
    
    // Create a map of the last 6 months
    const today = new Date();
    const last6Months = [];
    const monthLabels = [];
    
    for (let i = 5; i >= 0; i--) {
      const month = new Date(today.getFullYear(), today.getMonth() - i, 1);
      const monthKey = `${month.getFullYear()}-${String(month.getMonth() + 1).padStart(2, '0')}`;
      last6Months.push(monthKey);
      
      // Format for display (e.g., "Jan 2023")
      const monthLabel = month.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
      monthLabels.push(monthLabel);
    }
    
    // Sum transactions by month
    const monthlyExpenses = last6Months.reduce((acc, month) => {
      acc[month] = 0;
      return acc;
    }, {});
    
    transactions.forEach(tx => {
      const date = new Date(tx.TransactionDate);
      const monthKey = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
      
      if (monthlyExpenses[monthKey] !== undefined) {
        monthlyExpenses[monthKey] += tx.Amount;
      }
    });
    
    // Create chart data
    setSpendingTrends({
      labels: monthLabels,
      datasets: [
        {
          label: 'Monthly Expenses',
          data: last6Months.map(month => monthlyExpenses[month]),
          borderColor: '#3b82f6',
          backgroundColor: 'rgba(59, 130, 246, 0.1)',
          borderWidth: 2,
          tension: 0.4,
          fill: true,
          pointBackgroundColor: '#3b82f6',
          pointRadius: 4,
          pointHoverRadius: 6
        }
      ]
    });
  }, [dashboardData]);

  const processCategoryBreakdown = useCallback(() => {
    const transactions = dashboardData?.transactions || [];
    const categories = dashboardData?.categories || [];
    
    // Calculate totals per category
    const categoryTotals = {};
    transactions.forEach(tx => {
      const catId = tx.CategoryID || 'uncategorized';
      if (!categoryTotals[catId]) {
        categoryTotals[catId] = 0;
      }
      categoryTotals[catId] += tx.Amount;
    });
    
    // Prepare data for pie chart
    const labels = [];
    const data = [];
    const backgroundColors = [
      '#3b82f6', '#ef4444', '#10b981', '#f59e0b', 
      '#8b5cf6', '#ec4899', '#6366f1', '#14b8a6',
      '#f97316', '#84cc16', '#06b6d4', '#a855f7'
    ];
    
    Object.keys(categoryTotals).forEach((catId, index) => {
      let catName = 'Uncategorized';
      if (catId !== 'uncategorized') {
        const category = categories.find(c => c.ID === Number(catId));
        if (category) {
          catName = category.Name;
        }
      }
      
      labels.push(catName);
      data.push(categoryTotals[catId]);
    });
    
    setCategoryBreakdown({
      labels,
      datasets: [
        {
          data,
          backgroundColor: backgroundColors.slice(0, data.length),
          borderWidth: 0,
          hoverOffset: 10
        }
      ]
    });
  }, [dashboardData]);

  const processMonthlySavings = useCallback(() => {
    const transactions = dashboardData?.transactions || [];
    
    // Create a map of the last 3 months
    const today = new Date();
    const last3Months = [];
    const monthLabels = [];
    
    for (let i = 2; i >= 0; i--) {
      const month = new Date(today.getFullYear(), today.getMonth() - i, 1);
      const monthKey = `${month.getFullYear()}-${String(month.getMonth() + 1).padStart(2, '0')}`;
      last3Months.push(monthKey);
      
      // Format for display
      const monthLabel = month.toLocaleDateString('en-US', { month: 'short' });
      monthLabels.push(monthLabel);
    }
    
    // Calculate total budget for each month (as a simulated "income")
    const monthlyBudget = last3Months.reduce((acc, month) => {
      acc[month] = 2000; // Default simulated monthly budget
      return acc;
    }, {});
    
    // Calculate total spending per month
    const monthlySpending = last3Months.reduce((acc, month) => {
      acc[month] = 0;
      return acc;
    }, {});
    
    transactions.forEach(tx => {
      const date = new Date(tx.TransactionDate);
      const monthKey = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
      
      if (monthlySpending[monthKey] !== undefined) {
        monthlySpending[monthKey] += tx.Amount;
      }
    });
    
    // Calculate savings (budget - spending)
    const monthlySavingsData = last3Months.map(month => {
      return Math.max(0, monthlyBudget[month] - monthlySpending[month]);
    });
    
    // Calculate spending
    const monthlySpendingData = last3Months.map(month => monthlySpending[month]);
    
    setMonthlySavings({
      labels: monthLabels,
      datasets: [
        {
          label: 'Spending',
          data: monthlySpendingData,
          backgroundColor: '#ef4444',
          borderWidth: 0,
          borderRadius: 4
        },
        {
          label: 'Savings',
          data: monthlySavingsData,
          backgroundColor: '#10b981',
          borderWidth: 0,
          borderRadius: 4
        }
      ]
    });
  }, [dashboardData]);

  // Process data for analytics charts
  const processAnalyticsData = useCallback(() => {
    if (!dashboardData || !dashboardData.transactions) return;
    
    processSpendingTrends();
    processCategoryBreakdown();
    processMonthlySavings();
  }, [dashboardData, processSpendingTrends, processCategoryBreakdown, processMonthlySavings]);
  
  // Generate user achievements based on their data
  const generateAchievements = useCallback(() => {
    const transactions = dashboardData?.transactions || [];
    const budgets = dashboardData?.budgets || [];
    const userAchievements = [];
    
    // First Transaction Achievement
    if (transactions.length > 0) {
      userAchievements.push({
        id: 'first-expense',
        name: 'First Steps',
        description: 'Recorded your first expense',
        icon: '💼',
        unlocked: true,
        date: new Date(transactions[0].TransactionDate).toLocaleDateString()
      });
    }
    
    // Consistent Tracker
    if (transactions.length >= 5) {
      userAchievements.push({
        id: 'consistent-tracker',
        name: 'Consistent Tracker',
        description: 'Recorded 5+ expenses',
        icon: '📊',
        unlocked: true,
        progress: Math.min(100, (transactions.length / 5) * 100)
      });
    }
    
    // Budget Maestro
    if (budgets.length >= 3) {
      userAchievements.push({
        id: 'budget-maestro',
        name: 'Budget Maestro',
        description: 'Created 3+ budgets',
        icon: '🏆',
        unlocked: true,
        progress: Math.min(100, (budgets.length / 3) * 100)
      });
    }
    
    // Category Specialist
    const uniqueCategories = new Set(transactions.map(tx => tx.CategoryID).filter(Boolean));
    if (uniqueCategories.size >= 3) {
      userAchievements.push({
        id: 'category-specialist',
        name: 'Category Specialist',
        description: 'Used 3+ different categories',
        icon: '🏷️',
        unlocked: true,
        progress: Math.min(100, (uniqueCategories.size / 3) * 100)
      });
    }
    
    // Money Saver (simplified simulation)
    const totalBudget = budgets.reduce((sum, budget) => sum + (budget.LimitAmount || 0), 0);
    const totalSpent = transactions.reduce((sum, tx) => sum + tx.Amount, 0);
    if (totalBudget > 0 && totalSpent <= totalBudget) {
      userAchievements.push({
        id: 'money-saver',
        name: 'Budget Champion',
        description: 'Stayed under budget',
        icon: '💰',
        unlocked: true
      });
    }
    
    // Add locked achievements
    const lockedAchievements = [
      {
        id: 'savings-goal',
        name: 'Saving Star',
        description: 'Reach your first savings goal',
        icon: '⭐',
        unlocked: false
      },
      {
        id: 'streak-master',
        name: 'Streak Master',
        description: 'Log in for 7 consecutive days',
        icon: '🔥',
        unlocked: false,
        progress: 30
      },
      {
        id: 'financial-wizard',
        name: 'Financial Wizard',
        description: 'Complete your financial profile',
        icon: '🧙‍♂️',
        unlocked: false,
        progress: 50
      }
    ];
    
    setAchievements([...userAchievements, ...lockedAchievements]);
  }, [dashboardData]);
  
  // Set body class for styling
  useEffect(() => {
    document.body.classList.add('insights-page');
    document.body.setAttribute('data-theme', theme);
    return () => {
      document.body.classList.remove('insights-page');
      document.body.removeAttribute('data-theme');
    };
  }, [theme]);
  
  // Load data when component mounts
  useEffect(() => {
    if (!token) return;
    
    const loadInsightsData = async () => {
      try {
        setIsLoading(true);
        
        // Only refresh if needed
        if (!dashboardData || !dashboardData.transactions || dashboardData.transactions.length === 0) {
          await refreshDashboardData();
        }
        
        // Process the data once available
        processAnalyticsData();
        generateAchievements();
        
      } catch (error) {
        console.error('Failed to load insights data', error);
      } finally {
        setIsLoading(false);
      }
    };
    
    loadInsightsData();
  }, [token, dashboardData, refreshDashboardData, processAnalyticsData, generateAchievements]);
  
  const handleLogout = () => {
    logout();
    navigate('/');
  };
  
  const handleBackToDashboard = () => {
    navigate('/dashboard');
  };
  
  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top',
        labels: {
          font: {
            size: 12
          }
        }
      },
      tooltip: {
        padding: 10,
        titleFont: {
          size: 14
        },
        bodyFont: {
          size: 13
        }
      }
    }
  };
  
  // Bar chart specific options
  const barChartOptions = {
    ...chartOptions,
    scales: {
      x: {
        stacked: false,
        grid: {
          display: false
        }
      },
      y: {
        stacked: false,
        beginAtZero: true,
        grid: {
          color: 'rgba(0, 0, 0, 0.05)'
        }
      }
    }
  };
  
  return (
    <div className="insights-container">
      <div className="insights-header">
        <div className="insights-header-content">
          <h2>Insights & Achievements</h2>
          <div className="insights-actions">
            <button 
              className="back-button" 
              onClick={handleBackToDashboard}
              title="Back to Dashboard"
            >
              ← Dashboard
            </button>
            <button 
              className="logout-icon-button" 
              onClick={handleLogout} 
              title="Logout"
            >
              <img src="/assets/logout.png" alt="Logout" />
            </button>
          </div>
        </div>
      </div>
      
      <div className="insights-tabs">
        <button 
          className={`tab-button ${activeTab === 'analytics' ? 'active' : ''}`} 
          onClick={() => setActiveTab('analytics')}
        >
          Analytics
        </button>
        <button 
          className={`tab-button ${activeTab === 'achievements' ? 'active' : ''}`} 
          onClick={() => setActiveTab('achievements')}
        >
          Achievements
        </button>
      </div>
      
      <div className="insights-main">
        {isLoading ? (
          <div className="insights-loading">Loading insights data...</div>
        ) : (
          <>
            {activeTab === 'analytics' && (
              <div className="analytics-content">
                <div className="insights-row">
                  <div className="insights-card spending-trends-card hover-card">
                    <h3>Spending Trends</h3>
                    <div className="chart-container">
                      {spendingTrends.labels && spendingTrends.labels.length > 0 ? (
                        <Line data={spendingTrends} options={chartOptions} />
                      ) : (
                        <div className="no-data-message">
                          <p>Not enough data to show spending trends</p>
                          <p>Add more transactions to see your spending over time</p>
                        </div>
                      )}
                    </div>
                  </div>
                  
                  <div className="insights-card category-breakdown-card hover-card">
                    <h3>Expense Categories</h3>
                    <div className="chart-container donut-container">
                      {categoryBreakdown.labels && categoryBreakdown.labels.length > 0 ? (
                        <Pie 
                          data={categoryBreakdown} 
                          options={{
                            ...chartOptions,
                            cutout: '60%',
                            plugins: {
                              ...chartOptions.plugins,
                              legend: {
                                position: 'right',
                                align: 'center'
                              }
                            }
                          }} 
                        />
                      ) : (
                        <div className="no-data-message">
                          <p>No category data available</p>
                          <p>Categorize your expenses to see this breakdown</p>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
                
                <div className="insights-row">
                  <div className="insights-card savings-card hover-card">
                    <h3>Monthly Budget vs. Spending</h3>
                    <div className="chart-container">
                      {monthlySavings.labels && monthlySavings.labels.length > 0 ? (
                        <Bar data={monthlySavings} options={barChartOptions} />
                      ) : (
                        <div className="no-data-message">
                          <p>Not enough data to show savings analysis</p>
                          <p>Set budgets and add transactions to track your savings</p>
                        </div>
                      )}
                    </div>
                  </div>
                  
                  <div className="insights-card financial-health-card hover-card">
                    <h3>Financial Health Tips</h3>
                    <div className="tips-container">
                      <div className="tip-item">
                        <div className="tip-icon">💡</div>
                        <div className="tip-content">
                          <h4>Track Daily Expenses</h4>
                          <p>Recording small purchases helps identify spending patterns</p>
                        </div>
                      </div>
                      
                      <div className="tip-item">
                        <div className="tip-icon">📊</div>
                        <div className="tip-content">
                          <h4>Set Realistic Budgets</h4>
                          <p>Start with your actual spending patterns and adjust gradually</p>
                        </div>
                      </div>
                      
                      <div className="tip-item">
                        <div className="tip-icon">🧠</div>
                        <div className="tip-content">
                          <h4>Financial Planning</h4>
                          <p>Dedicate time each month to review and adjust your financial goals</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
            
            {activeTab === 'achievements' && (
              <div className="achievements-content">
                <div className="achievements-grid">
                  {achievements.map(achievement => (
                    <div 
                      key={achievement.id} 
                      className={`achievement-card hover-card ${achievement.unlocked ? 'unlocked' : 'locked'}`}
                    >
                      <div className="achievement-icon">{achievement.icon}</div>
                      <div className="achievement-info">
                        <h3 className="achievement-name">{achievement.name}</h3>
                        <p className="achievement-description">{achievement.description}</p>
                        
                        {achievement.progress !== undefined && (
                          <div className="achievement-progress">
                            <div className="progress-bar">
                              <div 
                                className="progress-fill" 
                                style={{ width: `${achievement.progress}%` }}
                              ></div>
                            </div>
                            <span className="progress-text">{Math.round(achievement.progress)}%</span>
                          </div>
                        )}
                        
                        {achievement.date && (
                          <div className="achievement-date">
                            Unlocked on {achievement.date}
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
                
                <div className="coming-soon-card hover-card">
                  <h3>More Achievements Coming Soon!</h3>
                  <p>We're working on adding more achievements to help track your financial journey.</p>
                  <div className="future-achievements">
                    <div className="future-achievement">
                      <span className="future-icon">🌟</span>
                      <span>Budget Master</span>
                    </div>
                    <div className="future-achievement">
                      <span className="future-icon">📈</span>
                      <span>Investment Starter</span>
                    </div>
                    <div className="future-achievement">
                      <span className="future-icon">🎯</span>
                      <span>Goal Crusher</span>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default Insights; 