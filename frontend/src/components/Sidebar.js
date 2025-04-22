import React, { useEffect, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { HomeIcon, UserIcon, InsightsIcon } from '../icons';

const Sidebar = () => {
  const location = useLocation();

  return (
    <div className="sidebar">
      <ul>
        <li className={`${location.pathname === '/dashboard' ? 'active' : ''}`}>
          <Link to="/dashboard">
            <HomeIcon />
            <span>Dashboard</span>
          </Link>
        </li>
        <li className={`${location.pathname === '/profile' ? 'active' : ''}`}>
          <Link to="/profile">
            <UserIcon />
            <span>Profile</span>
          </Link>
        </li>
        <li className={`${location.pathname === '/insights' ? 'active' : ''}`}>
          <Link to="/insights">
            <InsightsIcon />
            <span>Insights</span>
          </Link>
        </li>
      </ul>
    </div>
  );
};

export default Sidebar; 