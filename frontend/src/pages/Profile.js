import React, { useState, useEffect, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import { fetchProfile, updateProfile } from '../utils/api';
import useTheme from '../hooks/useTheme';
import './Profile.css';
import { InsightsIcon } from '../icons';

const Profile = () => {
  const { user, logout } = useContext(AuthContext);
  const navigate = useNavigate();
  const token = user?.token;
  const { theme, setTheme } = useTheme();
  
  const [profile, setProfile] = useState({
    username: '',
    email: '',
    firstName: '',
    lastName: '',
    phoneNumber: '',
    currency: 'USD',
    notificationsEnabled: true,
    theme: theme // Initialize with current theme
  });
  
  const [isLoading, setIsLoading] = useState(true);
  const [message, setMessage] = useState('');
  
  useEffect(() => {
    // Add auth-page class and theme attribute to body
    document.body.classList.add('auth-page');
    document.body.setAttribute('data-theme', theme);
    
    return () => {
      // Cleanup
      document.body.classList.remove('auth-page');
      document.body.removeAttribute('data-theme');
    };
  }, [theme]);
  
  useEffect(() => {
    const getProfile = async () => {
      if (!token) return;
      
      try {
        setIsLoading(true);
        const response = await fetchProfile(token);
        setProfile(p => ({
          ...p,
          ...response.data,
        }));
      } catch (error) {
        console.error('Failed to fetch profile', error);
        setMessage('Failed to load profile data. Please try again.');
      } finally {
        setIsLoading(false);
      }
    };
    
    getProfile();
  }, [token]);
  
  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    const newValue = type === 'checkbox' ? checked : value;
    
    setProfile(prev => ({
      ...prev,
      [name]: newValue
    }));

    // If theme is changed, update it immediately
    if (name === 'theme') {
      setTheme(value);
    }
  };
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    try {
      setIsLoading(true);
      await updateProfile(token, {
        firstName: profile.firstName,
        lastName: profile.lastName,
        phoneNumber: profile.phoneNumber,
        currency: profile.currency,
        notificationsEnabled: profile.notificationsEnabled,
        theme: profile.theme
      });
      
      setMessage('Profile updated successfully!');
      
      // If theme was changed, apply it immediately
      document.documentElement.setAttribute('data-theme', profile.theme);
      localStorage.setItem('theme', profile.theme);
      
    } catch (error) {
      console.error('Failed to update profile', error);
      setMessage('Failed to update profile. Please try again.');
    } finally {
      setIsLoading(false);
      
      // Clear message after 3 seconds
      setTimeout(() => {
        setMessage('');
      }, 3000);
    }
  };
  
  const handleLogout = () => {
    logout();
    navigate('/');
  };
  
  const handleBackToDashboard = () => {
    navigate('/dashboard');
  };
  
  const currencyOptions = [
    { value: 'USD', label: 'US Dollar ($)' },
    { value: 'EUR', label: 'Euro (€)' },
    { value: 'GBP', label: 'British Pound (£)' },
    { value: 'JPY', label: 'Japanese Yen (¥)' },
    { value: 'INR', label: 'Indian Rupee (₹)' },
    { value: 'CAD', label: 'Canadian Dollar (C$)' },
    { value: 'AUD', label: 'Australian Dollar (A$)' }
  ];
  
  return (
    <div className="profile-container">
      <div className="profile-header">
        <div className="profile-header-content">
          <h2>User Profile</h2>
          <div className="profile-actions">
            <button 
              className="back-button" 
              onClick={handleBackToDashboard}
              title="Back to Dashboard"
            >
              ← Dashboard
            </button>
            <button 
              className="insights-icon-button" 
              onClick={() => navigate('/insights')} 
              title="Insights"
            >
              <InsightsIcon />
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
      
      <div className="profile-main">
        <div className="profile-card hover-card">
          {isLoading ? (
            <div className="profile-loading">Loading profile...</div>
          ) : (
            <form onSubmit={handleSubmit}>
              <div className="profile-section">
                <h3>Account Information</h3>
                <div className="profile-field">
                  <label>Username</label>
                  <input 
                    type="text"
                    name="username"
                    value={profile.username}
                    disabled
                    className="disabled-input"
                  />
                </div>
                <div className="profile-field">
                  <label>Email</label>
                  <input 
                    type="email"
                    name="email"
                    value={profile.email}
                    disabled
                    className="disabled-input"
                  />
                </div>
              </div>
              
              <div className="profile-section">
                <h3>Personal Information</h3>
                <div className="profile-field">
                  <label>First Name</label>
                  <input 
                    type="text"
                    name="firstName"
                    value={profile.firstName || ''}
                    onChange={handleChange}
                    placeholder="Enter your first name"
                  />
                </div>
                <div className="profile-field">
                  <label>Last Name</label>
                  <input 
                    type="text"
                    name="lastName"
                    value={profile.lastName || ''}
                    onChange={handleChange}
                    placeholder="Enter your last name"
                  />
                </div>
                <div className="profile-field">
                  <label>Phone Number</label>
                  <input 
                    type="text"
                    name="phoneNumber"
                    value={profile.phoneNumber || ''}
                    onChange={handleChange}
                    placeholder="Enter your phone number"
                  />
                </div>
              </div>
              
              <div className="profile-section">
                <h3>Preferences</h3>
                <div className="profile-field">
                  <label>Currency</label>
                  <select 
                    name="currency"
                    value={profile.currency}
                    onChange={handleChange}
                  >
                    {currencyOptions.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="profile-field theme-selection">
                  <label>Theme</label>
                  <div className="theme-options">
                    <label className={`theme-option ${profile.theme === 'light' ? 'selected' : ''}`}>
                      <input
                        type="radio"
                        name="theme"
                        value="light"
                        checked={profile.theme === 'light'}
                        onChange={handleChange}
                      />
                      <span className="theme-label">Light</span>
                      <div className="theme-preview light-preview"></div>
                    </label>
                    <label className={`theme-option ${profile.theme === 'dark' ? 'selected' : ''}`}>
                      <input
                        type="radio"
                        name="theme"
                        value="dark"
                        checked={profile.theme === 'dark'}
                        onChange={handleChange}
                      />
                      <span className="theme-label">Dark</span>
                      <div className="theme-preview dark-preview"></div>
                    </label>
                  </div>
                </div>
                <div className="profile-field checkbox-field">
                  <label className="checkbox-container">
                    <input
                      type="checkbox"
                      name="notificationsEnabled"
                      checked={profile.notificationsEnabled}
                      onChange={handleChange}
                    />
                    <span className="custom-checkbox"></span>
                    Enable Notifications
                  </label>
                </div>
              </div>
              
              {message && <div className="profile-message">{message}</div>}
              
              <div className="profile-actions-bottom">
                <button type="submit" className="save-button">
                  Save Changes
                </button>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  );
};

export default Profile; 