import React, { useEffect, useState, useContext } from 'react';
import { AuthContext } from '../context/AuthContext';
import {
  getBudgets,
  getTransactions,
  createTransaction,
  getCategories,
  deleteTransaction,
  createCategory,
  createBudget
} from '../utils/api';
import { Pie } from 'react-chartjs-2';
import 'chart.js/auto';
import './Dashboard.css';
import { useNavigate } from 'react-router-dom';

const Dashboard = () => {
  const { user } = useContext(AuthContext);
  const token = user?.token;
  const navigate = useNavigate();

  const [budgets, setBudgets] = useState([]);
  const [categories, setCategories] = useState([]);
  const [transactions, setTransactions] = useState([]);

  const [newExpense, setNewExpense] = useState({
    description: '',
    amount: '',
    categoryName: '',
    transactionDate: ''
  });

  // New state for budget form
  const [newBudget, setNewBudget] = useState({
    categoryName: '',
    amount: '',
    startDate: '',
    endDate: ''
  });

  // For chart
  const [distributionData, setDistributionData] = useState({});

  const [confirmDialog, setConfirmDialog] = useState({
    isOpen: false,
    expenseId: null
  });

  const [categoryBudgetPrompt, setCategoryBudgetPrompt] = useState({
    isOpen: false,
    categoryName: ''
  });

  const [addExpenseConfirm, setAddExpenseConfirm] = useState({
    isOpen: false,
    expenseData: null
  });

  // ====== Fetch data on mount ======
  useEffect(() => {
    if (!token) return;
    fetchData();
  }, [token]);

  useEffect(() => {
    document.body.classList.add('dashboard-page');
    return () => {
      document.body.classList.remove('dashboard-page');
    };
  }, []);

  const fetchData = async () => {
    try {
      const [budRes, catRes, txRes] = await Promise.all([
        getBudgets(token),
        getCategories(token),
        getTransactions(token)
      ]);
      setBudgets(budRes.data.budgets);
      setCategories(catRes.data.categories);
      setTransactions(txRes.data.transactions);
    } catch (error) {
      console.error('Failed to fetch data', error);
      alert('Error fetching data');
    }
  };

  // ====== Recalc chart whenever transactions change ======
  useEffect(() => {
    calculateDistribution(transactions);
  }, [transactions]);

  const calculateDistribution = (txs) => {
    const categoryTotals = {};
    txs.forEach((tx) => {
      const catId = tx.CategoryID || 'Uncategorized';
      if (!categoryTotals[catId]) {
        categoryTotals[catId] = 0;
      }
      categoryTotals[catId] += tx.Amount;
    });

    const labels = [];
    const data = [];
    Object.keys(categoryTotals).forEach((catId) => {
      const cat = categories.find((c) => c.ID === Number(catId));
      const label = cat ? cat.Name : 'Uncategorized';
      labels.push(label);
      data.push(categoryTotals[catId]);
    });

    setDistributionData({
      labels,
      datasets: [
        {
          data,
          backgroundColor: [
            '#3498db',
            '#e67e22',
            '#2ecc71',
            '#9b59b6',
            '#f1c40f',
            '#e74c3c'
          ],
          hoverBackgroundColor: [
            '#2980b9',
            '#d35400',
            '#27ae60',
            '#8e44ad',
            '#f39c12',
            '#c0392b'
          ]
        }
      ]
    });
  };

  const formatDate = (dateStr) => {
    const options = { year: 'numeric', month: 'short', day: 'numeric' };
    return new Date(dateStr).toLocaleDateString('en-US', options);
  };

  // ====== Add New Expense ======
  const handleAddExpense = () => {
    const { description, amount, categoryName, transactionDate } = newExpense;
    if (!description || !amount || !transactionDate) {
      alert('Please fill all fields (description, amount, date)');
      return;
    }
    const amountNum = parseFloat(amount);
    if (isNaN(amountNum) || amountNum <= 0) {
      alert('Amount must be a positive number');
      return;
    }

    setAddExpenseConfirm({
      isOpen: true,
      expenseData: { ...newExpense, amount: amountNum }
    });
  };

  const confirmAddExpense = async (confirmed) => {
    if (confirmed && addExpenseConfirm.expenseData) {
      const { description, amount, categoryName, transactionDate } = addExpenseConfirm.expenseData;
      
      try {
        let chosenCategoryId = null;
        let isNewCategory = false;
        if (categoryName.trim() && categoryName.trim().toLowerCase() !== 'other') {
          const existingCat = categories.find(
            (cat) =>
              cat.Name.toLowerCase() === categoryName.trim().toLowerCase()
          );
          if (existingCat) {
            chosenCategoryId = existingCat.ID;
          } else {
            const res = await createCategory(token, { name: categoryName.trim() });
            const newCat = res.data.category;
            chosenCategoryId = newCat.ID;
            setCategories([...categories, newCat]);
            isNewCategory = true;
          }
        }

        await createTransaction(token, {
          description,
          amount,
          categoryId: chosenCategoryId,
          transactionDate
        });

        setNewExpense({
          description: '',
          amount: '',
          categoryName: '',
          transactionDate: ''
        });

        await fetchData();

        if (isNewCategory) {
          setCategoryBudgetPrompt({
            isOpen: true,
            categoryName: categoryName.trim()
          });
        } else {
          alert('Expense added successfully');
        }
      } catch (error) {
        console.error('Failed to create transaction', error);
        alert(error.response?.data?.error || 'Failed to create expense');
      }
    }
    setAddExpenseConfirm({ isOpen: false, expenseData: null });
  };

  const handleBudgetPromptResponse = (wantsToSetBudget) => {
    if (wantsToSetBudget) {
      setNewBudget({
        ...newBudget,
        categoryName: categoryBudgetPrompt.categoryName
      });
      document.querySelector('.set-budget-section')?.scrollIntoView({ behavior: 'smooth' });
    }
    setCategoryBudgetPrompt({ isOpen: false, categoryName: '' });
  };

  // ====== Delete Expense ======
  const handleDeleteExpense = async (txId) => {
    setConfirmDialog({
      isOpen: true,
      expenseId: txId
    });
  };

  const confirmDelete = async (confirmed) => {
    if (confirmed && confirmDialog.expenseId) {
      try {
        await deleteTransaction(token, confirmDialog.expenseId);
        fetchData();
      } catch (error) {
        console.error('Failed to delete transaction', error);
        alert(error.response?.data?.error || 'Failed to delete expense');
      }
    }
    setConfirmDialog({ isOpen: false, expenseId: null });
  };

  // ====== Summaries ======
  const totalExpenses = transactions.reduce((sum, tx) => sum + tx.Amount, 0);
  const sortedBudgets = [...budgets].sort((a, b) => {
    const catA = a.CategoryID
      ? categories.find((c) => c.ID === a.CategoryID)?.Name || 'Other'
      : 'Other';
    const catB = b.CategoryID
      ? categories.find((c) => c.ID === b.CategoryID)?.Name || 'Other'
      : 'Other';
    return catA.localeCompare(catB);
  });

  const recentTxs = [...transactions]
    .sort((a, b) => new Date(b.TransactionDate) - new Date(a.TransactionDate))
    .slice(0, 5);

  const handleAddBudget = async () => {
    const { categoryName, amount, startDate, endDate } = newBudget;
    if (!categoryName || !amount || !startDate || !endDate) {
      alert('Please fill all budget fields');
      return;
    }

    const amountNum = parseFloat(amount);
    if (isNaN(amountNum) || amountNum <= 0) {
      alert('Budget amount must be a positive number');
      return;
    }

    if (new Date(endDate) <= new Date(startDate)) {
      alert('End date must be after start date');
      return;
    }

    try {
      let categoryId = null;
      if (categoryName.trim()) {
        const existingCat = categories.find(
          (cat) => cat.Name.toLowerCase() === categoryName.trim().toLowerCase()
        );
        if (existingCat) {
          categoryId = existingCat.ID;
        } else {
          const res = await createCategory(token, { name: categoryName.trim() });
          categoryId = res.data.category.ID;
          setCategories([...categories, res.data.category]);
        }
      }

      await createBudget(token, {
        categoryId,
        limitAmount: amountNum,
        startDate,
        endDate
      });

      alert('Budget set successfully');
      fetchData();
      setNewBudget({
        categoryName: '',
        amount: '',
        startDate: '',
        endDate: ''
      });
    } catch (error) {
      console.error('Failed to set budget', error);
      alert(error.response?.data?.error || 'Failed to set budget');
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  return (
    <div className="dashboard-container">
      <div className="dashboard-header">
        <div className="dashboard-header-content">
          <h2>Total Expenses: ${totalExpenses.toFixed(2)}</h2>
          <button className="logout-icon-button" onClick={handleLogout} title="Logout">
            <img src="/assets/logout.png" alt="" />
          </button>
        </div>
      </div>

      <div className="dashboard-main">
        {/* LEFT COLUMN */}
        <div className="left-column">
          <div className="add-expense-card hover-card">
            <h3>Add New Expense</h3>
            <div className="expense-field">
              <label>Description</label>
              <input
                type="text"
                value={newExpense.description}
                onChange={(e) =>
                  setNewExpense({ ...newExpense, description: e.target.value })
                }
              />
            </div>
            <div className="expense-field">
              <label>Amount</label>
              {/* number input with no spinner arrows => see CSS */}
              <input
                type="number"
                value={newExpense.amount}
                onChange={(e) =>
                  setNewExpense({ ...newExpense, amount: e.target.value })
                }
              />
            </div>
            <div className="expense-field">
              <label>Category</label>
              {/* Using input + datalist so user can type or pick an existing category */}
              <input
                list="categoryList"
                value={newExpense.categoryName}
                onChange={(e) =>
                  setNewExpense({ ...newExpense, categoryName: e.target.value })
                }
              />
              <datalist id="categoryList">
                <option value="Other" />
                {categories.map((cat) => (
                  <option key={cat.ID} value={cat.Name} />
                ))}
              </datalist>
            </div>
            <div className="expense-field">
              <label>Date</label>
              <input
                type="date"
                value={newExpense.transactionDate}
                onChange={(e) =>
                  setNewExpense({ ...newExpense, transactionDate: e.target.value })
                }
              />
            </div>
            <button onClick={handleAddExpense}>Add Expense</button>
          </div>

          <div className="budget-overview-card hover-card">
            <h3>Budget Overview</h3>
            
            {/* Add New Budget Form */}
            <div className="set-budget-section">
              <h4>Set New Budget</h4>
              <div className="budget-field">
                <label>Category</label>
                <input
                  list="categoryList"
                  value={newBudget.categoryName}
                  onChange={(e) =>
                    setNewBudget({ ...newBudget, categoryName: e.target.value })
                  }
                  placeholder="Select or type category"
                />
              </div>
              <div className="budget-field">
                <label>Budget Amount</label>
                <input
                  type="number"
                  value={newBudget.amount}
                  onChange={(e) =>
                    setNewBudget({ ...newBudget, amount: e.target.value })
                  }
                  placeholder="Enter amount"
                />
              </div>
              <div className="budget-dates">
                <div className="budget-field">
                  <label>Start Date</label>
                  <input
                    type="date"
                    value={newBudget.startDate}
                    onChange={(e) =>
                      setNewBudget({ ...newBudget, startDate: e.target.value })
                    }
                  />
                </div>
                <div className="budget-field">
                  <label>End Date</label>
                  <input
                    type="date"
                    value={newBudget.endDate}
                    onChange={(e) =>
                      setNewBudget({ ...newBudget, endDate: e.target.value })
                    }
                  />
                </div>
              </div>
              <button onClick={handleAddBudget} className="set-budget-button">
                Set Budget
              </button>
            </div>

            <div className="budget-list">
              <h4>Current Budgets</h4>
              {sortedBudgets.length === 0 ? (
                <p className="no-budgets">No budgets found</p>
              ) : (
                sortedBudgets.map((b) => {
                  const catName = b.CategoryID
                    ? categories.find((c) => c.ID === b.CategoryID)?.Name || 'Other'
                    : 'Other';
                  const limit = b.LimitAmount ?? 0;
                  const remaining = b.RemainingAmount ?? 0;
                  const used = limit - remaining;
                  const progress = limit > 0 ? (used / limit) * 100 : 0;
                  const dateRange = b.StartDate && b.EndDate 
                    ? `${formatDate(b.StartDate)} - ${formatDate(b.EndDate)}`
                    : 'No date range';
                  
                  return (
                    <div key={b.ID} className="budget-row">
                      <div className="budget-header">
                        <span className="category-name">{catName}</span>
                        <span className="date-range">{dateRange}</span>
                      </div>
                      <div className="budget-label">
                        <span>Used: ${used.toFixed(2)}</span>
                        <span>Budget: ${limit.toFixed(2)}</span>
                      </div>
                      <div className="budget-progress-bar">
                        <div
                          className="budget-progress-fill"
                          style={{ width: `${Math.min(progress, 100)}%` }}
                        />
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>
        </div>

        {/* RIGHT COLUMN */}
        <div className="right-column">
          <div className="expense-distribution-card hover-card">
            <h3>Expense Distribution</h3>
            <div className="pie-chart-container">
              {distributionData.labels && distributionData.labels.length > 0 ? (
                <Pie 
                  data={distributionData}
                  options={{
                    responsive: true,
                    maintainAspectRatio: true,
                    plugins: {
                      legend: {
                        position: 'top',
                      }
                    }
                  }}
                />
              ) : (
                <p>No transactions yet</p>
              )}
            </div>
          </div>

          {/* #2: Wider container for the table => use a special class or inline style */}
          <div className="recent-expenses-card hover-card wider-card">
            <h3>Recent Expenses</h3>
            <div className="recent-expenses-table-container">
              {recentTxs.length === 0 ? (
                <p>No recent expenses</p>
              ) : (
                <table>
                  <thead>
                    <tr>
                      <th>Date</th>
                      <th>Description</th>
                      <th>Category</th>
                      <th>Amount</th>
                      {/* No "Del" header—just an empty cell, or remove altogether */}
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {recentTxs.map((tx) => {
                      const cat = tx.CategoryID
                        ? categories.find((c) => c.ID === tx.CategoryID)?.Name
                        : 'Other';
                      return (
                        <tr key={tx.ID}>
                          <td>{formatDate(tx.TransactionDate)}</td>
                          <td>{tx.Description}</td>
                          <td>{cat || 'Other'}</td>
                          <td>${tx.Amount.toFixed(2)}</td>
                          <td>
                            <button
                              className="delete-button"
                              onClick={() => handleDeleteExpense(tx.ID)}
                              title="Delete expense"
                            >
                              &#128465;
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Add this JSX right before the closing </div> of dashboard-container */}
      {confirmDialog.isOpen && (
        <div className="modal-overlay">
          <div className="confirmation-dialog">
            <p>Are you sure you want to delete this expense?</p>
            <div className="confirmation-dialog-buttons">
              <button 
                className="confirmation-dialog-button confirmation-yes" 
                onClick={() => confirmDelete(true)}
              >
                Yes
              </button>
              <button 
                className="confirmation-dialog-button confirmation-no" 
                onClick={() => confirmDelete(false)}
              >
                No
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add this JSX right before the closing </div> of dashboard-container, after the confirmation dialog */}
      {categoryBudgetPrompt.isOpen && (
        <div className="modal-overlay">
          <div className="confirmation-dialog">
            <p>New Category "{categoryBudgetPrompt.categoryName}" found. Would you like to set a budget for it?</p>
            <div className="confirmation-dialog-buttons">
              <button 
                className="confirmation-dialog-button confirmation-yes" 
                onClick={() => handleBudgetPromptResponse(true)}
              >
                Yes
              </button>
              <button 
                className="confirmation-dialog-button confirmation-no" 
                onClick={() => handleBudgetPromptResponse(false)}
              >
                No
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add this JSX before the closing </div> of dashboard-container, after other dialogs */}
      {addExpenseConfirm.isOpen && (
        <div className="modal-overlay">
          <div className="confirmation-dialog">
            <p>Are you sure you want to add this expense?</p>
            <div className="expense-summary">
              <div>Description: {addExpenseConfirm.expenseData?.description}</div>
              <div>Amount: ${addExpenseConfirm.expenseData?.amount.toFixed(2)}</div>
              <div>Category: {addExpenseConfirm.expenseData?.categoryName || 'Other'}</div>
              <div>Date: {addExpenseConfirm.expenseData?.transactionDate}</div>
            </div>
            <div className="confirmation-dialog-buttons">
              <button 
                className="confirmation-dialog-button confirmation-yes" 
                onClick={() => confirmAddExpense(true)}
              >
                Yes
              </button>
              <button 
                className="confirmation-dialog-button confirmation-no" 
                onClick={() => confirmAddExpense(false)}
              >
                No
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Dashboard;
