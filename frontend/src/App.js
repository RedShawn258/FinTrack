import React, { useContext } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { AuthProvider, AuthContext } from './context/AuthContext';
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
import './App.css';

const ProtectedRoute = ({ children }) => {
  const { user } = useContext(AuthContext);
  if (!user) {
    return <Navigate to="/login" />;
  }
  return children;
};

const AppContent = () => {
  const location = useLocation();
  const hideNavbar = ['/login', '/signup', '/forgot-password'].includes(location.pathname);

  return (
    <>
      {!hideNavbar && <Navbar />}
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />

        {/* Protected Dashboard */}
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />

        {/* Guides */}
        <Route path="/guides/expense-tracking" element={<ExpenseTracking />} />
        <Route path="/guides/smart-budgeting" element={<SmartBudgeting />} />
        <Route path="/guides/savings-goals" element={<SavingsGoals />} />

        {/* Budget Planning Guide */}
        <Route path="/budget-planning" element={<BudgetPlanningGuide />} />

        {/* Redirect any unknown routes to landing page */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </>
  );
};

const App = () => {
  return (
    <AuthProvider>
      <Router>
        <AppContent />
      </Router>
    </AuthProvider>
  );
};

export default App;