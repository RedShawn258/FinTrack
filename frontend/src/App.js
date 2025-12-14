import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { Provider, useSelector, useDispatch } from 'react-redux';
import { store } from './store/store';
import { initializeAuth } from './store/slices/authSlice';
import Navbar from './components/Navbar';
import Login from './pages/Login';
import Signup from './pages/Signup';
import ForgotPassword from './pages/ForgotPassword';
import Dashboard from './pages/Dashboard';
import Landing from './pages/Landing';
import ExpenseTracking from './pages/ExpenseTracking';
import SmartBudgeting from './pages/SmartBudgeting';
import SavingsGoals from './pages/SavingsGoals';
import BudgetPlanningGuide from './pages/BudgetPlanning';
import Profile from './pages/Profile';
import './App.css';
import ScrollToTop from './ScrollToTop';
import ResetPassword from './pages/ResetPassword';
import Insights from './pages/Insights';
import './styles/darkMode.css';

const ProtectedRoute = ({ children }) => {
  const isAuthenticated = useSelector((state) => state.auth.isAuthenticated);
  const isLoading = useSelector((state) => state.auth.isLoading);

  if (isLoading) {
    return <div>Loading...</div>; // Or a proper loading component
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  return children;
};

const AppContent = () => {
  const location = useLocation();
  const dispatch = useDispatch();
  const hideNavbar = ['/', '/login', '/signup', '/forgot-password', '/reset-password', '/dashboard', '/profile', '/insights'].includes(location.pathname);

  // Initialize auth state on app load
  useEffect(() => {
    dispatch(initializeAuth());
  }, [dispatch]);

  useEffect(() => {
    // Initialize theme from localStorage or default to 'light'
    const savedTheme = localStorage.getItem('theme') || 'light';
    document.documentElement.setAttribute('data-theme', savedTheme);
  }, []);

  return (
    <>
      {!hideNavbar && <Navbar />}
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />

        {/* Protected Dashboard */}
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />
        
        {/* Protected Profile */}
        <Route
          path="/profile"
          element={
            <ProtectedRoute>
              <Profile />
            </ProtectedRoute>
          }
        />

        {/* Guides */}
        <Route path="/guides/expense-tracking" element={<ExpenseTracking />} />
        <Route path="/guides/smart-budgeting" element={<SmartBudgeting />} />
        <Route path="/guides/savings-goals" element={<SavingsGoals />} />

        {/* Budget Planning Guide */}
        <Route path="/budget-planning" element={<BudgetPlanningGuide />} />

        {/* Insights */}
        <Route path="/insights" element={
          <ProtectedRoute>
            <Insights />
          </ProtectedRoute>
        } />

        {/* Redirect any unknown routes to landing page */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </>
  );
};

const App = () => {
  return (
    <Provider store={store}>
      <Router>
        <ScrollToTop />
        <AppContent />
      </Router>
    </Provider>
  );
};

export default App;