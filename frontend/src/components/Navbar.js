// src/components/Navbar.js
import React, { useContext } from 'react';
import { Link } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

const Navbar = () => {
  const { user, logout } = useContext(AuthContext);

  return (
    <nav className="navbar">
      <div>
        <Link to="/" className="navbar-brand">
          FinTrack
        </Link>
      </div>
      <div>
        {user ? (
          <>
            <Link to="/dashboard">Dashboard</Link>
            {/* we can keep or remove these extra links if you don't have those pages */}
            {/* <Link to="/budgets">Budgets</Link>
            <Link to="/categories">Categories</Link>
            <Link to="/transactions">Transactions</Link> */}
            <button onClick={logout}>Logout</button>
          </>
        ) : (
          <>
            <Link to="/login">Login</Link>
            <Link to="/signup">Signup</Link>
          </>
        )}
      </div>
    </nav>
  );
};

export default Navbar;
